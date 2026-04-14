package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ExpiredOrder 超时待支付订单（带商品项，用于解锁库存）
type ExpiredOrder struct {
	ID     uint64
	Items  []*ExpiredOrderItem
}

// ExpiredOrderItem 订单商品项（仅解锁库存需要的字段）
type ExpiredOrderItem struct {
	SkuID    uint64
	Quantity int
}

// OrderRepository 订单仓库接口（用于定时任务）
type OrderRepository interface {
	// GetExpiredOrders 查询超时的待支付订单（含商品项）
	GetExpiredOrders(ctx context.Context, timeoutMinutes int) ([]*ExpiredOrder, error)
	// CancelOrders 批量取消订单（仅更新状态）
	CancelOrders(ctx context.Context, orderIDs []uint64) (int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓库
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// GetExpiredOrders 查询超时的待支付订单，联表取商品项
func (r *orderRepository) GetExpiredOrders(ctx context.Context, timeoutMinutes int) ([]*ExpiredOrder, error) {
	deadline := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)

	// 查超时订单 ID
	var orderIDs []uint64
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("id").
		Where("status = 1 AND created_at < ?", deadline).
		Pluck("id", &orderIDs).Error
	if err != nil || len(orderIDs) == 0 {
		return nil, err
	}

	// 按订单 ID 查商品项（sku_id, quantity）
	type rawItem struct {
		OrderID  uint64
		SkuID    uint64
		Quantity int
	}
	var rawItems []rawItem
	err = r.db.WithContext(ctx).
		Table("order_items").
		Select("order_id, sku_id, quantity").
		Where("order_id IN ?", orderIDs).
		Scan(&rawItems).Error
	if err != nil {
		return nil, err
	}

	// 按 order_id 分组
	itemMap := make(map[uint64][]*ExpiredOrderItem, len(orderIDs))
	for _, ri := range rawItems {
		itemMap[ri.OrderID] = append(itemMap[ri.OrderID], &ExpiredOrderItem{
			SkuID:    ri.SkuID,
			Quantity: ri.Quantity,
		})
	}

	orders := make([]*ExpiredOrder, 0, len(orderIDs))
	for _, id := range orderIDs {
		orders = append(orders, &ExpiredOrder{
			ID:    id,
			Items: itemMap[id],
		})
	}
	return orders, nil
}

// CancelOrders 批量取消订单（status=0 已取消）
func (r *orderRepository) CancelOrders(ctx context.Context, orderIDs []uint64) (int64, error) {
	result := r.db.WithContext(ctx).
		Table("orders").
		Where("id IN ? AND status = 1", orderIDs).
		Updates(map[string]interface{}{
			"status":      0,
			"cancel_time": time.Now(),
		})
	return result.RowsAffected, result.Error
}
