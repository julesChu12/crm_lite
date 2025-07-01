package middleware

import (
	"bytes"
	"io"
	"time"

	"crm_lite/internal/core/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinLogger 返回一个Gin中间件，该中间件使用Zap记录请求日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 读取请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// 重要：重置请求体，使其可以被后续的处理器重新读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 处理请求
		c.Next()

		// 请求处理完毕后记录日志
		cost := time.Since(start)

		// 直接使用 Raw logger，它的调用深度对中间件场景是正确的
		middlewareLogger := logger.GetGlobalLogger().Raw()

		middlewareLogger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("params", string(bodyBytes)),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}
