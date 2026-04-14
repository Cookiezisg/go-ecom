package review

import (
	"github.com/zeromicro/go-zero/zrpc"

	"ecommerce-system/internal/pkg/client"
)

// Config 评价服务配置
type Config struct {
	zrpc.RpcServerConf
	Database DatabaseConfig
	MongoDB  *MongoDBConfig
	OrderRpc client.RpcConf // 订单服务地址，用于校验订单状态
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URI      string
	Database string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}
