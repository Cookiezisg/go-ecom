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
	redis    *redis.Client
	esClient *search.Client
}

// NewSearchRepository 创建搜索仓库
func NewSearchRepository(redis *redis.Client, esClient *search.Client) SearchRepository {
	return &searchRepository{
		redis:    redis,
		esClient: esClient,
	}
}

// SearchProducts 搜索商品（使用Elasticsearch）
func (r *searchRepository) SearchProducts(ctx context.Context, keyword string, categoryID uint64, page, pageSize int, sortBy string) ([]map[string]interface{}, int64, error) {
	// 如果Elasticsearch不可用，返回空结果
	if r.esClient == nil {
		return []map[string]interface{}{}, 0, nil
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
			"status": 1, // 1-上架
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

// BuildProductIndex 构建商品索引
func (r *searchRepository) BuildProductIndex(ctx context.Context, productIDs []uint64) error {
	// 实际实现应该将商品数据索引到Elasticsearch
	// 这里简化处理
	return nil
}
