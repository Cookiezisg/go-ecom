// Package client 封装各下游 gRPC 服务客户端，统一管理连接、超时与重试。
package client

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// RpcConf 下游 gRPC 服务配置
type RpcConf struct {
	Endpoint string        // "host:port"，例如 "127.0.0.1:8081"
	Timeout  time.Duration // 单次调用超时，0 表示使用默认值 5s
}

// defaultTimeout 未配置时的默认调用超时
const defaultTimeout = 5 * time.Second

func (c *RpcConf) timeout() time.Duration {
	if c.Timeout <= 0 {
		return defaultTimeout
	}
	return c.Timeout
}

// newConn 创建 gRPC 连接（非阻塞，连接在首次 RPC 调用时建立）
func newConn(conf RpcConf) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		conf.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                20 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
}
