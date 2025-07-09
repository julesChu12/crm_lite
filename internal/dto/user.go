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

// GetUserRequest 通过 URI 获取用户信息的请求
type GetUserRequest struct {
	UUID string `uri:"uuid" binding:"required,uuid"`
}

// UserResponse 返回给客户端的用户信息
// 不应包含密码等敏感数据
type UserResponse struct {
	UUID      string   `json:"uuid"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	RealName  string   `json:"real_name"`
	Phone     string   `json:"phone,omitempty"`
	Avatar    string   `json:"avatar,omitempty"`
	IsActive  bool     `json:"is_active"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
}

// --- Admin-facing DTOs for User Management ---

// AdminCreateUserRequest 管理员创建用户请求体
type AdminCreateUserRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required,min=6"`
	Email    string  `json:"email"    binding:"required,email"`
	RealName string  `json:"real_name"`
	Phone    string  `json:"phone"`
	Avatar   string  `json:"avatar"`
	IsActive *bool   `json:"is_active"` // 使用指针以区分 "未提供" 和 "设置为false"
	RoleIDs  []int64 `json:"role_ids"`  // 关联的角色ID列表
}

// AdminUpdateUserRequest 管理员更新用户请求体
type AdminUpdateUserRequest struct {
	Email    string  `json:"email,omitempty"`
	RealName string  `json:"real_name,omitempty"`
	Phone    string  `json:"phone,omitempty"`
	Avatar   string  `json:"avatar,omitempty"`
	IsActive *bool   `json:"is_active"`
	RoleIDs  []int64 `json:"role_ids"`
}

// UserListRequest 查询用户列表的请求参数
type UserListRequest struct {
	Username string `form:"username"`  // 按用户名模糊搜索
	Email    string `form:"email"`     // 按邮箱搜索
	IsActive *bool  `form:"is_active"` // 按状态筛选
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int64           `json:"total"`
	Users []*UserResponse `json:"users"`
}
