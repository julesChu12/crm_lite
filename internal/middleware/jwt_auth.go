package middleware

import (
	"crm_lite/internal/core/resource"
	"crm_lite/pkg/resp"
	"crm_lite/pkg/utils"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 验证 JWT 并将用户信息注入到上下文
func JWTAuthMiddleware(resManager *resource.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp.Error(c, resp.CodeUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			resp.Error(c, resp.CodeUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			resp.Error(c, resp.CodeUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// 检查黑名单
		cache, err := resource.Get[*resource.CacheResource](resManager, resource.CacheServiceKey)
		if err != nil {
			resp.SystemError(c, fmt.Errorf("failed to get cache resource for auth: %w", err))
			c.Abort()
			return
		}
		isBlacklisted, _ := cache.Client.Exists(c.Request.Context(), "jti:"+claims.ID).Result()
		if isBlacklisted > 0 {
			resp.Error(c, resp.CodeUnauthorized, "token has been invalidated")
			c.Abort()
			return
		}

		// 将 claims 注入到 context，便于后续 handler（如 Logout）使用
		c.Set("claims", claims)

		// 将用户信息注入到 context
		ctx := utils.WithUser(c.Request.Context(), claims.UserID, claims.Username)
		c.Request = c.Request.WithContext(ctx)

		// 同时将用户信息存入 gin.Context，便于在 handler 中直接使用
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}
