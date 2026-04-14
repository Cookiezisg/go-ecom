package job

import (
	"log"

	"gorm.io/gorm"

	"ecommerce-system/internal/pkg/client"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/service/job/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config     Config
	DB         *gorm.DB
	InvClient  *client.InventoryClient
	OrderRepo  repository.OrderRepository
	CouponRepo repository.CouponRepository
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

	ctx := &ServiceContext{
		Config:     c,
		DB:         db,
		OrderRepo:  repository.NewOrderRepository(db),
		CouponRepo: repository.NewCouponRepository(db),
	}

	// 库存服务客户端（用于取消超时订单时解锁预占库存，可选）
	if c.InventoryRpc.Endpoint != "" {
		ic, err := client.NewInventoryClient(c.InventoryRpc)
		if err != nil {
			log.Fatalf("初始化库存服务客户端失败: %v", err)
		}
		ctx.InvClient = ic
	}

	return ctx
}
