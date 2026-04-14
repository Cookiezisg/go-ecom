package logistics

import (
	"gorm.io/gorm"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/pkg/idgen"
	"ecommerce-system/internal/service/logistics/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config        Config
	DB            *gorm.DB
	IDGen         *idgen.Generator
	LogisticsRepo repository.LogisticsRepository
}

// NewServiceContext 创建服务上下文。DB 初始化失败直接 Fatal，不静默放行。
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

	var ig *idgen.Generator
	if c.BizRedis.Host != "" {
		rdb := cache.MustNewRedis(&cache.Config{
			Host:         c.BizRedis.Host,
			Port:         c.BizRedis.Port,
			Password:     c.BizRedis.Password,
			Database:     c.BizRedis.Database,
			PoolSize:     c.BizRedis.PoolSize,
			MinIdleConns: c.BizRedis.MinIdleConns,
		})
		ig = idgen.New(rdb)
	}

	return &ServiceContext{
		Config:        c,
		DB:            db,
		IDGen:         ig,
		LogisticsRepo: repository.NewLogisticsRepository(db),
	}
}
