package logistics

import (
	"gorm.io/gorm"

	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/service/logistics/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config        Config
	DB            *gorm.DB
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

	return &ServiceContext{
		Config:        c,
		DB:            db,
		LogisticsRepo: repository.NewLogisticsRepository(db),
	}
}
