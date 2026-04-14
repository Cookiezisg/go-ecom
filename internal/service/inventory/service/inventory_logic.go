package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecommerce-system/internal/pkg/cache"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/service/inventory/model"
	"ecommerce-system/internal/service/inventory/repository"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// InventoryLogic 库存业务逻辑
type InventoryLogic struct {
	inventoryRepo    repository.InventoryRepository
	inventoryLogRepo repository.InventoryLogRepository
	cache            *cache.CacheOperations
	mqProducer       *mq.Producer
}

// NewInventoryLogic 创建库存业务逻辑
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

// -----------------------------------------------------------------------
// GetInventory
// -----------------------------------------------------------------------

// GetInventoryRequest 获取库存请求
type GetInventoryRequest struct {
	SkuID uint64
}

// GetInventoryResponse 获取库存响应
type GetInventoryResponse struct {
	Inventory *model.Inventory
}

// GetInventory 获取库存（Redis 优先，缺失时加载 DB 并同步）
func (l *InventoryLogic) GetInventory(ctx context.Context, req *GetInventoryRequest) (*GetInventoryResponse, error) {
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		stockStr, err := l.cache.Get(ctx, cacheKey)
		if err == nil && stockStr != "" {
			var stock int
			fmt.Sscanf(stockStr, "%d", &stock)
			return &GetInventoryResponse{
				Inventory: &model.Inventory{SkuID: req.SkuID, AvailableStock: stock},
			}, nil
		}
	}

	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeStockNotFound)
		}
		return nil, apperrors.NewInternalError("查询库存失败: " + err.Error())
	}
	if inventory == nil {
		return nil, apperrors.NewError(apperrors.CodeStockNotFound)
	}

	// 同步到 Redis（永不过期，由写操作维护）
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		_ = l.cache.Set(ctx, cacheKey, inventory.AvailableStock, 0)
	}

	return &GetInventoryResponse{Inventory: inventory}, nil
}

// -----------------------------------------------------------------------
// BatchGetInventory
// -----------------------------------------------------------------------

// BatchGetInventoryRequest 批量获取库存请求
type BatchGetInventoryRequest struct {
	SkuIDs []uint64
}

// BatchGetInventoryResponse 批量获取库存响应
type BatchGetInventoryResponse struct {
	Inventories []*model.Inventory
}

// BatchGetInventory 批量获取库存（直接读 DB）
func (l *InventoryLogic) BatchGetInventory(ctx context.Context, req *BatchGetInventoryRequest) (*BatchGetInventoryResponse, error) {
	inventories, err := l.inventoryRepo.BatchGetBySkuIDs(ctx, req.SkuIDs)
	if err != nil {
		return nil, apperrors.NewInternalError("批量获取库存失败: " + err.Error())
	}
	return &BatchGetInventoryResponse{Inventories: inventories}, nil
}

// -----------------------------------------------------------------------
// LockStock（下单预占）
// -----------------------------------------------------------------------

// LockStockRequest 锁定库存请求
type LockStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// LockStock 锁定库存（预占）：MySQL 原子更新，同步失效 Redis 缓存
func (l *InventoryLogic) LockStock(ctx context.Context, req *LockStockRequest) error {
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil || inventory == nil {
		return apperrors.NewError(apperrors.CodeStockNotFound, "库存记录不存在")
	}

	if inventory.AvailableStock < req.Quantity {
		return apperrors.NewError(apperrors.CodeStockInsufficient)
	}

	if err := l.inventoryRepo.LockStock(ctx, req.SkuID, req.Quantity); err != nil {
		return apperrors.NewInternalError("锁定库存失败: " + err.Error())
	}

	// 使 Redis 缓存失效，避免后续读到脏数据
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		_ = l.cache.Delete(ctx, cacheKey)
	}

	_ = l.inventoryLogRepo.Create(ctx, &model.InventoryLog{
		SkuID:       req.SkuID,
		OrderID:     &req.OrderID,
		Type:        3, // 锁定
		Quantity:    req.Quantity,
		BeforeStock: inventory.AvailableStock,
		AfterStock:  inventory.AvailableStock - req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	})

	return nil
}

// -----------------------------------------------------------------------
// DeductStock（支付成功后扣减）
// -----------------------------------------------------------------------

