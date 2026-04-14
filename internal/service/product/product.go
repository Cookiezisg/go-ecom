package product

import (
	"context"
	"log"

	v1 "ecommerce-system/api/product/v1"
	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/pkg/outbox"
	"ecommerce-system/internal/service/product/repository"
	"ecommerce-system/internal/service/product/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config       Config
	DB           *gorm.DB
	Redis        *redis.Client
	Cache        *cache.CacheOperations
	MQProducer   *mq.Producer
	OutboxRepo   *outbox.Repo
	ProductRepo  repository.ProductRepository
	CategoryRepo repository.CategoryRepository
	SkuRepo      repository.SkuRepository
	BannerRepo   repository.BannerRepository
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
		Config:       c,
		DB:           db,
		Redis:        rdb,
		Cache:        cache.NewCacheOperations(rdb),
		ProductRepo:  repository.NewProductRepository(db),
		CategoryRepo: repository.NewCategoryRepository(db),
		SkuRepo:      repository.NewSkuRepository(db),
		BannerRepo:   repository.NewBannerRepository(db),
		OutboxRepo:   outbox.NewRepo(db),
	}

	// Kafka 生产者可选（不影响主链路，仅用于 outbox relay）
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

	// 启动 Outbox Relay：把 outbox_event 异步投递到 Kafka（用于 ES 数据同步）
	if ctx.OutboxRepo != nil && ctx.MQProducer != nil {
		relay := outbox.NewRelay(ctx.OutboxRepo, ctx.MQProducer, outbox.RelayConfig{})
		go relay.Start(context.Background())
	}

	return ctx
}

// ProductService 商品服务
type ProductService struct {
	v1.UnimplementedProductServiceServer
	svcCtx *ServiceContext
	logic  *service.ProductLogic
}

// NewProductService 创建商品服务
func NewProductService(svcCtx *ServiceContext) *ProductService {
	return &ProductService{
		svcCtx: svcCtx,
		logic: service.NewProductLogic(
			svcCtx.DB,
			svcCtx.OutboxRepo,
			svcCtx.ProductRepo,
			svcCtx.CategoryRepo,
			svcCtx.SkuRepo,
			svcCtx.BannerRepo,
			svcCtx.Cache,
			svcCtx.MQProducer,
		),
	}
}
