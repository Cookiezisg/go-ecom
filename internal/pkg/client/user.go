package client

import (
	"context"
	"fmt"

	userv1 "ecommerce-system/api/user/v1"

	"google.golang.org/grpc"
)

// UserClient 用户服务客户端
type UserClient struct {
	conn    *grpc.ClientConn
	client  userv1.UserServiceClient
	timeout RpcConf
}

// NewUserClient 创建用户服务客户端
func NewUserClient(conf RpcConf) (*UserClient, error) {
	conn, err := newConn(conf)
	if err != nil {
		return nil, fmt.Errorf("dial user service %s: %w", conf.Endpoint, err)
	}
	return &UserClient{
		conn:    conn,
		client:  userv1.NewUserServiceClient(conn),
		timeout: conf,
	}, nil
}

// Close 关闭连接
func (c *UserClient) Close() error {
	return c.conn.Close()
}

// GetAddressByID 获取用户指定地址，address_id=0 时返回默认地址
func (c *UserClient) GetAddressByID(ctx context.Context, userID, addressID int64) (*userv1.Address, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.GetAddressList(ctx, &userv1.GetAddressListRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("get address list: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("user %d has no address", userID)
	}

	// 找指定地址
	if addressID > 0 {
		for _, addr := range resp.Data {
			if addr.Id == addressID {
				return addr, nil
			}
		}
		return nil, fmt.Errorf("address %d not found for user %d", addressID, userID)
	}

	// 找默认地址
	for _, addr := range resp.Data {
		if addr.IsDefault == 1 {
			return addr, nil
		}
	}

	// 没有默认地址，返回第一个
	return resp.Data[0], nil
}
