package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"ecommerce-system/internal/pkg/cache"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/pkg/outbox"
	"ecommerce-system/internal/service/product/model"
	"ecommerce-system/internal/service/product/repository"

	"gorm.io/gorm"
)

func isDuplicateSkuCodeErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	// MySQL: Error 1062 (23000): Duplicate entry 'xxx' for key 'sku.uk_sku_code'
	return strings.Contains(s, "Duplicate entry") && strings.Contains(s, "uk_sku_code")
}

// ProductLogic 商品业务逻辑
type ProductLogic struct {
	db           *gorm.DB
	outboxRepo   *outbox.Repo
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	skuRepo      repository.SkuRepository
	bannerRepo   repository.BannerRepository
	cache        *cache.CacheOperations
	mqProducer   *mq.Producer
}

// NewProductLogic 创建商品业务逻辑
func NewProductLogic(
	db *gorm.DB,
	outboxRepo *outbox.Repo,
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	skuRepo repository.SkuRepository,
	bannerRepo repository.BannerRepository,
	cache *cache.CacheOperations,
	mqProducer *mq.Producer,
) *ProductLogic {
	return &ProductLogic{
		db:           db,
		outboxRepo:   outboxRepo,
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		skuRepo:      skuRepo,
		bannerRepo:   bannerRepo,
		cache:        cache,
		mqProducer:   mqProducer,
	}
}

type GetProductRequest struct {
	ID      uint64
	SpuCode string
}

type GetProductResponse struct {
	Product *model.Product
	Skus    []*model.Sku
}

func (l *ProductLogic) GetProduct(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	// 这里是product服务的查询，所以检查一下这里的链接的情况

	if l.productRepo == nil {
		return nil, apperrors.NewInternalError("数据库没有初始化")
	}

	var productID uint64
	var err error

	//这里是一个比较傻逼的逻辑，先想办法获取productID，如果没有，就根据SpuCode查询一次，获取到productID后再查询一次，这样就能复用之前的查询逻辑了
	if req.ID != 0 {
		productID = req.ID
	} else if req.SpuCode != "" {
		product, err := l.productRepo.GetBySpuCode(ctx, req.SpuCode)
		if err != nil {
			return nil, apperrors.NewInternalError("根据SpuCode查询商品失败: " + err.Error())
		}
		if product == nil {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "商品不存在")
		}
		productID = product.ID
	} else {
		return nil, apperrors.NewInvalidParamError("商品ID或SPU编码不能为空")
	}

	// 瞎几把走一下缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, productID)
		var cachedResp GetProductResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// 从数据库查询
	product, err := l.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
		}
		return nil, apperrors.NewInternalError("查询商品失败: " + err.Error())
	}
	if product == nil {
		return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
	}

	skus, err := l.skuRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询商品SKU失败: " + err.Error())
	}

	if len(skus) > 0 {
		minPrice := math.MaxFloat64
		totalStock := 0
		hasActive := false
		for _, s := range skus {
			if s == nil || s.Price != 1 {
				continue
			}
			hasActive = true
			if s.Price > 0 && s.Price < minPrice {
				minPrice = s.Price
			}
			if s.Stock > 0 {
				totalStock += s.Stock
			}
		}
		if hasActive {
			if minPrice != math.MaxFloat64 {
				product.Price = minPrice
			}
			product.Stock = totalStock
		}
	}

	resp := &GetProductResponse{
		Product: product,
		Skus:    skus,
	}

	// 写缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, productID)
		_ = l.cache.Set(ctx, cacheKey, resp, 1*time.Hour)
	}

	return resp, nil
}

// ListProductsRequest 获取商品列表请求
type ListProductsRequest struct {
	CategoryID uint64
	BrandID    uint64
	Keyword    string
	Status     int8
	IsHot      int8 // -1-全部, 0-否, 1-是
	Page       int
	PageSize   int
	Sort       string
}

