package payment

import (
	"log"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/client"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/idgen"
	"ecommerce-system/internal/service/payment/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config         Config
	DB             *gorm.DB
	Redis          *redis.Client
	Cache          *cache.CacheOperations
	IDGen          *idgen.Generator
	PaymentRepo    repository.PaymentRepository
	PaymentLogRepo repository.PaymentLogRepository
	OrderClient    *client.OrderClient
	InvClient      *client.InventoryClient
}

// NewServiceContext 创建服务上下文。DB/Redis 初始化失败直接 Fatal，不静默放行。
func NewServiceContext(c Config) *ServiceContext {
	db := database.MustNewMySQL(&database.Config{
		Host:            c.Database.Host,
		Port:            c.Database.Port,
		User:            c.Database.User,
		Password:        c.Database.Password,
		Database:        c.Database.Database,
		Charset:         c.Database.Charset,
		MaxOpenConns:    c.Database.MaxOpenConns,
		MaxIdleConns:    c.Database.MaxIdleConns,
		ConnMaxLifetime: c.Database.ConnMaxLifetime,
		ConnMaxIdleTime: c.Database.ConnMaxIdleTime,
	})

	rdb := cache.MustNewRedis(&cache.Config{
		Host:         c.BizRedis.Host,
		Port:         c.BizRedis.Port,
		Password:     c.BizRedis.Password,
		Database:     c.BizRedis.Database,
		PoolSize:     c.BizRedis.PoolSize,
		MinIdleConns: c.BizRedis.MinIdleConns,
	})

	ctx := &ServiceContext{
		Config:         c,
		DB:             db,
		Redis:          rdb,
		Cache:          cache.NewCacheOperations(rdb),
		IDGen:          idgen.New(rdb),
		PaymentRepo:    repository.NewPaymentRepository(db),
		PaymentLogRepo: repository.NewPaymentLogRepository(db),
	}

	if c.OrderRpc.Endpoint != "" {
		oc, err := client.NewOrderClient(c.OrderRpc)
		if err != nil {
			log.Fatalf("初始化订单服务客户端失败: %v", err)
		}
		ctx.OrderClient = oc
	}

	if c.InventoryRpc.Endpoint != "" {
		ic, err := client.NewInventoryClient(c.InventoryRpc)
		if err != nil {
			log.Fatalf("初始化库存服务客户端失败: %v", err)
		}
		ctx.InvClient = ic
	}

	return ctx
}
