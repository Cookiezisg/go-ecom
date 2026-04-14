package service

import (
	"context"
	"encoding/json"
	"fmt"

	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/service/inventory/model"
	"ecommerce-system/internal/service/inventory/repository"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// InventoryConsumer 库存 Kafka 消费者
// 消费 TopicInventoryDeducted 消息，将 Redis 的扣减结果异步持久化到 MySQL。
type InventoryConsumer struct {
	db               *gorm.DB
	inventoryRepo    repository.InventoryRepository
	inventoryLogRepo repository.InventoryLogRepository
}

// NewInventoryConsumer 创建库存消费者
func NewInventoryConsumer(
	db *gorm.DB,
	inventoryRepo repository.InventoryRepository,
	inventoryLogRepo repository.InventoryLogRepository,
) *InventoryConsumer {
	return &InventoryConsumer{
		db:               db,
		inventoryRepo:    inventoryRepo,
		inventoryLogRepo: inventoryLogRepo,
	}
}

// deductMessage Kafka 消息结构（与 DeductStock 发出的格式对齐）
type deductMessage struct {
	SkuID    uint64 `json:"sku_id"`
	Quantity int    `json:"quantity"`
	OrderID  uint64 `json:"order_id"`
	NewStock int64  `json:"new_stock"`
}

// Consume 消费库存扣减消息，幂等地将扣减结果写入 MySQL
func (c *InventoryConsumer) Consume(ctx context.Context, message *mq.Message) error {
	var msg deductMessage
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return fmt.Errorf("序列化消息数据失败: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &msg); err != nil {
		return fmt.Errorf("解析库存扣减消息失败: %w", err)
	}

	logx.Infof("消费库存扣减消息: sku_id=%d quantity=%d order_id=%d", msg.SkuID, msg.Quantity, msg.OrderID)

	// MySQL 原子扣减（locked_stock - quantity，sold_stock + quantity）
	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inventory := &model.Inventory{}
		if err := tx.Where("sku_id = ?", msg.SkuID).First(inventory).Error; err != nil {
			return fmt.Errorf("查询库存记录失败 sku_id=%d: %w", msg.SkuID, err)
		}

		// 检查幂等：sold_stock 已包含本次订单则跳过
		// 简单幂等策略：检查库存流水中是否已有该 order_id 的扣减记录
		var existCount int64
		if err := tx.Model(&model.InventoryLog{}).
			Where("sku_id = ? AND order_id = ? AND type = ?", msg.SkuID, msg.OrderID, 5).
			Count(&existCount).Error; err != nil {
			return fmt.Errorf("检查幂等失败: %w", err)
		}
		if existCount > 0 {
			logx.Infof("库存扣减消息已处理（幂等），跳过 sku_id=%d order_id=%d", msg.SkuID, msg.OrderID)
			return nil
		}

		// 执行 MySQL 扣减：locked_stock → sold_stock
		result := tx.Model(inventory).
			Where("sku_id = ? AND locked_stock >= ?", msg.SkuID, msg.Quantity).
			Updates(map[string]interface{}{
				"locked_stock": gorm.Expr("locked_stock - ?", msg.Quantity),
				"sold_stock":   gorm.Expr("sold_stock + ?", msg.Quantity),
			})
		if result.Error != nil {
			return fmt.Errorf("MySQL 扣减库存失败: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			logx.Errorf("库存扣减 MySQL 行未更新（可能锁库存不足）sku_id=%d order_id=%d", msg.SkuID, msg.OrderID)
		}

		// 写库存流水
		orderID := msg.OrderID
		log := &model.InventoryLog{
			SkuID:       msg.SkuID,
			OrderID:     &orderID,
			Type:        5, // 扣减
			Quantity:    msg.Quantity,
			BeforeStock: inventory.AvailableStock,
			AfterStock:  inventory.AvailableStock - msg.Quantity,
			Remark:      "Kafka同步扣减",
		}
		return tx.Create(log).Error
	})

	if err != nil {
		logx.Errorf("处理库存扣减消息失败 sku_id=%d: %v", msg.SkuID, err)
		return err
	}

	logx.Infof("库存扣减消息处理完成 sku_id=%d order_id=%d", msg.SkuID, msg.OrderID)
	return nil
}
