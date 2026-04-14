package client

import (
	"context"
	"fmt"

	productv1 "ecommerce-system/api/product/v1"

	"google.golang.org/grpc"
)

// ProductClient 商品服务客户端
type ProductClient struct {
	conn    *grpc.ClientConn
	client  productv1.ProductServiceClient
	timeout RpcConf
}

// NewProductClient 创建商品服务客户端
func NewProductClient(conf RpcConf) (*ProductClient, error) {
	conn, err := newConn(conf)
	if err != nil {
		return nil, fmt.Errorf("dial product service %s: %w", conf.Endpoint, err)
	}
	return &ProductClient{
		conn:    conn,
		client:  productv1.NewProductServiceClient(conn),
		timeout: conf,
	}, nil
}

// Close 关闭连接
func (c *ProductClient) Close() error {
	return c.conn.Close()
}

// GetSku 根据 SKU ID 获取 SKU 信息
func (c *ProductClient) GetSku(ctx context.Context, skuID int64) (*productv1.Sku, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.GetSku(ctx, &productv1.GetSkuRequest{Id: skuID})
	if err != nil {
		return nil, fmt.Errorf("get sku %d: %w", skuID, err)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("sku %d not found", skuID)
	}
	return resp.Data, nil
}

// GetProduct 根据 product ID 获取商品信息
func (c *ProductClient) GetProduct(ctx context.Context, productID int64) (*productv1.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.GetProduct(ctx, &productv1.GetProductRequest{Id: productID})
	if err != nil {
		return nil, fmt.Errorf("get product %d: %w", productID, err)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("product %d not found", productID)
	}
	return resp.Data, nil
}
