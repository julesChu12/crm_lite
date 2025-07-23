package middleware

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NewSimpleCustomerAccessMiddleware 返回一个基于 manager_id 简单层级的权限校验中间件。
// 必须在 JWTAuthMiddleware 之后调用，才能从上下文中获取 user_id、roles。
func NewSimpleCustomerAccessMiddleware(resManager *resource.Manager) gin.HandlerFunc {
	hierarchySvc := service.NewSimpleHierarchyService(resManager)

	return func(c *gin.Context) {
		// 仅拦截涉及 customer_id 的路由
		customerIDStr := c.Param("id")
		if customerIDStr == "" {
			c.Next()
			return
		}

		customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err != nil {
			resp.Error(c, http.StatusBadRequest, "无效的客户ID")
			c.Abort()
			return
		}

		userIDVal, exists := c.Get("user_id")
		if !exists {
			resp.Error(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		userID := userIDVal.(int64)

		rolesVal, _ := c.Get("roles")
		var roles []string
		if rolesVal != nil {
			roles = rolesVal.([]string)
		}

		// admin 角色放行
		for _, r := range roles {
			if r == "admin" {
				c.Next()
				return
			}
		}

		// 如果是负责人本人或上级，放行
		allowed, err := hierarchySvc.CanAccessCustomer(c.Request.Context(), userID, customerID)
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, "权限检查失败")
			c.Abort()
			return
		}
		if !allowed {
			resp.Error(c, http.StatusForbidden, "无权访问该客户")
			c.Abort()
			return
		}

		c.Next()
	}
}
