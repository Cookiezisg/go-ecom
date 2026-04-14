package cart

import (
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/client"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/service/cart/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config        Config
	DB            *gorm.DB
	Redis         *redis.Client
	ProductClient *client.ProductClient
	InvClient     *client.InventoryClient
	CartRepo      repository.CartRepository
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
		Config:   c,
		DB:       db,
		Redis:    rdb,
		CartRepo: repository.NewCartRepository(db, rdb),
	}

	if c.ProductRpc.Endpoint != "" {
		pc, err := client.NewProductClient(c.ProductRpc)
		if err != nil {
			log.Fatalf("初始化商品服务客户端失败: %v", err)
		}
		ctx.ProductClient = pc
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
