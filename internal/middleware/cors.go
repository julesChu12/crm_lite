package middleware

import (
	"crm_lite/internal/core/config"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewCorsMiddleware 创建CORS中间件
func NewCorsMiddleware() gin.HandlerFunc {
	corsConfig := config.GetInstance().CORS

	// 如果配置为空，提供开发环境默认允许的本地端口列表
	if len(corsConfig.AllowOrigins) == 0 || (len(corsConfig.AllowOrigins) == 1 && corsConfig.AllowOrigins[0] == "*") {
		corsConfig.AllowOrigins = []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"http://localhost:8081",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:8081",
		}
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查Origin是否在允许列表中
		if origin != "" && isOriginAllowed(origin, corsConfig.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// 设置其他CORS头部
		c.Header("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowHeaders, ", "))
		c.Header("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposeHeaders, ", "))

		if corsConfig.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if corsConfig.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", int(corsConfig.MaxAge.Seconds())))
		}

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// isOriginAllowed 检查Origin是否被允许
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// 支持通配符匹配（简单实现）
		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}
	}
	return false
}
