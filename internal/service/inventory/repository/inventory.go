package repository

import (
	"context"
	"ecommerce-system/internal/service/inventory/model"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	// GetBySkuID 根据SKU ID获取库存
	GetBySkuID(ctx context.Context, skuID uint64) (*model.Inventory, error)
	// BatchGetBySkuIDs 批量获取库存
	BatchGetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*model.Inventory, error)
	// Create 创建库存记录
	Create(ctx context.Context, inventory *model.Inventory) error
	// Update 更新库存
	Update(ctx context.Context, inventory *model.Inventory) error
	// LockStock 锁定库存（使用乐观锁）
	LockStock(ctx context.Context, skuID uint64, quantity int) error
	// DeductStock 扣减库存
	DeductStock(ctx context.Context, skuID uint64, quantity int) error
	// UnlockStock 解锁库存a
	UnlockStock(ctx context.Context, skuID uint64, quantity int) error
	// RollbackStock 回退库存
	RollbackStock(ctx context.Context, skuID uint64, quantity int) error
	// StockIn 入库
	StockIn(ctx context.Context, skuID uint64, quantity int) error
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}
