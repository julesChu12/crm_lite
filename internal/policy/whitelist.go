package policy

// PublicRoute 定义了一个公开访问的路由，它不需要身份验证或授权。
type PublicRoute struct {
	Method string
	Path   string
}

// GetPublicRoutes 返回应用程序中所有公开路由的列表。
// 这些路由将从 Casbin 策略生成中被排除。
func GetPublicRoutes() []PublicRoute {
	return []PublicRoute{
		// 认证相关端点
		{Method: "POST", Path: "/api/v1/auth/login"},
		{Method: "POST", Path: "/api/v1/auth/register"},
		{Method: "POST", Path: "/api/v1/auth/refresh"},
		{Method: "POST", Path: "/api/v1/auth/forgot-password"},
		{Method: "POST", Path: "/api/v1/auth/reset-password"},
	}
}
