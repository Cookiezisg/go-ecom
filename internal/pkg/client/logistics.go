package client

import (
	"context"
	"fmt"

	logisticsv1 "ecommerce-system/api/logistics/v1"

	"google.golang.org/grpc"
)

// LogisticsClient 物流服务客户端
type LogisticsClient struct {
	conn    *grpc.ClientConn
	client  logisticsv1.LogisticsServiceClient
	timeout RpcConf
}

// NewLogisticsClient 创建物流服务客户端
func NewLogisticsClient(conf RpcConf) (*LogisticsClient, error) {
	conn, err := newConn(conf)
	if err != nil {
		return nil, fmt.Errorf("dial logistics service %s: %w", conf.Endpoint, err)
	}
	return &LogisticsClient{
		conn:    conn,
		client:  logisticsv1.NewLogisticsServiceClient(conn),
		timeout: conf,
	}, nil
}

// Close 关闭连接
func (c *LogisticsClient) Close() error {
	return c.conn.Close()
}

// CreateLogistics 为订单创建物流运单
func (c *LogisticsClient) CreateLogistics(ctx context.Context, orderID int64, orderNo, companyCode, receiverName, receiverPhone, receiverAddress string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.CreateLogistics(ctx, &logisticsv1.CreateLogisticsRequest{
		OrderId:         orderID,
		OrderNo:         orderNo,
		CompanyCode:     companyCode,
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		ReceiverAddress: receiverAddress,
	})
	if err != nil {
		return "", fmt.Errorf("create logistics for order %s: %w", orderNo, err)
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("create logistics for order %s: %s", orderNo, resp.Message)
	}
	if resp.Data == nil {
		return "", fmt.Errorf("create logistics: empty response for order %s", orderNo)
	}
	return resp.Data.LogisticsNo, nil
}