// DeductStockRequest 扣减库存请求
type DeductStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// DeductStock 扣减库存（原子 Lua 脚本，Redis 优先，再 Kafka 异步同步 MySQL）
// 修复说明：旧实现先 GET 再 DECRBY 存在 TOCTOU 竞态，高并发下会超卖。
// 新实现用 AtomicDeductStock（单条 Lua 脚本）保证原子性。
func (l *InventoryLogic) DeductStock(ctx context.Context, req *DeductStockRequest) error {
	if l.cache == nil {
		// 无缓存时降级为 MySQL 直扣
		return l.deductFromDB(ctx, req)
	}

	cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)

	newStock, err := l.cache.AtomicDeductStock(ctx, cacheKey, int64(req.Quantity))
	if err != nil {
		logx.Errorf("AtomicDeductStock Redis 错误 sku_id=%d: %v，降级到 DB", req.SkuID, err)
		return l.deductFromDB(ctx, req)
	}

	if newStock < 0 {
		// -1：key 不存在或库存不足；重新从 DB 加载后重试
		inventory, dbErr := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
		if dbErr != nil || inventory == nil {
			return apperrors.NewError(apperrors.CodeStockNotFound)
		}
		if inventory.AvailableStock < req.Quantity {
			return apperrors.NewError(apperrors.CodeStockInsufficient)
		}
		// 写入缓存后重试
		_ = l.cache.Set(ctx, cacheKey, inventory.AvailableStock, 0)
		newStock, err = l.cache.AtomicDeductStock(ctx, cacheKey, int64(req.Quantity))
		if err != nil || newStock < 0 {
			return apperrors.NewError(apperrors.CodeStockInsufficient)
		}
	}

	// 异步通过 Kafka 同步到 MySQL
	if l.mqProducer != nil {
		msg := mq.NewMessage(mq.TopicInventoryDeducted, map[string]interface{}{
			"sku_id":    req.SkuID,
			"quantity":  req.Quantity,
			"order_id":  req.OrderID,
			"new_stock": newStock,
		})
		_ = l.mqProducer.PublishWithKey(ctx, mq.TopicInventoryDeducted, fmt.Sprintf("%d", req.SkuID), msg)
	} else {
		// 无 Kafka 时直接同步 DB
		return l.deductFromDB(ctx, req)
	}

	return nil
}

// deductFromDB MySQL 直接扣减库存（降级路径）
func (l *InventoryLogic) deductFromDB(ctx context.Context, req *DeductStockRequest) error {
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil || inventory == nil {
		return apperrors.NewError(apperrors.CodeStockNotFound)
	}
	if inventory.AvailableStock < req.Quantity {
		return apperrors.NewError(apperrors.CodeStockInsufficient)
	}

	if err := l.inventoryRepo.DeductStock(ctx, req.SkuID, req.Quantity); err != nil {
		return apperrors.NewInternalError("扣减库存失败: " + err.Error())
	}

	_ = l.inventoryLogRepo.Create(ctx, &model.InventoryLog{
		SkuID:       req.SkuID,
		OrderID:     &req.OrderID,
		Type:        5, // 扣减
		Quantity:    req.Quantity,
		BeforeStock: inventory.AvailableStock,
		AfterStock:  inventory.AvailableStock - req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	})

	return nil
}

// -----------------------------------------------------------------------
// UnlockStock（取消订单释放预占）
// -----------------------------------------------------------------------

// UnlockStockRequest 解锁库存请求
type UnlockStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// UnlockStock 解锁库存，同步更新 Redis 缓存
func (l *InventoryLogic) UnlockStock(ctx context.Context, req *UnlockStockRequest) error {
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil || inventory == nil {
		return apperrors.NewError(apperrors.CodeStockNotFound)
	}

	if err := l.inventoryRepo.UnlockStock(ctx, req.SkuID, req.Quantity); err != nil {
		return apperrors.NewInternalError("解锁库存失败: " + err.Error())
	}

	// 解锁后 available_stock 增加，同步更新 Redis
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		_, _ = l.cache.AtomicRollbackStock(ctx, cacheKey, int64(req.Quantity))
	}

	_ = l.inventoryLogRepo.Create(ctx, &model.InventoryLog{
		SkuID:       req.SkuID,
		OrderID:     &req.OrderID,
		Type:        4, // 解锁
		Quantity:    req.Quantity,
		BeforeStock: inventory.LockedStock,
		AfterStock:  inventory.LockedStock - req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	})

	return nil
}

