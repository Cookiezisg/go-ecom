package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type GetSkuRequest struct {
	ID      uint64
	SkuCode string
}

type GetSkuResponse struct {
	Sku *model.Sku
}

func (l *ProductLogic) GetSku(ctx context.Context, req *GetSkuRequest) (*GetSkuResponse, error) {

	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库没有初始化")
	}

	var skuID uint64
	var err error

	if req.ID > 0 {
		skuID = req.ID
	} else if req.SkuCode != "" {
		sku, err := l.skuRepo.GetBySkuCode(ctx, req.SkuCode)
		if err != nil {
			return nil, apperrors.NewInternalError("根据SKU编码查询SKU失败: " + err.Error())
		}
		if sku == nil {
			return nil, apperrors.NewError(apperrors.CodeSkuNotFound, "SKU不存在")
		}
		skuID = sku.ID
	} else {
		return nil, apperrors.NewInvalidParamError("SKU ID或SKU编码不能为空")
	}

	// 走一下缓存试试看
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, skuID)
		var cachedResp GetSkuResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	sku, err := l.skuRepo.GetByID(ctx, skuID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if sku == nil {
		return nil, apperrors.NewError(apperrors.CodeSkuNotFound, "SKU不存在")
	}

	resp := &GetSkuResponse{
		Sku: sku,
	}

	// 写缓存 1小时
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, skuID)
		_ = l.cache.Set(ctx, cacheKey, resp, 1*time.Hour)
	}

	return resp, nil
}

type ListSkusRequest struct {
	ProductID uint64
	Status    int8
	Page      int
	PageSize  int
}

