package middleware

import (
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/policy"
	"crm_lite/pkg/resp"
	"crm_lite/pkg/utils"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// CtxKeyIsPublic 在 gin.Context 中标记当前请求是否属于公开路由。
const CtxKeyIsPublic = "is_public"

// NewJWTAuthMiddleware 返回一个 Gin 中间件，用于完成 JWT 认证。
// 使用示例：
//
//	router.Use(middleware.NewJWTAuthMiddleware(resManager))
func NewJWTAuthMiddleware(resManager *resource.Manager) gin.HandlerFunc {
	// 1. 预计算公开路由 Map（白名单）
	publics := policy.GetPublicRoutes()
	publicRoutes := make(map[string]struct{}, len(publics))
	for _, r := range publics {
		publicRoutes[r.Method+":"+r.Path] = struct{}{}
	}

	// 2. 返回真正的处理中间件
	return func(c *gin.Context) {
		// 2.1 判断是否为公开路由
		routeKey := c.Request.Method + ":" + c.Request.URL.Path
		if _, ok := publicRoutes[routeKey]; ok {
			c.Set(CtxKeyIsPublic, true)
			c.Next()
			return
		}
		c.Set(CtxKeyIsPublic, false)

		// 2.2 解析 Authorization 头
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

		// 2.3 解析并校验 token
		// 使用解析出的 Bearer token 进行校验，并读取全局配置中的 JWT 设置
		claims, err := utils.ParseToken(parts[1], config.GetInstance().Auth.JWTOptions)
		if err != nil {
			resp.Error(c, resp.CodeUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// 2.4 检查黑名单（登出后失效等场景）
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

		// 2.5 将用户信息注入上下文，供后续业务使用
		c.Set(ContextKeyClaims, claims)
		ctx := utils.WithUser(c.Request.Context(), claims.UserID, claims.Username)
		c.Request = c.Request.WithContext(ctx)
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRoles, claims.Roles)

		c.Next()
	}
}
