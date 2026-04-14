package client

import (
	"context"
	"fmt"

	inventoryv1 "ecommerce-system/api/inventory/v1"

	"google.golang.org/grpc"
)

// InventoryClient 库存服务客户端
type InventoryClient struct {
	conn    *grpc.ClientConn
	client  inventoryv1.InventoryServiceClient
	timeout RpcConf
}

// NewInventoryClient 创建库存服务客户端
func NewInventoryClient(conf RpcConf) (*InventoryClient, error) {
	conn, err := newConn(conf)
	if err != nil {
		return nil, fmt.Errorf("dial inventory service %s: %w", conf.Endpoint, err)
	}
	return &InventoryClient{
		conn:    conn,
		client:  inventoryv1.NewInventoryServiceClient(conn),
		timeout: conf,
	}, nil
}

// Close 关闭连接
func (c *InventoryClient) Close() error {
	return c.conn.Close()
}

// LockStock 锁定库存（下单时预占）
func (c *InventoryClient) LockStock(ctx context.Context, skuID int64, quantity int32, orderID int64, remark string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.LockStock(ctx, &inventoryv1.LockStockRequest{
		SkuId:    skuID,
		Quantity: quantity,
		OrderId:  orderID,
		Remark:   remark,
	})
	if err != nil {
		return fmt.Errorf("lock stock sku=%d qty=%d: %w", skuID, quantity, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("lock stock sku=%d: %s", skuID, resp.Message)
	}
	return nil
}

// UnlockStock 解锁库存（取消订单时释放预占）
func (c *InventoryClient) UnlockStock(ctx context.Context, skuID int64, quantity int32, orderID int64, remark string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.UnlockStock(ctx, &inventoryv1.UnlockStockRequest{
		SkuId:    skuID,
		Quantity: quantity,
		OrderId:  orderID,
		Remark:   remark,
	})
	if err != nil {
		return fmt.Errorf("unlock stock sku=%d qty=%d: %w", skuID, quantity, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("unlock stock sku=%d: %s", skuID, resp.Message)
	}
	return nil
}

// DeductStock 扣减库存（支付成功后从锁定库存转为已售）
func (c *InventoryClient) DeductStock(ctx context.Context, skuID int64, quantity int32, orderID int64, remark string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.DeductStock(ctx, &inventoryv1.DeductStockRequest{
		SkuId:    skuID,
		Quantity: quantity,
		OrderId:  orderID,
		Remark:   remark,
	})
	if err != nil {
		return fmt.Errorf("deduct stock sku=%d qty=%d: %w", skuID, quantity, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("deduct stock sku=%d: %s", skuID, resp.Message)
	}
	return nil
}

// RollbackStock 回退库存（退款后将已售库存还回可用）
func (c *InventoryClient) RollbackStock(ctx context.Context, skuID int64, quantity int32, orderID int64, remark string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.RollbackStock(ctx, &inventoryv1.RollbackStockRequest{
		SkuId:    skuID,
		Quantity: quantity,
		OrderId:  orderID,
		Remark:   remark,
	})
	if err != nil {
		return fmt.Errorf("rollback stock sku=%d qty=%d: %w", skuID, quantity, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("rollback stock sku=%d: %s", skuID, resp.Message)
	}
	return nil
}

// BatchGetInventory 批量查询库存
func (c *InventoryClient) BatchGetInventory(ctx context.Context, skuIDs []int64) ([]*inventoryv1.Inventory, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.BatchGetInventory(ctx, &inventoryv1.BatchGetInventoryRequest{SkuIds: skuIDs})
	if err != nil {
		return nil, fmt.Errorf("batch get inventory: %w", err)
	}
	return resp.Data, nil
}