type ListSkusResponse struct {
	Skus       []*model.Sku
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

func (l *ProductLogic) ListSkus(ctx context.Context, req *ListSkusRequest) (*ListSkusResponse, error) {

	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库没有初始化")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	listReq := &repository.ListSkusRequest{
		ProductID: req.ProductID,
		Status:    req.Status,
		Page:      page,
		PageSize:  pageSize,
	}

	skus, total, err := l.skuRepo.List(ctx, listReq)
	if err != nil {
		return nil, apperrors.NewInternalError("查询SKU列表失败: " + err.Error())
	}

	if req.ProductID > 0 && req.Status == 1 && total == 0 {
		existing, err := l.skuRepo.GetByProductID(ctx, req.ProductID)
		if err == nil && len(existing) == 0 && l.productRepo != nil {
			product, err := l.productRepo.GetByID(ctx, req.ProductID)
			if err == nil && product != nil {
				defaultSku := &model.Sku{
					ProductID: req.ProductID,
					SkuCode:   fmt.Sprintf("SKU%d-DEFAULT", req.ProductID),
					Name:      product.Name,
					Specs:     "{}",
					Price:     product.Price,
					Stock:     product.Stock,
					Status:    1,
				}
				if product.OriginalPrice != nil {
					defaultSku.OriginalPrice = product.OriginalPrice
				}
				// 优先用本地主图
				if product.LocalMainImage != "" {
					defaultSku.Image = product.LocalMainImage
				} else {
					defaultSku.Image = product.MainImage
				}

				_ = l.skuRepo.Create(ctx, defaultSku) // 忽略重复创建错误（并发情况下）

				// 清理商品详情缓存，避免 SKU 列表和详情不一致
				if l.cache != nil {
					_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, req.ProductID))
				}

				// 重新查询一次（确保拿到 ID / 最新数据）
				skus, total, _ = l.skuRepo.List(ctx, listReq)
			}
		}
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	return &ListSkusResponse{
		Skus:       skus,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// CreateSkuRequest 创建SKU请求
type CreateSkuRequest struct {
	ProductID     uint64
	SkuCode       string
	Name          string
	Specs         map[string]string // 规格属性
	Price         float64
	OriginalPrice *float64
	Stock         int
	Image         string
	Weight        *float64
	Volume        *float64
	Status        int8
}

// CreateSkuResponse 创建SKU响应
type CreateSkuResponse struct {
	Sku *model.Sku
}

// CreateSku 创建SKU
func (l *ProductLogic) CreateSku(ctx context.Context, req *CreateSkuRequest) (*CreateSkuResponse, error) {
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：SKU 变更需要触发商品索引更新
	if l.db != nil && l.outboxRepo != nil {
		var outSku *model.Sku
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			productRepoTx := repository.NewProductRepository(tx)
			skuRepoTx := repository.NewSkuRepository(tx)

			// 验证必填字段
			if req.ProductID == 0 {
				return apperrors.NewInvalidParamError("商品ID不能为空")
			}
			if req.SkuCode == "" {
				return apperrors.NewInvalidParamError("SKU编码不能为空")
			}
			if req.Name == "" {
				return apperrors.NewInvalidParamError("SKU名称不能为空")
			}
			if len(req.Specs) == 0 {
				return apperrors.NewInvalidParamError("规格属性不能为空")
			}
			if req.Price <= 0 {
				return apperrors.NewInvalidParamError("价格必须大于0")
			}

			// 检查商品是否存在
			product, err := productRepoTx.GetByID(ctx, req.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
				}
				return apperrors.NewInternalError("查询商品失败: " + err.Error())
			}
			if product == nil {
				return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
			}

			// sku_code 唯一：如果存在软删除记录，则“恢复并覆盖字段”
			existingAny, err := skuRepoTx.GetBySkuCodeUnscoped(ctx, req.SkuCode)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if existingAny != nil {
				// 未删除：直接提示已存在
				if !existingAny.DeletedAt.Valid {
					return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
				}
				// 已删除：恢复并更新（等价于重新创建）
				specsJSON, err := json.Marshal(req.Specs)
				if err != nil {
					return apperrors.NewInternalError("规格属性格式错误: " + err.Error())
				}
				now := time.Now()
				updates := map[string]any{
					"product_id":     req.ProductID,
					"sku_code":       req.SkuCode,
					"name":           req.Name,
					"specs":          string(specsJSON),
					"price":          req.Price,
					"original_price": req.OriginalPrice,
					"stock":          req.Stock,
					"image":          req.Image,
					"weight":         req.Weight,
					"volume":         req.Volume,
					"status":         req.Status,
					"updated_at":     now,
				}
				if err := skuRepoTx.RestoreAndUpdateByID(ctx, existingAny.ID, updates); err != nil {
					if isDuplicateSkuCodeErr(err) {
						return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
					}
					return apperrors.NewInternalError("恢复SKU失败: " + err.Error())
				}
				restored, _ := skuRepoTx.GetByID(ctx, existingAny.ID)
				if restored == nil {
					restored = existingAny
				}
				outSku = restored

				// 同事务写 outbox
				payloadBytes, _ := json.Marshal(map[string]any{"product_id": req.ProductID})
				payload := string(payloadBytes)
				evt := &outbox.Event{
					AggregateType: "product",
					AggregateID:   fmt.Sprintf("%d", req.ProductID),
					EventType:     outbox.EventProductUpserted,
					Payload:       &payload,
					Status:        outbox.StatusPending,
				}
				return l.outboxRepo.CreateInTx(ctx, tx, evt)
			}

			// 将规格属性转换为JSON
			specsJSON, err := json.Marshal(req.Specs)
			if err != nil {
				return apperrors.NewInternalError("规格属性格式错误: " + err.Error())
			}

			now := time.Now()
			sku := &model.Sku{
				ProductID:     req.ProductID,
				SkuCode:       req.SkuCode,
				Name:          req.Name,
				Specs:         string(specsJSON),
				Price:         req.Price,
				OriginalPrice: req.OriginalPrice,
				Stock:         req.Stock,
				Image:         req.Image,
				Weight:        req.Weight,
				Volume:        req.Volume,
				Status:        req.Status,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := skuRepoTx.Create(ctx, sku); err != nil {
				if isDuplicateSkuCodeErr(err) {
					return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
				}
				return apperrors.NewInternalError("创建SKU失败: " + err.Error())
			}
			outSku = sku

			// 同事务写 outbox
			payloadBytes, _ := json.Marshal(map[string]any{"product_id": sku.ProductID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", sku.ProductID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			// 业务错误直接透传
			return nil, err
		}

		// 清除缓存（事务提交后）
		if outSku != nil && l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, outSku.ID))
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, outSku.ProductID))
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}

		return &CreateSkuResponse{Sku: outSku}, nil
	}

	// 验证必填字段
	if req.ProductID == 0 {
		return nil, apperrors.NewInvalidParamError("商品ID不能为空")
	}
	if req.SkuCode == "" {
		return nil, apperrors.NewInvalidParamError("SKU编码不能为空")
	}
	if req.Name == "" {
		return nil, apperrors.NewInvalidParamError("SKU名称不能为空")
	}
	if len(req.Specs) == 0 {
		return nil, apperrors.NewInvalidParamError("规格属性不能为空")
	}
	if req.Price <= 0 {
		return nil, apperrors.NewInvalidParamError("价格必须大于0")
	}

	// 检查商品是否存在
	product, err := l.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
		}
		return nil, apperrors.NewInternalError("查询商品失败: " + err.Error())
	}
	if product == nil {
		return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
	}

	// sku_code 唯一：如果存在软删除记录，则“恢复并覆盖字段”（用户期望：删了能用同一个 sku_code 重新添加）
	// 如果存在未删除记录，则报已存在。
	existingAny, err := l.skuRepo.GetBySkuCodeUnscoped(ctx, req.SkuCode)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if existingAny != nil {
		// 未删除：直接提示已存在
		if !existingAny.DeletedAt.Valid {
			return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
		}
		// 已删除：恢复并更新（等价于重新创建）
		specsJSON, err := json.Marshal(req.Specs)
		if err != nil {
			return nil, apperrors.NewInternalError("规格属性格式错误: " + err.Error())
		}
		now := time.Now()
		updates := map[string]any{
			"product_id":     req.ProductID,
			"sku_code":       req.SkuCode,
			"name":           req.Name,
			"specs":          string(specsJSON),
			"price":          req.Price,
			"original_price": req.OriginalPrice,
			"stock":          req.Stock,
			"image":          req.Image,
			"weight":         req.Weight,
			"volume":         req.Volume,
			"status":         req.Status,
			"updated_at":     now,
		}
		if err := l.skuRepo.RestoreAndUpdateByID(ctx, existingAny.ID, updates); err != nil {
			if isDuplicateSkuCodeErr(err) {
				return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
			}
			return nil, apperrors.NewInternalError("恢复SKU失败: " + err.Error())
		}
		restored, _ := l.skuRepo.GetByID(ctx, existingAny.ID)
		if restored == nil {
			restored = existingAny
		}
		// 清除缓存
		if l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, existingAny.ID))
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, req.ProductID))
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}
		return &CreateSkuResponse{Sku: restored}, nil
	}

	// 将规格属性转换为JSON
	specsJSON, err := json.Marshal(req.Specs)
	if err != nil {
		return nil, apperrors.NewInternalError("规格属性格式错误: " + err.Error())
	}

	now := time.Now()
	sku := &model.Sku{
		ProductID:     req.ProductID,
		SkuCode:       req.SkuCode,
		Name:          req.Name,
		Specs:         string(specsJSON),
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Stock:         req.Stock,
		Image:         req.Image,
		Weight:        req.Weight,
		Volume:        req.Volume,
		Status:        req.Status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := l.skuRepo.Create(ctx, sku); err != nil {
		if isDuplicateSkuCodeErr(err) {
			return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
		}
		return nil, apperrors.NewInternalError("创建SKU失败: " + err.Error())
	}

	// 清除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, sku.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		// SKU 变更会影响商品详情（规格/价格）与商品列表（最低价/总库存）
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, sku.ProductID))
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &CreateSkuResponse{
		Sku: sku,
	}, nil
}

