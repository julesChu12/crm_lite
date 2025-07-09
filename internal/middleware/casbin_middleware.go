package middleware

import (
	"crm_lite/internal/core/resource"
	"crm_lite/pkg/resp"
	"fmt"

	"github.com/gin-gonic/gin"
)

// NewCasbinMiddleware 创建一个 Casbin 授权中间件（gin.HandlerFunc）。
// 调用方式示例：
//
//	router.Use(middleware.NewCasbinMiddleware(resManager))
func NewCasbinMiddleware(rm *resource.Manager) gin.HandlerFunc {
	// 1. 获取 Casbin Enforcer 实例
	casbinRes, err := resource.Get[*resource.CasbinResource](rm, resource.CasbinServiceKey)
	if err != nil {
		panic(fmt.Sprintf("failed to get casbin resource for middleware: %v", err))
	}
	enforcer := casbinRes.GetEnforcer()

	// 2. 返回具体的 gin.HandlerFunc
	return func(c *gin.Context) {
		// 2.1 公开路由直接放行（上游 JWT 中间件已设置标记）
		if isPublic, exists := c.Get(CtxKeyIsPublic); exists && isPublic.(bool) {
			c.Next()
			return
		}

		// 2.2 从 context 中获取用户角色
		rolesVal, exists := c.Get("roles")
		if !exists {
			resp.Error(c, resp.CodeForbidden, "access denied: user roles not found in context")
			c.Abort()
			return
		}
		roles, ok := rolesVal.([]string)
		if !ok || len(roles) == 0 {
			resp.Error(c, resp.CodeForbidden, "access denied: invalid user roles")
			c.Abort()
			return
		}

		// 2.3 获取请求的对象和动作
		obj := c.Request.URL.Path
		act := c.Request.Method

		// 2.4 遍历角色做权限校验
		for _, sub := range roles {
			allowed, err := enforcer.Enforce(sub, obj, act)
			if err != nil {
				resp.Error(c, resp.CodeInternalError, "casbin authorization failed")
				c.Abort()
				return
			}
			if allowed {
				c.Next()
				return
			}
		}

		// 2.5 若所有角色都无权限
		resp.Error(c, resp.CodeForbidden, fmt.Sprintf("access denied: no permission for action '%s' on resource '%s'", act, obj))
		c.Abort()
	}
}
