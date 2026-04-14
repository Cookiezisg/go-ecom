package order

import (
	"log"

	v1 "ecommerce-system/api/order/v1"
	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/service/order/repository"
	"ecommerce-system/internal/service/order/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config        Config
	DB            *gorm.DB
	Redis         *redis.Client
	Cache         *cache.CacheOperations
	MQProducer    *mq.Producer
	OrderRepo     repository.OrderRepository
	OrderItemRepo repository.OrderItemRepository
	OrderLogRepo  repository.OrderLogRepository
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
		Config:        c,
		DB:            db,
		Redis:         rdb,
		Cache:         cache.NewCacheOperations(rdb),
		OrderRepo:     repository.NewOrderRepository(db),
		OrderItemRepo: repository.NewOrderItemRepository(db),
		OrderLogRepo:  repository.NewOrderLogRepository(db),
	}

	// Kafka 生产者可选（不影响主链路）
	if len(c.Kafka.Brokers) > 0 {
		mqProducer, err := mq.NewProducer(&mq.Config{
			Brokers:       c.Kafka.Brokers,
			Version:       c.Kafka.Version,
			ProducerAsync: true,
		})
		if err != nil {
			log.Printf("警告：初始化Kafka生产者失败: %v", err)
		} else {
			ctx.MQProducer = mqProducer
		}
	}

	return ctx
}

// OrderService 订单服务
type OrderService struct {
	v1.UnimplementedOrderServiceServer
	svcCtx *ServiceContext
	logic  *service.OrderLogic
}

// NewOrderService 创建订单服务
func NewOrderService(svcCtx *ServiceContext) *OrderService {
	return &OrderService{
		svcCtx: svcCtx,
		logic: service.NewOrderLogic(
			svcCtx.OrderRepo,
			svcCtx.OrderItemRepo,
			svcCtx.OrderLogRepo,
			svcCtx.Cache,
			svcCtx.MQProducer,
		),
	}
}
