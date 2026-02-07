package database

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // 秒
	ConnMaxIdleTime int // 秒
}

// NewMySQL 创建MySQL连接
func NewMySQL(cfg *Config) (*gorm.DB, error) {
	var dsn string

	// 如果 Host 是 localhost 或 127.0.0.1，尝试使用 Unix socket
	// 这样可以兼容 skip_networking=ON 的情况（MySQL 只监听 socket）
	// 但如果 socket 不存在，自动回退到 TCP 连接（适用于生产环境）
	if cfg.Host == "localhost" || cfg.Host == "127.0.0.1" || cfg.Host == "::1" {
		// 尝试常见的 socket 路径
		socketPaths := []string{
			"/tmp/mysql.sock",
			"/var/run/mysqld/mysqld.sock",
			"/var/lib/mysql/mysql.sock",
			"/usr/local/var/mysql/mysql.sock",
		}

		var socketPath string
		for _, path := range socketPaths {
			if _, err := os.Stat(path); err == nil {
				socketPath = path
				break
			}
		}

		if socketPath != "" {
			// 找到 socket 文件，使用 socket 连接
			dsn = fmt.Sprintf("%s:%s@unix(%s)/%s?charset=%s&parseTime=True&loc=Local",
				cfg.User,
				cfg.Password,
				socketPath,
				cfg.Database,
				cfg.Charset,
			)
		} else {
			// Socket 不存在，使用 TCP 连接（生产环境通常是这样）
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
				cfg.User,
				cfg.Password,
				cfg.Host,
				cfg.Port,
				cfg.Database,
				cfg.Charset,
			)
		}
	} else {
		// 非本地地址，使用 TCP 连接
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
			cfg.Charset,
		)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return db, nil
}
