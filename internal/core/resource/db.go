package resource

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
)

// DBResource 封装了GORM数据库连接
type DBResource struct {
	*gorm.DB
	opts config.DBOptions
}

// NewDBResource 创建一个新的数据库资源实例
func NewDBResource(opts config.DBOptions) *DBResource {
	return &DBResource{opts: opts}
}

// Initialize 实现了Resource接口，用于初始化数据库连接
func (d *DBResource) Initialize(ctx context.Context) error {
	var dialector gorm.Dialector
	var err error

	dsn := d.buildDSN()

	switch d.opts.Driver {
	case "postgres":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", d.opts.Driver)
	}

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.NewGormLoggerAdapter().LogMode(gormlogger.Silent),
	}

	// 根据环境设置日志级别
	if d.opts.Driver != "postgres" { // 简单的判断，例如在非生产环境开启详细日志
		gormConfig.Logger = logger.NewGormLoggerAdapter().LogMode(gormlogger.Info)
	}

	// 建立连接
	d.DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(int(d.opts.MaxOpenConns))
	sqlDB.SetMaxIdleConns(int(d.opts.MaxIdleConns))
	sqlDB.SetConnMaxLifetime(d.opts.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.GetGlobalLogger().SugaredLogger.Infof("Database resource initialized successfully for driver: %s", d.opts.Driver)
	return nil
}

// Close 实现了Resource接口，用于关闭数据库连接
func (d *DBResource) Close(ctx context.Context) error {
	if d.DB == nil {
		return nil
	}
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB for closing: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	logger.GetGlobalLogger().SugaredLogger.Info("Database resource closed successfully.")
	return nil
}

// buildDSN 根据配置构建数据库连接字符串
func (d *DBResource) buildDSN() string {
	switch d.opts.Driver {
	case "postgres":
		// e.g. "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			d.opts.Host, d.opts.User, d.opts.Password, d.opts.DBName, d.opts.Port, d.opts.SSLMode, d.opts.TimeZone)
	case "mysql":
		// e.g. "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.opts.User, d.opts.Password, d.opts.Host, d.opts.Port, d.opts.DBName)
	default:
		return ""
	}
}
