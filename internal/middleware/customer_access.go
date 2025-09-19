package middleware

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/identity/impl"
	"crm_lite/pkg/resp"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewSimpleCustomerAccessMiddleware 返回一个基于 manager_id 简单层级的权限校验中间件。
// 必须在 JWTAuthMiddleware 之后调用，才能从上下文中获取 user_id、roles。
func NewSimpleCustomerAccessMiddleware(resManager *resource.Manager) gin.HandlerFunc {
	hierarchySvc := impl.NewHierarchyService(resManager)

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

		userIDVal, exists := c.Get(ContextKeyUserID)
		if !exists {
			resp.Error(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}

		rolesVal, _ := c.Get(ContextKeyRoles)
		var roles []string
		if rolesVal != nil {
			roles = rolesVal.([]string)
		}
		fmt.Println("roles", roles)
		// admin 角色放行
		for _, r := range roles {
			if r == "super_admin" {
				c.Next()
				return
			}
		}
		userUUID := userIDVal.(string)
		// 根据UUID查询用户ID
		dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, "数据库资源获取失败")
			c.Abort()
			return
		}
		q := query.Use(dbRes.DB)
		user, err := q.AdminUser.WithContext(c.Request.Context()).Where(q.AdminUser.UUID.Eq(userUUID)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				resp.Error(c, http.StatusUnauthorized, "用户不存在")
			} else {
				resp.Error(c, http.StatusInternalServerError, "用户查询失败")
			}
			c.Abort()
			return
		}
		userID := user.ID
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
