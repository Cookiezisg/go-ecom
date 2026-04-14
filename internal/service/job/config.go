package job

import (
	"github.com/zeromicro/go-zero/zrpc"

	"ecommerce-system/internal/pkg/client"
)

// Config 定时任务服务配置
type Config struct {
	zrpc.RpcServerConf
	Database     DatabaseConfig
	InventoryRpc client.RpcConf // 库存服务地址，取消订单时解锁库存
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
