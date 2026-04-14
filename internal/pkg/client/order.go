package client

import (
	"context"
	"fmt"

	orderv1 "ecommerce-system/api/order/v1"

	"google.golang.org/grpc"
)

// OrderClient 订单服务客户端
type OrderClient struct {
	conn    *grpc.ClientConn
	client  orderv1.OrderServiceClient
	timeout RpcConf
}

// NewOrderClient 创建订单服务客户端
func NewOrderClient(conf RpcConf) (*OrderClient, error) {
	conn, err := newConn(conf)
	if err != nil {
		return nil, fmt.Errorf("dial order service %s: %w", conf.Endpoint, err)
	}
	return &OrderClient{
		conn:    conn,
		client:  orderv1.NewOrderServiceClient(conn),
		timeout: conf,
	}, nil
}

// Close 关闭连接
func (c *OrderClient) Close() error {
	return c.conn.Close()
}

// PayOrder 通知订单服务支付成功（订单状态：待支付→待发货）
func (c *OrderClient) PayOrder(ctx context.Context, orderID int64, orderNo, paymentNo string, paymentMethod int32) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.PayOrder(ctx, &orderv1.PayOrderRequest{
		OrderId:       orderID,
		OrderNo:       orderNo,
		PaymentMethod: paymentMethod,
		PaymentNo:     paymentNo,
	})
	if err != nil {
		return fmt.Errorf("pay order %s: %w", orderNo, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("pay order %s: %s", orderNo, resp.Message)
	}
	return nil
}

// CancelOrder 取消订单
func (c *OrderClient) CancelOrder(ctx context.Context, orderID int64, orderNo, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		Id:      orderID,
		OrderNo: orderNo,
		Reason:  reason,
	})
	if err != nil {
		return fmt.Errorf("cancel order %s: %w", orderNo, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("cancel order %s: %s", orderNo, resp.Message)
	}
	return nil
}

// RefundOrder 通知订单服务退款完成（订单状态→已退款）
func (c *OrderClient) RefundOrder(ctx context.Context, orderID int64, orderNo, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.RefundOrder(ctx, &orderv1.RefundOrderRequest{
		OrderId: orderID,
		OrderNo: orderNo,
		Reason:  reason,
	})
	if err != nil {
		return fmt.Errorf("refund order %s: %w", orderNo, err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("refund order %s: %s", orderNo, resp.Message)
	}
	return nil
}

// GetOrder 获取订单详情（支付服务需要查订单 items 来回滚库存）
func (c *OrderClient) GetOrder(ctx context.Context, orderID int64, orderNo string) (*orderv1.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.GetOrder(ctx, &orderv1.GetOrderRequest{
		Id:      orderID,
		OrderNo: orderNo,
	})
	if err != nil {
		return nil, fmt.Errorf("get order %s: %w", orderNo, err)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("order %s not found", orderNo)
	}
	return resp.Data, nil
}
