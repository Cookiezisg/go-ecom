package repository

import (
	"context"
	"ecommerce-system/internal/service/cart/model"

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
	return &cartRepository{db: db, redis: redis}
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID uint64) ([]*model.Cart, error) {
}

func (r *cartRepository) AddItem(ctx context.Context, cart *model.Cart) error {
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
