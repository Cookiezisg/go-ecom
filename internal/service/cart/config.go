package cart

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Database DatabaseConfig
	BizRedis RedisConfig
	JWT      JWTConfig
}

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	User            string
	Password        string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
}

type JWTConfig struct {
	Secret string
	Expire int64
}