// UpdateSkuRequest 更新SKU请求
type UpdateSkuRequest struct {
	ID            uint64
	SkuCode       string
	Name          string
	Specs         map[string]string
	Price         float64
	OriginalPrice *float64
	Stock         int
	Image         string
	Weight        *float64
	Volume        *float64
	Status        int8
}

// UpdateSkuResponse 更新SKU响应
type UpdateSkuResponse struct {
	Sku *model.Sku
}

// UpdateSku 更新SKU
func (l *ProductLogic) UpdateSku(ctx context.Context, req *UpdateSkuRequest) (*UpdateSkuResponse, error) {
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：SKU 变更需要触发商品索引更新
	if l.db != nil && l.outboxRepo != nil {
		var outSku *model.Sku
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			skuRepoTx := repository.NewSkuRepository(tx)
			if req.ID == 0 {
				return apperrors.NewInvalidParamError("SKU ID不能为空")
			}
			sku, err := skuRepoTx.GetByID(ctx, req.ID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
				}
				return apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if sku == nil {
				return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
			}

			// 更新字段（复用原逻辑）
			if req.SkuCode != "" {
				if req.SkuCode != sku.SkuCode {
					existingSku, err := skuRepoTx.GetBySkuCode(ctx, req.SkuCode)
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return apperrors.NewInternalError("查询SKU失败: " + err.Error())
					}
					if existingSku != nil && existingSku.ID != req.ID {
						return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
					}
					deletedSku, err := skuRepoTx.GetBySkuCodeUnscoped(ctx, req.SkuCode)
					if err == nil && deletedSku != nil && deletedSku.ID != req.ID {
						return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
					}
				}
				sku.SkuCode = req.SkuCode
			}
			if req.Name != "" {
				sku.Name = req.Name
			}
			if len(req.Specs) > 0 {
				specsJSON, err := json.Marshal(req.Specs)
				if err != nil {
					return apperrors.NewInternalError("规格属性格式错误: " + err.Error())
				}
				sku.Specs = string(specsJSON)
			}
			if req.Price > 0 {
				sku.Price = req.Price
			}
			if req.OriginalPrice != nil {
				sku.OriginalPrice = req.OriginalPrice
			}
			if req.Stock >= 0 {
				sku.Stock = req.Stock
			}
			if req.Image != "" {
				sku.Image = req.Image
			}
			if req.Weight != nil {
				sku.Weight = req.Weight
			}
			if req.Volume != nil {
				sku.Volume = req.Volume
			}
			if req.Status >= 0 {
				sku.Status = req.Status
			}
			sku.UpdatedAt = time.Now()

			if err := skuRepoTx.Update(ctx, sku); err != nil {
				if isDuplicateSkuCodeErr(err) {
					return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
				}
				return apperrors.NewInternalError("更新SKU失败: " + err.Error())
			}
			outSku = sku

			// 同事务写 outbox
			payloadBytes, _ := json.Marshal(map[string]any{"product_id": sku.ProductID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", sku.ProductID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			return nil, err
		}

		// 清除缓存
		if outSku != nil && l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, outSku.ID))
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, outSku.ProductID))
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}

		return &UpdateSkuResponse{Sku: outSku}, nil
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("SKU ID不能为空")
	}

	// 获取现有SKU
	sku, err := l.skuRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
		}
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if sku == nil {
		return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
	}

	// 更新字段
	if req.SkuCode != "" {
		// 检查SKU编码是否与其他SKU重复
		if req.SkuCode != "" && req.SkuCode != sku.SkuCode {
			existingSku, err := l.skuRepo.GetBySkuCode(ctx, req.SkuCode)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if existingSku != nil && existingSku.ID != req.ID {
				return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
			}
			// 软删除记录也会占用唯一索引
			deletedSku, err := l.skuRepo.GetBySkuCodeUnscoped(ctx, req.SkuCode)
			if err == nil && deletedSku != nil && deletedSku.ID != req.ID {
				return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
			}
		}
		sku.SkuCode = req.SkuCode
	}
	if req.Name != "" {
		sku.Name = req.Name
	}
	if len(req.Specs) > 0 {
		specsJSON, err := json.Marshal(req.Specs)
		if err != nil {
			return nil, apperrors.NewInternalError("规格属性格式错误: " + err.Error())
		}
		sku.Specs = string(specsJSON)
	}
	if req.Price > 0 {
		sku.Price = req.Price
	}
	if req.OriginalPrice != nil {
		sku.OriginalPrice = req.OriginalPrice
	}
	if req.Stock >= 0 {
		sku.Stock = req.Stock
	}
	if req.Image != "" {
		sku.Image = req.Image
	}
	if req.Weight != nil {
		sku.Weight = req.Weight
	}
	if req.Volume != nil {
		sku.Volume = req.Volume
	}
	if req.Status >= 0 {
		sku.Status = req.Status
	}

	sku.UpdatedAt = time.Now()

	if err := l.skuRepo.Update(ctx, sku); err != nil {
		if isDuplicateSkuCodeErr(err) {
			return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
		}
		return nil, apperrors.NewInternalError("更新SKU失败: " + err.Error())
	}

	// 清除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, sku.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, sku.ProductID))
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &UpdateSkuResponse{
		Sku: sku,
	}, nil
}

