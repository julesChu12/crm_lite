package resource

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"crm_lite/internal/core/config"
)

// TestCacheResource_Initialize 使用 miniredis 验证 CacheResource 能正常连接、读写、关闭。
func TestCacheResource_Initialize(t *testing.T) {
	ctx := context.Background()

	// 启动内存 Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	host, portStr, _ := strings.Cut(mr.Addr(), ":")
	port, _ := strconv.Atoi(portStr)

	opts := config.CacheOptions{
		Driver: "redis",
		Redis: config.RedisConfig{
			Host: host,
			Port: port,
		},
	}

	cacheRes := NewCacheResource(opts)
	require.NoError(t, cacheRes.Initialize(ctx))
	defer cacheRes.Close(ctx)

	// 写入并读取
	err = cacheRes.Client.Set(ctx, "foo", "bar", 0).Err()
	require.NoError(t, err)

	val, err := cacheRes.Client.Get(ctx, "foo").Result()
	require.NoError(t, err)
	require.Equal(t, "bar", val)
}
