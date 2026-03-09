package service

import (
	"context"
	"ecommerce-system/internal/pkg/cache"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/service/inventory/model"
	"ecommerce-system/internal/service/inventory/repository"

	"errors"
	"fmt"

	"gorm.io/gorm"
)

type InventoryLogic struct {
	inventoryRepo    repository.InventoryRepository
	inventoryLogRepo repository.InventoryLogRepository
	cache            *cache.CacheOperations
	mqProducer       *mq.Producer
}

func NewInventoryLogic(
	inventoryRepo repository.InventoryRepository,
	inventoryLogRepo repository.InventoryLogRepository,
	cache *cache.CacheOperations,
	mqProducer *mq.Producer,
) *InventoryLogic {
	return &InventoryLogic{
		inventoryRepo:    inventoryRepo,
		inventoryLogRepo: inventoryLogRepo,
		cache:            cache,
		mqProducer:       mqProducer,
	}
}

type GetInventoryRequest struct {
	SkuID uint64
}

type GetInventoryResponse struct {
	Inventory *model.Inventory
}

// GetInventory 获取库存（优先从Redis读取）
func (l *InventoryLogic) GetInventory(ctx context.Context, req *GetInventoryRequest) (*GetInventoryResponse, error) {
	// 优先从Redis读取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		stockStr, err := l.cache.Get(ctx, cacheKey)
		if err == nil && stockStr != "" {
			// 从Redis读取成功，构造返回对象
			var stock int
			fmt.Sscanf(stockStr, "%d", &stock)
			inventory := &model.Inventory{
				SkuID:          req.SkuID,
				AvailableStock: stock,
			}
			return &GetInventoryResponse{Inventory: inventory}, nil
		}
	}

	// 从数据库读取
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("库存不存在")
		}
		return nil, apperrors.NewInternalError("查询库存失败: " + err.Error())
	}
	if inventory == nil {
		return nil, apperrors.NewNotFoundError("库存不存在")
	}

	// 同步到Redis
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		_ = l.cache.Set(ctx, cacheKey, inventory.AvailableStock, 0) // 永不过期
	}

	return &GetInventoryResponse{
		Inventory: inventory,
	}, nil
}

// BatchGetInventoryRequest 批量获取库存请求
type BatchGetInventoryRequest struct {
	SkuIDs []uint64
}

// BatchGetInventoryResponse 批量获取库存响应
type BatchGetInventoryResponse struct {
	Inventories []*model.Inventory
}

// BatchGetInventory 批量获取库存
func (l *InventoryLogic) BatchGetInventory(ctx context.Context, req *BatchGetInventoryRequest) (*BatchGetInventoryResponse, error) {
	inventories, err := l.inventoryRepo.BatchGetBySkuIDs(ctx, req.SkuIDs)
	if err != nil {
		return nil, apperrors.NewInternalError("批量获取库存失败")
	}

	return &BatchGetInventoryResponse{
		Inventories: inventories,
	}, nil
}
