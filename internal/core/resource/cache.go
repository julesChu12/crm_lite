package resource

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
)

// CacheResource 封装了go-redis客户端
type CacheResource struct {
	*redis.Client
	opts config.CacheOptions
}

// NewCacheResource 创建一个新的缓存资源实例
func NewCacheResource(opts config.CacheOptions) *CacheResource {
	return &CacheResource{opts: opts}
}

// Initialize 实现了Resource接口，用于初始化Redis连接
func (c *CacheResource) Initialize(ctx context.Context) error {
	if c.opts.Driver != "redis" {
		logger.GetGlobalLogger().SugaredLogger.Warnf("Cache driver is '%s', skipping Redis initialization.", c.opts.Driver)
		return nil
	}

	redisOpts := &redis.Options{
		Addr:            fmt.Sprintf("%s:%d", c.opts.Redis.Host, c.opts.Redis.Port),
		Password:        c.opts.Redis.Password,
		DB:              c.opts.Redis.DB,
		PoolSize:        c.opts.Redis.PoolSize,
		MinIdleConns:    c.opts.Redis.MinIdleConns,
		ConnMaxLifetime: c.opts.Redis.MaxConnAge,
		ConnMaxIdleTime: c.opts.Redis.IdleTimeout,
		DialTimeout:     c.opts.Redis.DialTimeout,
		ReadTimeout:     c.opts.Redis.ReadTimeout,
		WriteTimeout:    c.opts.Redis.WriteTimeout,
	}

	c.Client = redis.NewClient(redisOpts)

	// 测试连接
	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.GetGlobalLogger().SugaredLogger.Info("Cache resource (Redis) initialized successfully.")
	return nil
}

// Close 实现了Resource接口，用于关闭Redis连接
func (c *CacheResource) Close(ctx context.Context) error {
	if c.Client == nil {
		return nil
	}

	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("failed to close redis connection: %w", err)
	}

	logger.GetGlobalLogger().SugaredLogger.Info("Cache resource (Redis) closed successfully.")
	return nil
}
