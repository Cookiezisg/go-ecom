package seckill

import (
	"log"

	v1 "ecommerce-system/api/seckill/v1"
	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/service/seckill/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config              Config
	DB                  *gorm.DB
	Redis               *redis.Client
	Cache               *cache.CacheOperations
	MQProducer          *mq.Producer
	SeckillActivityRepo repository.SeckillActivityRepository
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
		Config:              c,
		DB:                  db,
		Redis:               rdb,
		Cache:               cache.NewCacheOperations(rdb),
		SeckillActivityRepo: repository.NewSeckillActivityRepository(db),
	}

	// Kafka 生产者可选（不影响秒杀主逻辑）
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

// SeckillService 秒杀服务
type SeckillService struct {
	v1.UnimplementedSeckillServiceServer
	svcCtx *ServiceContext
}

// NewSeckillService 创建秒杀服务
func NewSeckillService(svcCtx *ServiceContext) *SeckillService {
	return &SeckillService{svcCtx: svcCtx}
}
