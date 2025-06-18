package dto

// RegisterRequest 用户注册请求体
// Password 将在服务层进行加密存储
// RealName / Email 等可选
// binding 标签用于请求验证
// swagger 注解可稍后补充
//
// swagger:model RegisterRequest
//
// example: {"username":"admin","password":"123456","email":"admin@example.com"}
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	RealName string `json:"real_name"`
}

// UpdateUserRequest 用户信息更新请求体
// 这里仅示例更新 RealName / Phone 等字段
// 真实项目中可能还需要头像等
//
// swagger:model UpdateUserRequest
type UpdateUserRequest struct {
	RealName string `json:"real_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}
