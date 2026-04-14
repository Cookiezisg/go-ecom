package order

import (
	"github.com/zeromicro/go-zero/zrpc"

	"ecommerce-system/internal/pkg/client"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // 秒
	ConnMaxIdleTime int // 秒
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

// Config 订单服务配置
type Config struct {
	zrpc.RpcServerConf
	Database      DatabaseConfig
	BizRedis      RedisConfig    // 业务侧使用的 Redis 配置
	Kafka         KafkaConfig
	UserRpc       client.RpcConf // 用户服务地址
	ProductRpc    client.RpcConf // 商品服务地址
	InventoryRpc  client.RpcConf // 库存服务地址
	LogisticsRpc  client.RpcConf // 物流服务地址（发货时建运单）
	PromotionRpc  client.RpcConf // 营销服务地址（创建订单时计算优惠）
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string
	Version string
}
