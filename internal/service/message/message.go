package message

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/service/message/repository"
	"ecommerce-system/internal/service/message/service"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config      Config
	DB          *gorm.DB
	Redis       *redis.Client
	Cache       *cache.CacheOperations
	MessageRepo repository.MessageRepository
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

	msgRepo := repository.NewMessageRepository(db)

	svcCtx := &ServiceContext{
		Config:      c,
		DB:          db,
		Redis:       rdb,
		Cache:       cache.NewCacheOperations(rdb),
		MessageRepo: msgRepo,
	}

	// Kafka 消费者（可选）：监听订单/支付事件并发送站内消息
	if c.Kafka != nil && len(c.Kafka.Brokers) > 0 {
		consumerGroup := c.Kafka.ConsumerGroup
		if consumerGroup == "" {
			consumerGroup = "message-service"
		}
		consumer, err := mq.NewConsumer(&mq.Config{
			Brokers:       c.Kafka.Brokers,
			Version:       c.Kafka.Version,
			ConsumerGroup: consumerGroup,
		})
		if err != nil {
			log.Printf("警告：初始化Kafka消费者失败: %v", err)
		} else {
			logic := service.NewMessageLogic(msgRepo)
			mc := service.NewMessageConsumer(logic)

			consumer.RegisterHandler(mq.TopicOrderCreated, mc.HandleOrderCreated)
			consumer.RegisterHandler(mq.TopicOrderCancelled, mc.HandleOrderCancelled)
			consumer.RegisterHandler(mq.TopicPaymentSuccess, mc.HandlePaymentSuccess)
			consumer.RegisterHandler(mq.TopicPaymentRefunded, mc.HandlePaymentRefunded)

			// 在后台启动消费者
			go func() {
				topics := []string{
					mq.TopicOrderCreated,
					mq.TopicOrderCancelled,
					mq.TopicPaymentSuccess,
					mq.TopicPaymentRefunded,
				}
				if err := consumer.Start(context.Background(), topics); err != nil {
					log.Printf("Kafka消费者退出: %v", err)
				}
			}()
		}
	}

	return svcCtx
}
