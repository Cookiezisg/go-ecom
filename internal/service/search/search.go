package search

import (
	"context"
	"log"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/mq"
	pkgsearch "ecommerce-system/internal/pkg/search"
	"ecommerce-system/internal/service/search/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config       Config
	Redis        *redis.Client
	ESClient     *pkgsearch.Client
	DB           *gorm.DB
	MQConsumer   *mq.Consumer
	SearchRepo   repository.SearchRepository
	SnapshotRepo repository.ProductSnapshotRepository
}

// NewServiceContext 创建服务上下文。Redis/DB 初始化失败直接 Fatal，不静默放行。
// ES 和 Kafka 消费者为可选依赖（搜索服务在没有 ES 时降级）。
func NewServiceContext(c Config) *ServiceContext {
	rdb := cache.MustNewRedis(&cache.Config{
		Host:         c.BizRedis.Host,
		Port:         c.BizRedis.Port,
		Password:     c.BizRedis.Password,
		Database:     c.BizRedis.Database,
		PoolSize:     c.BizRedis.PoolSize,
		MinIdleConns: c.BizRedis.MinIdleConns,
	})

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

	ctx := &ServiceContext{
		Config:       c,
		Redis:        rdb,
		DB:           db,
		SnapshotRepo: repository.NewProductSnapshotRepository(db),
	}

	// Elasticsearch 可选（无 ES 时搜索降级为 MySQL 全文检索）
	if len(c.Elasticsearch.Addresses) > 0 {
		esClient, err := pkgsearch.NewElasticsearchClient(&pkgsearch.Config{
			Addresses: c.Elasticsearch.Addresses,
			Username:  c.Elasticsearch.Username,
			Password:  c.Elasticsearch.Password,
		})
		if err != nil {
			log.Printf("警告：初始化Elasticsearch客户端失败: %v", err)
		} else {
			ctx.ESClient = esClient
			_ = esClient.CreateIndex(context.Background(), repository.ProductIndexName, repository.ProductIndexMapping)
		}
	}

	ctx.SearchRepo = repository.NewSearchRepository(rdb, ctx.ESClient)

	// Kafka 消费者可选（用于监听商品数据变更事件，同步到 ES）
	if len(c.Kafka.Brokers) > 0 && c.Kafka.ConsumerGroup != "" {
		consumer, err := mq.NewConsumer(&mq.Config{
			Brokers:       c.Kafka.Brokers,
			Version:       c.Kafka.Version,
			ConsumerGroup: c.Kafka.ConsumerGroup,
		})
		if err != nil {
			log.Printf("警告：初始化Kafka消费者失败: %v", err)
		} else {
			ctx.MQConsumer = consumer
			ctx.MQConsumer.RegisterHandler(mq.TopicDataSync, func(cctx context.Context, msg *mq.Message) error {
				return ctx.handleDataSyncMessage(cctx, msg)
			})
			go func() {
				_ = ctx.MQConsumer.Start(context.Background(), []string{mq.TopicDataSync})
			}()
		}
	}

	return ctx
}
