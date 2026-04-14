package review

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/mongodb"
	"ecommerce-system/internal/service/review/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config          Config
	DB              *gorm.DB
	MongoDB         *mongodb.Client
	ReviewRepo      repository.ReviewRepository
	ReviewReplyRepo repository.ReviewReplyRepository
}

// NewServiceContext 创建服务上下文。DB 初始化失败直接 Fatal，不静默放行。
// MongoDB 为可选依赖（无 MongoDB 时评价图片等富媒体字段不可用）。
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

	ctx := &ServiceContext{
		Config:          c,
		DB:              db,
		ReviewRepo:      repository.NewReviewRepository(db, nil),
		ReviewReplyRepo: repository.NewReviewReplyRepository(db),
	}

	// MongoDB 可选（存储评价富媒体内容）
	if c.MongoDB != nil && c.MongoDB.URI != "" && c.MongoDB.Database != "" {
		mongoClient, err := mongodb.NewMongoDBClient(&mongodb.Config{
			URI:            c.MongoDB.URI,
			Database:       c.MongoDB.Database,
			MaxPoolSize:    100,
			MinPoolSize:    10,
			ConnectTimeout: 10,
		})
		if err != nil {
			log.Printf("警告：初始化MongoDB客户端失败: %v", err)
		} else {
			ctx.MongoDB = mongoClient
			_ = initReviewIndexes(context.Background(), mongoClient)
			// 重新初始化 ReviewRepo，带上 MongoDB
			ctx.ReviewRepo = repository.NewReviewRepository(db, mongoClient)
		}
	}

	return ctx
}

// initReviewIndexes 初始化评价集合索引
func initReviewIndexes(ctx context.Context, mongoClient *mongodb.Client) error {
	indexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"product_id": 1, "status": 1, "created_at": -1}},
		{Keys: map[string]interface{}{"user_id": 1, "created_at": -1}},
		{Keys: map[string]interface{}{"rating": 1, "created_at": -1}},
	}
	collection := mongoClient.Collection("reviews")
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}
