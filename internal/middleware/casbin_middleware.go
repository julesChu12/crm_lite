package middleware

import (
	"crm_lite/internal/core/resource"
	"crm_lite/pkg/resp"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// CasbinMiddleware 结构体，用于封装依赖
type CasbinMiddleware struct {
	enforcer     *casbin.Enforcer
	publicRoutes map[string]struct{}
}

// NewCasbinMiddleware 创建一个新的 CasbinMiddleware 实例
func NewCasbinMiddleware(rm *resource.Manager) *CasbinMiddleware {
	// 获取 Casbin Enforcer
	casbinRes, err := resource.Get[*resource.CasbinResource](rm, resource.CasbinServiceKey)
	if err != nil {
		// 在中间件创建时，如果依赖不可用，直接 panic
		panic(fmt.Sprintf("failed to get casbin resource for middleware: %v", err))
	}

	// 加载白名单路由 - 这段代码现在可以移除，因为JWT中间件会处理
	// publics := policy.GetPublicRoutes()
	// publicRoutesMap := make(map[string]struct{}, len(publics))
	// for _, r := range publics {
	// 	key := r.Method + ":" + r.Path
	// 	publicRoutesMap[key] = struct{}{}
	// }

	return &CasbinMiddleware{
		enforcer: casbinRes.GetEnforcer(),
		// publicRoutes: publicRoutesMap, // 移除
	}
}

// Check 是 Gin 的处理函数，执行权限检查
func (m *CasbinMiddleware) Check(c *gin.Context) {
	// 1. 检查是否为上游中间件（JWT）标记的公开路由
	if isPublic, exists := c.Get(CtxKeyIsPublic); exists && isPublic.(bool) {
		c.Next()
		return
	}

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

	// 3. 获取请求的 Path 和 Method
	obj := c.Request.URL.Path
	act := c.Request.Method

	// 4. 遍历用户的所有角色，检查权限
	var passed bool
	for _, sub := range roles {
		// 使用 Casbin 进行权限检查
		allowed, err := m.enforcer.Enforce(sub, obj, act)
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