// DeleteSkuRequest 删除SKU请求
type DeleteSkuRequest struct {
	ID uint64
}

// DeleteSkuResponse 删除SKU响应
type DeleteSkuResponse struct{}

// DeleteSku 删除SKU（软删除）
func (l *ProductLogic) DeleteSku(ctx context.Context, req *DeleteSkuRequest) (*DeleteSkuResponse, error) {
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：SKU 变更需要触发商品索引更新
	if l.db != nil && l.outboxRepo != nil {
		var productID uint64
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			skuRepoTx := repository.NewSkuRepository(tx)
			if req.ID == 0 {
				return apperrors.NewInvalidParamError("SKU ID不能为空")
			}
			sku, err := skuRepoTx.GetByID(ctx, req.ID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
				}
				return apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if sku == nil {
				return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
			}
			productID = sku.ProductID

			if err := skuRepoTx.Delete(ctx, req.ID); err != nil {
				return apperrors.NewInternalError("删除SKU失败: " + err.Error())
			}

			payloadBytes, _ := json.Marshal(map[string]any{"product_id": productID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", productID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			return nil, err
		}

		// 清除缓存
		if l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, req.ID))
			if productID > 0 {
				_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, productID))
			}
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}

		return &DeleteSkuResponse{}, nil
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("SKU ID不能为空")
	}

	// 检查SKU是否存在
	sku, err := l.skuRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
		}
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if sku == nil {
		return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
	}

	// 执行软删除
	if err := l.skuRepo.Delete(ctx, req.ID); err != nil {
		return nil, apperrors.NewInternalError("删除SKU失败: " + err.Error())
	}

	// 清除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, req.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, sku.ProductID))
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &DeleteSkuResponse{}, nil
}

