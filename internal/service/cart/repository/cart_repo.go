package repository

import (
	"context"
	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/service/cart/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CartRepository interface {
	GetByUserID(ctx context.Context, userID uint64) ([]*model.Cart, error)
	AddItem(ctx context.Context, cart *model.Cart) error
	UpdateQuantity(ctx context.Context, userID, skuID uint64, quantity int) error
	RemoveItem(ctx context.Context, userID, skuIDs []uint64) error
	ClearCart(ctx context.Context, userID uint64) error
	SelectItem(ctx context.Context, userID, skuIDs []uint64, isSelected int8) error
	BatchSelect(ctx context.Context, userID uint64, skuIDs []uint64, isSelected int8) error
	SyncToDB(ctx context.Context, userID uint64) error
}

type cartRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewCartRepository(db *gorm.DB, redis *redis.Client) CartRepository {
	return &cartRepository{
		db:    db,
		redis: redis,
	}
}

// 辅助函数 拿到redis的key
func (r *cartRepository) getCartKey(userID uint64) string {
	return cache.BuildKey(cache.KeyPrefixCart, "user:", userID)
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID uint64) ([]*model.Cart, error) {
	key := r.getCartKey(userID)

	items, err := r.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return []*model.Cart{}, nil
	}

	carts := make([]*model.Cart, 0, len(items))

	for _, itemJSON := range items {
		var item model.Cart
		if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
			continue
		}
		carts = append(carts, &item)
	}

	return carts, nil
}

func (r *cartRepository) AddItem(ctx context.Context, cart *model.Cart) error {
	key := r.getCartKey(cart.UserID)
	itemKey := fmt.Sprints("%d", cart.SkuID)

	exists, err := r.redis.HExists(ctx, key, itemKey).Result()
	if err != nil {
		return err
	}

	if exists {
		// 更新数量
		var existingCart model.Cart
		itemJson, _ := r.redis.HGet(ctx, key, itemKey).Result()
		json.Unmarshal([]byte(itemJson), &existingCart)
		existingCart.Quantity += cart.Quantity
		existingCart.UpdatedAt = time.Now()

		itemJSONBytes, _ := json.Marshal(existingCart)
		err := r.redis.HSet(ctx, key, itemKey, itemJSONBytes).Err()
		if err != nil {
			return err
		}

		return r.redis.Expire(ctx, key, 20*24*time.Hour).Err()
	}

	// 如果是新增商品的话
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = time.Now()
	itemJSONBytes, _ := json.Marshal(cart)
	err = r.redis.HSet(ctx, key, itemKey, itemJSONBytes).Err()
	if err != nil {
		return err
	}

	return r.redis.Expire(ctx, key, 20*24*time.Hour).Err()
}

func (r *cartRepository) UpdateQuantity(ctx context.Context, userID, skuID uint64, quantity int) error {
}

func (r *cartRepository) RemoveItem(ctx context.Context, userID, skuIDs []uint64) error {
}

func (r *cartRepository) ClearCart(ctx context.Context, userID uint64) error {
}

func (r *cartRepository) SelectItem(ctx context.Context, userID, skuIDs []uint64, isSelected int8) error {
}

func (r *cartRepository) BatchSelect(ctx context.Context, userID uint64, skuIDs []uint64, isSelected int8) error {
}

func (r *cartRepository) SyncToDB(ctx context.Context, userID uint64) error {
}