// ListProductsResponse 获取商品列表响应
type ListProductsResponse struct {
	Products   []*model.Product
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

func (l *ProductLogic) ListProducts(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {

	if l.productRepo == nil {
		return nil, apperrors.NewInternalError("数据库没有初始化")
	}

	//参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.Sort == "" {
		req.Sort = "default"
	}
	// 兼容：点击“主分类”时，需要过滤出其所有子分类（以及更深层级）下的商品
	// 前端仍然只传 category_id=主分类ID，后端自动展开为 category_id IN (...)
	var expandedCategoryIDs []uint64
	if req.CategoryID > 0 && l.categoryRepo != nil {
		ids, err := l.collectDescendantCategoryIDs(ctx, req.CategoryID)
		if err == nil && len(ids) > 1 {
			expandedCategoryIDs = ids
		}
	}

	// 缓存的key
	if l.cache != nil {
		cacheKey := cache.BuildKey(
			cache.KeyPrefixProductList,
			req.CategoryID, req.BrandID, req.Keyword, req.Status, req.IsHot, req.Page, req.PageSize, req.Sort,
		)
		var cachedResp ListProductsResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	repoReq := &repository.ListProductsRequest{
		CategoryID:  req.CategoryID,
		CategoryIDs: expandedCategoryIDs,
		BrandID:     req.BrandID,
		Keyword:     req.Keyword,
		Status:      req.Status,
		IsHot:       req.IsHot,
		Page:        req.Page,
		PageSize:    req.PageSize,
		Sort:        req.Sort,
	}
	products, total, err := l.productRepo.List(ctx, repoReq)
	if err != nil {
		return nil, apperrors.NewInternalError("查询商品列表失败: " + err.Error())
	}

	// 口径统一：列表页返回的 product.price / product.stock 用“上架 SKU 聚合值”兜底（避免前端只拿到 SPU 原值）
	// 使用批量聚合，避免 N+1
	if l.skuRepo != nil && len(products) > 0 {
		ids := make([]uint64, 0, len(products))
		for _, p := range products {
			if p != nil && p.ID > 0 {
				ids = append(ids, p.ID)
			}
		}
		if len(ids) > 0 {
			agg, err := l.skuRepo.GetAggByProductIDs(ctx, ids, 1)
			if err == nil && len(agg) > 0 {
				for _, p := range products {
					if p == nil {
						continue
					}
					if a, ok := agg[p.ID]; ok {
						if a.MinPrice > 0 {
							p.Price = a.MinPrice
						}
						// 总库存可能很大，这里做个保底转换
						if a.TotalStock >= 0 {
							if a.TotalStock > int64(^uint(0)>>1) {
								p.Stock = int(^uint(0) >> 1)
							} else {
								p.Stock = int(a.TotalStock)
							}
						}
						// 如果你未来要在列表展示“原价”，可以用 a.MinOriginalPrice 去兜底 p.OriginalPrice
					}
				}
			}
		}
	}

	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	resp := &ListProductsResponse{
		Products:   products,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	// 写缓存 10分钟
	if l.cache != nil {
		cacheKey := cache.BuildKey(
			cache.KeyPrefixProductList,
			req.CategoryID, req.BrandID, req.Keyword, req.Status, req.IsHot, req.Page, req.PageSize, req.Sort,
		)
		_ = l.cache.Set(ctx, cacheKey, resp, 10*time.Minute)
	}

	return resp, nil

}

// collectDescendantCategoryIDs 收集某个分类的所有子孙分类 ID（包含自身）
// - 仅依赖 parent_id 关系，支持任意层级
// - 用 visited 防止异常数据导致死循环
func (l *ProductLogic) collectDescendantCategoryIDs(ctx context.Context, rootID uint64) ([]uint64, error) {
	if rootID == 0 || l.categoryRepo == nil {
		return nil, nil
	}
	visited := map[uint64]bool{}
	queue := []uint64{rootID}
	out := make([]uint64, 0, 16)

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if id == 0 || visited[id] {
			continue
		}
		visited[id] = true
		out = append(out, id)

		children, err := l.categoryRepo.GetByParentID(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, c := range children {
			if c != nil && c.ID > 0 && !visited[c.ID] {
				queue = append(queue, c.ID)
			}
		}
	}
	return out, nil
}
