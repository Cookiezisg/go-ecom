package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RecommendItem 推荐商品（强类型，替代 map[string]interface{}）
type RecommendItem struct {
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	MainImage string  `json:"main_image"`
	Price     float64 `json:"price"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
}

// RecommendRepository 推荐仓库接口
type RecommendRepository interface {
	// GetPersonalizedRecommend 获取个性化推荐
	GetPersonalizedRecommend(ctx context.Context, userID uint64, limit int) ([]*RecommendItem, error)
	// GetSimilarProducts 获取相似商品
	GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*RecommendItem, error)
	// GetHotProducts 获取热门商品
	GetHotProducts(ctx context.Context, categoryID uint64, limit int) ([]*RecommendItem, error)
	// GetRealtimeRecommend 获取实时推荐
	GetRealtimeRecommend(ctx context.Context, userID uint64, limit int) ([]*RecommendItem, error)
}

type recommendRepository struct {
	redis *redis.Client
}

// NewRecommendRepository 创建推荐仓库
func NewRecommendRepository(redis *redis.Client) RecommendRepository {
	return &recommendRepository{redis: redis}
}

// hotProductsKey 返回热门商品 Redis Key
func hotProductsKey(categoryID uint64) string {
	if categoryID > 0 {
		return fmt.Sprintf("recommend:hot:category:%d", categoryID)
	}
	return "recommend:hot:products"
}

// personalizedKey 返回个性化推荐 Redis Key
func personalizedKey(userID uint64) string {
	return fmt.Sprintf("recommend:personalized:%d", userID)
}

// realtimeKey 返回实时推荐 Redis Key
func realtimeKey(userID uint64) string {
	return fmt.Sprintf("recommend:realtime:%d", userID)
}

// similarKey 返回相似商品 Redis Key
func similarKey(productID uint64) string {
	return fmt.Sprintf("recommend:similar:%d", productID)
}

// fetchFromRedis 从 Redis ZSet 读取推荐列表（score desc）
func (r *recommendRepository) fetchFromRedis(ctx context.Context, key string, limit int) ([]*RecommendItem, error) {
	results, err := r.redis.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil || len(results) == 0 {
		return []*RecommendItem{}, nil
	}

	items := make([]*RecommendItem, 0, len(results))
	for _, z := range results {
		var item RecommendItem
		if err := json.Unmarshal([]byte(z.Member.(string)), &item); err != nil {
			continue
		}
		item.Score = z.Score
		items = append(items, &item)
	}
	return items, nil
}

// GetPersonalizedRecommend 获取个性化推荐
func (r *recommendRepository) GetPersonalizedRecommend(ctx context.Context, userID uint64, limit int) ([]*RecommendItem, error) {
	return r.fetchFromRedis(ctx, personalizedKey(userID), limit)
}

// GetSimilarProducts 获取相似商品
func (r *recommendRepository) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*RecommendItem, error) {
	return r.fetchFromRedis(ctx, similarKey(productID), limit)
}

// GetHotProducts 获取热门商品（从 Redis ZSet 读取全量热门榜，按类目过滤）
func (r *recommendRepository) GetHotProducts(ctx context.Context, categoryID uint64, limit int) ([]*RecommendItem, error) {
	return r.fetchFromRedis(ctx, hotProductsKey(categoryID), limit)
}

// GetRealtimeRecommend 获取实时推荐（基于用户最近行为的实时推荐）
func (r *recommendRepository) GetRealtimeRecommend(ctx context.Context, userID uint64, limit int) ([]*RecommendItem, error) {
	return r.fetchFromRedis(ctx, realtimeKey(userID), limit)
}
