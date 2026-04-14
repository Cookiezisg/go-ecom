package user

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	v1 "ecommerce-system/api/user/v1"
	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/database"
	"ecommerce-system/internal/service/user/repository"
	userservice "ecommerce-system/internal/service/user/service"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config         Config
	DB             *gorm.DB
	Redis          *redis.Client
	Cache          *cache.CacheOperations
	UserRepo       repository.UserRepository
	CredentialRepo repository.CredentialRepository
	AddressRepo    repository.AddressRepository
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

	return &ServiceContext{
		Config:         c,
		DB:             db,
		Redis:          rdb,
		Cache:          cache.NewCacheOperations(rdb),
		UserRepo:       repository.NewUserRepository(db),
		CredentialRepo: repository.NewCredentialRepository(db),
		AddressRepo:    repository.NewAddressRepository(db),
	}
}

// UserService 用户服务
type UserService struct {
	v1.UnimplementedUserServiceServer
	svcCtx       *ServiceContext
	logic        *userservice.UserLogic
	addressLogic *userservice.AddressLogic
}

// NewUserService 创建用户服务
func NewUserService(svcCtx *ServiceContext) *UserService {
	return &UserService{
		svcCtx:       svcCtx,
		logic:        userservice.NewUserLogic(svcCtx.UserRepo, svcCtx.CredentialRepo, svcCtx.AddressRepo, svcCtx.Cache),
		addressLogic: userservice.NewAddressLogic(svcCtx.AddressRepo, svcCtx.Cache),
	}
}
