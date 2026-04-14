package payment

import (
	"github.com/zeromicro/go-zero/zrpc"

	"ecommerce-system/internal/pkg/client"
)

// Config 支付服务配置
type Config struct {
	zrpc.RpcServerConf
	Database     DatabaseConfig
	BizRedis     RedisConfig // 业务侧使用的 Redis 配置
	Payment      PaymentConfig
	OrderRpc     client.RpcConf // 订单服务地址（支付成功后回调）
	InventoryRpc client.RpcConf // 库存服务地址（退款时回退库存）
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

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	WeChat WeChatConfig
	Alipay AlipayConfig
}

// WeChatConfig 微信支付配置
type WeChatConfig struct {
	AppID     string
	MchID     string
	APIKey    string
	NotifyURL string
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID      string
	PrivateKey string
	PublicKey  string
	NotifyURL  string
}