// -----------------------------------------------------------------------
// RollbackStock（退款回退已售库存）
// -----------------------------------------------------------------------

// RollbackStockRequest 回退库存请求
type RollbackStockRequest struct {
	SkuID    uint64
	Quantity int
	OrderID  uint64
	Remark   string
}

// RollbackStock 回退库存（退款场景），同步更新 Redis
func (l *InventoryLogic) RollbackStock(ctx context.Context, req *RollbackStockRequest) error {
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil || inventory == nil {
		return apperrors.NewError(apperrors.CodeStockNotFound)
	}

	if err := l.inventoryRepo.RollbackStock(ctx, req.SkuID, req.Quantity); err != nil {
		return apperrors.NewInternalError("回退库存失败: " + err.Error())
	}

	// 回退后 available_stock 增加，同步 Redis
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		_, _ = l.cache.AtomicRollbackStock(ctx, cacheKey, int64(req.Quantity))
	}

	_ = l.inventoryLogRepo.Create(ctx, &model.InventoryLog{
		SkuID:       req.SkuID,
		OrderID:     &req.OrderID,
		Type:        6, // 回退
		Quantity:    req.Quantity,
		BeforeStock: inventory.SoldStock,
		AfterStock:  inventory.SoldStock - req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	})

	return nil
}

// -----------------------------------------------------------------------
// StockIn（入库）
// -----------------------------------------------------------------------

// StockInRequest 入库请求
type StockInRequest struct {
	SkuID    uint64
	Quantity int
	Remark   string
}

// StockIn 入库，同步更新 Redis
func (l *InventoryLogic) StockIn(ctx context.Context, req *StockInRequest) error {
	inventory, err := l.inventoryRepo.GetBySkuID(ctx, req.SkuID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.NewInternalError("查询库存失败: " + err.Error())
	}

	if inventory == nil {
		inventory = &model.Inventory{
			SkuID:             req.SkuID,
			TotalStock:        req.Quantity,
			AvailableStock:    req.Quantity,
			LockedStock:       0,
			SoldStock:         0,
			LowStockThreshold: 10,
		}
		if err := l.inventoryRepo.Create(ctx, inventory); err != nil {
			return apperrors.NewInternalError("创建库存记录失败: " + err.Error())
		}
	} else {
		if err := l.inventoryRepo.StockIn(ctx, req.SkuID, req.Quantity); err != nil {
			return apperrors.NewInternalError("入库失败: " + err.Error())
		}
	}

	// 同步 Redis
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixInventoryStock, req.SkuID)
		_, _ = l.cache.AtomicRollbackStock(ctx, cacheKey, int64(req.Quantity))
	}

	_ = l.inventoryLogRepo.Create(ctx, &model.InventoryLog{
		SkuID:       req.SkuID,
		Type:        1, // 入库
		Quantity:    req.Quantity,
		BeforeStock: inventory.AvailableStock,
		AfterStock:  inventory.AvailableStock + req.Quantity,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
	})

	return nil
}

// -----------------------------------------------------------------------
// GetInventoryLog
// -----------------------------------------------------------------------

// GetInventoryLogRequest 获取库存流水请求
type GetInventoryLogRequest struct {
	SkuID    uint64
	Page     int
	PageSize int
}

// GetInventoryLogResponse 获取库存流水响应
type GetInventoryLogResponse struct {
	Logs  []*model.InventoryLog
	Total int64
}

// GetInventoryLog 获取库存流水
func (l *InventoryLogic) GetInventoryLog(ctx context.Context, req *GetInventoryLogRequest) (*GetInventoryLogResponse, error) {
	logs, total, err := l.inventoryLogRepo.GetBySkuID(ctx, req.SkuID, req.Page, req.PageSize)
	if err != nil {
		return nil, apperrors.NewInternalError("获取库存流水失败: " + err.Error())
	}
	return &GetInventoryLogResponse{Logs: logs, Total: total}, nil
}
