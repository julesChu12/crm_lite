package middleware

import (
	"crm_lite/internal/core/resource"
	"crm_lite/pkg/resp"
	"fmt"

	"github.com/gin-gonic/gin"
)

// CasbinMiddleware 基于 Casbin 的权限访问控制中间件
func CasbinMiddleware(rm *resource.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Casbin Enforcer
		casbinRes, err := resource.Get[*resource.CasbinResource](rm, resource.CasbinServiceKey)
		if err != nil {
			resp.Error(c, resp.CodeInternalError, "casbin service not available")
			c.Abort()
			return
		}
		enforcer := casbinRes.GetEnforcer()

		// 2. 从 gin.Context 中获取用户角色
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

		// 3. 获取请求的 Path (obj) 和 Method (act)
		obj := c.Request.URL.Path
		act := c.Request.Method

		// 4. 遍历用户的所有角色，检查权限
		var passed bool
		for _, sub := range roles {
			// 使用 Casbin 进行权限检查
			allowed, err := enforcer.Enforce(sub, obj, act)
			if err != nil {
				// 如果 Casbin 在检查时出错，记录日志并拒绝访问
				fmt.Printf("casbin enforce error: %v\n", err)
				resp.Error(c, resp.CodeInternalError, "casbin authorization failed")
				c.Abort()
				return
			}
			if allowed {
				passed = true
				break // 任何一个角色有权限即可通过
			}
		}

		// 5. 如果所有角色都没有权限，则拒绝访问
		if !passed {
			resp.Error(c, resp.CodeForbidden, fmt.Sprintf("access denied: no permission for action '%s' on resource '%s'", act, obj))
			c.Abort()
			return
		}

		// 权限检查通过
		c.Next()
	}
}
