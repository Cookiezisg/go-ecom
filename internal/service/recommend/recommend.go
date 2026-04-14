package recommend

import (
	"github.com/redis/go-redis/v9"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/service/recommend/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config        Config
	Redis         *redis.Client
	RecommendRepo repository.RecommendRepository
}

// NewServiceContext 创建服务上下文。Redis 初始化失败直接 Fatal，不静默放行。
func NewServiceContext(c Config) *ServiceContext {
	rdb := cache.MustNewRedis(&cache.Config{
		Host:         c.BizRedis.Host,
		Port:         c.BizRedis.Port,
		Password:     c.BizRedis.Password,
		Database:     c.BizRedis.Database,
		PoolSize:     c.BizRedis.PoolSize,
		MinIdleConns: c.BizRedis.MinIdleConns,
	})

	return &ServiceContext{
		Config:        c,
		Redis:         rdb,
		RecommendRepo: repository.NewRecommendRepository(rdb),
	}
}
