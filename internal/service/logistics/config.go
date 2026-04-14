package logistics

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 物流服务配置
type Config struct {
	zrpc.RpcServerConf
	Database DatabaseConfig
	BizRedis RedisConfig // 业务侧 Redis，用于 idgen 物流单号生成
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