// GetCategoryListRequest 获取类目列表请求
type GetCategoryListRequest struct {
	ParentID uint64
	Level    int8
	Status   int8
}

// GetCategoryListResponse 获取类目列表响应
type GetCategoryListResponse struct {
	Categories []*model.Category
}

// GetCategoryList 获取类目列表
func (l *ProductLogic) GetCategoryList(ctx context.Context, req *GetCategoryListRequest) (*GetCategoryListResponse, error) {
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	var categories []*model.Category
	var err error

	if req.ParentID > 0 {
		categories, err = l.categoryRepo.GetByParentID(ctx, req.ParentID)
	} else {
		categories, err = l.categoryRepo.GetAll(ctx, req.Status)
	}

	if err != nil {
		return nil, apperrors.NewInternalError("查询类目失败: " + err.Error())
	}

	return &GetCategoryListResponse{
		Categories: categories,
	}, nil
}

// GetCategoryTreeRequest 获取类目树请求
type GetCategoryTreeRequest struct {
	Status int8
}

// GetCategoryTreeResponse 获取类目树响应
type GetCategoryTreeResponse struct {
	Categories []*CategoryTreeNode
}

// CategoryTreeNode 类目树节点
type CategoryTreeNode struct {
	*model.Category
	Children []*CategoryTreeNode `json:"children"`
}

func (l *ProductLogic) GetCategoryTree(ctx context.Context, req *GetCategoryTreeRequest) (*GetCategoryTreeResponse, error) {
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	cacheKey := cache.BuildKey(cache.KeyPrefixCategoryTree, req.Status)

	if l.cache != nil {
		var cachedResp GetCategoryTreeResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	allCategories, err := l.categoryRepo.GetAll(ctx, req.Status)
	if err != nil {
		return nil, apperrors.NewInternalError("查询类目失败: " + err.Error())
	}

	tree := buildCategoryTree(allCategories)

	resp := &GetCategoryTreeResponse{
		Categories: tree,
	}

	if l.cache != nil {
		_ = l.cache.Set(ctx, cacheKey, resp, 30*time.Minute)
	}

	return resp, nil
}

func buildCategoryTree(categories []*model.Category) []*CategoryTreeNode {
	categoryMap := make(map[uint64]*CategoryTreeNode)
	for _, c := range categories {
		categoryMap[c.ID] = &CategoryTreeNode{
			Category: c,
			Children: []*CategoryTreeNode{},
		}
	}

	var rootNodes []*CategoryTreeNode
	for _, node := range categoryMap {
		if node.ParentID == 0 {
			rootNodes = append(rootNodes, node)
		} else if parentNode, exists := categoryMap[node.ParentID]; exists {
			parentNode.Children = append(parentNode.Children, node)
		}
	}

	return rootNodes
}

func (l *ProductLogic) clearCategoryTreeCache(ctx context.Context) {
	if l.cache == nil {
		return
	}

	status := []int8{-1, 0, 1, 2}
	for _, s := range status {
		cacheKey := cache.BuildKey(cache.KeyPrefixCategoryTree, s)
		_ = l.cache.Delete(ctx, cacheKey)
	}
}

// parseJSONArray 解析JSON数组字符串
func parseJSONArray(jsonStr string) ([]string, error) {
	if jsonStr == "" {
		return []string{}, nil
	}
	var result []string
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// parseJSONMap 解析JSON对象字符串
func parseJSONMap(jsonStr string) (map[string]string, error) {
	if jsonStr == "" {
		return map[string]string{}, nil
	}
	var result map[string]string
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}
