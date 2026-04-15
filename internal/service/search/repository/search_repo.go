package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"ecommerce-system/internal/pkg/search"
)

// SearchRepository 搜索仓库接口
type SearchRepository interface {
	// SearchProducts 搜索商品
	SearchProducts(ctx context.Context, keyword string, categoryID uint64, page, pageSize int, sortBy string) ([]map[string]interface{}, int64, error)
	// GetSearchSuggestions 获取搜索建议
	GetSearchSuggestions(ctx context.Context, keyword string, limit int) ([]string, error)
	// GetHotKeywords 获取搜索热词
	GetHotKeywords(ctx context.Context, limit int) ([]string, error)
	// BuildProductIndex 构建商品索引
	BuildProductIndex(ctx context.Context, productIDs []uint64) error
}

type searchRepository struct {
	redis        *redis.Client
	esClient     *search.Client
	snapshotRepo ProductSnapshotRepository
}

// NewSearchRepository 创建搜索仓库
func NewSearchRepository(redis *redis.Client, esClient *search.Client, snapshotRepo ProductSnapshotRepository) SearchRepository {
	return &searchRepository{
		redis:        redis,
		esClient:     esClient,
		snapshotRepo: snapshotRepo,
	}
}

// SearchProducts 搜索商品（使用Elasticsearch）
func (r *searchRepository) SearchProducts(ctx context.Context, keyword string, categoryID uint64, page, pageSize int, sortBy string) ([]map[string]interface{}, int64, error) {
	// 如果Elasticsearch不可用，返回空结果
	if r.esClient == nil {
		return []map[string]interface{}{}, 0, nil
	}

	// 分页默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 构建查询
	query := map[string]interface{}{
		"from":  (page - 1) * pageSize,
		"size":  pageSize,
		"query": map[string]interface{}{},
	}

	// 构建查询条件
	mustClauses := []map[string]interface{}{}

	// 关键词搜索
	if keyword != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  keyword,
				"fields": []string{"name^3", "subtitle^2", "detail"},
				"type":   "best_fields",
			},
		})
	}

	// 类目筛选
	if categoryID > 0 {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"category_id": categoryID,
			},
		})
	}

	// 状态筛选（只搜索上架商品）
	mustClauses = append(mustClauses, map[string]interface{}{
		"term": map[string]interface{}{
			"status": 1,
		},
	})

	if len(mustClauses) > 0 {
		query["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		}
	} else {
		query["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	// 排序
	sort := []map[string]interface{}{}
	switch sortBy {
	case "sales":
		sort = append(sort, map[string]interface{}{"sales": "desc"})
	case "price_asc":
		sort = append(sort, map[string]interface{}{"price": "asc"})
	case "price_desc":
		sort = append(sort, map[string]interface{}{"price": "desc"})
	default:
		// 默认：相关性 + 销量
		sort = append(sort, map[string]interface{}{"_score": "desc"})
		sort = append(sort, map[string]interface{}{"sales": "desc"})
	}
	query["sort"] = sort

	// 执行搜索
	results, total, err := r.esClient.Search(ctx, ProductIndexName, query)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索失败: %w", err)
	}

	return results, total, nil
}

// GetSearchSuggestions 获取搜索建议
func (r *searchRepository) GetSearchSuggestions(ctx context.Context, keyword string, limit int) ([]string, error) {
	// 从Redis获取搜索建议
	key := "search:suggestions:" + keyword
	suggestions, err := r.redis.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return []string{}, nil
	}
	return suggestions, nil
}

// GetHotKeywords 获取搜索热词
func (r *searchRepository) GetHotKeywords(ctx context.Context, limit int) ([]string, error) {
	// 从Redis获取搜索热词
	key := "search:hot:keywords"
	keywords, err := r.redis.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return []string{}, nil
	}
	return keywords, nil
}

// BuildProductIndex 构建商品索引，将商品数据从 MySQL 同步到 Elasticsearch
// productIDs 为空时自动全量索引所有商品
func (r *searchRepository) BuildProductIndex(ctx context.Context, productIDs []uint64) error {
	if r.esClient == nil {
		return fmt.Errorf("Elasticsearch 客户端未初始化")
	}
	if r.snapshotRepo == nil {
		return fmt.Errorf("snapshotRepo 未初始化")
	}

	// 未指定 ID 时全量索引：和 ES 现有文档做 diff，发现 stale 文档则重建索引
	if len(productIDs) == 0 {
		var err error
		productIDs, err = r.snapshotRepo.ListAllProductIDs(ctx)
		if err != nil {
			return fmt.Errorf("获取商品列表失败: %w", err)
		}

		// 构建 MySQL ID 集合
		mysqlIDs := make(map[string]struct{}, len(productIDs))
		for _, id := range productIDs {
			mysqlIDs[fmt.Sprintf("%d", id)] = struct{}{}
		}

		// 拉取 ES 中现有的所有文档 ID
		esIDs, err := r.esClient.ListAllDocumentIDs(ctx, ProductIndexName)
		if err != nil {
			return fmt.Errorf("获取 ES 文档列表失败: %w", err)
		}

		// 检测 stale 文档（ES 有、MySQL 没有）
		stale := 0
		for _, esID := range esIDs {
			if _, ok := mysqlIDs[esID]; !ok {
				stale++
			}
		}

		// 有 stale 文档时重建整个索引，否则只做 upsert
		if stale > 0 {
			if err := r.esClient.DeleteIndex(ctx, ProductIndexName); err != nil {
				return fmt.Errorf("删除旧索引失败: %w", err)
			}
			if err := r.esClient.CreateIndex(ctx, ProductIndexName, ProductIndexMapping); err != nil {
				return fmt.Errorf("重建索引失败: %w", err)
			}
		}
	}

	for _, id := range productIDs {
		doc, err := r.snapshotRepo.BuildProductDocument(ctx, id)
		if err != nil {
			return fmt.Errorf("构建商品 %d 文档失败: %w", id, err)
		}
		docID := fmt.Sprintf("%d", id)
		if err := r.esClient.IndexDocument(ctx, ProductIndexName, docID, doc); err != nil {
			return fmt.Errorf("索引商品 %d 失败: %w", id, err)
		}
	}
	return nil
}
