// Package identity 身份认证域服务接口
// 职责：用户认证、用户管理、角色权限管理
// 核心原则：统一身份认证、细粒度权限控制、安全的会话管理
package identity

import "context"

// AuthenticateRequest 登录认证请求
type AuthenticateRequest struct {
	Username  string `json:"username"`   // 用户名
	Password  string `json:"password"`   // 密码
	IP        string `json:"ip"`         // 客户端IP
	UserAgent string `json:"user_agent"` // 用户代理
}

// AuthenticateResponse 登录认证响应
type AuthenticateResponse struct {
	Token        string   `json:"token"`         // JWT访问令牌
	RefreshToken string   `json:"refresh_token"` // 刷新令牌
	UserID       string   `json:"user_id"`       // 用户ID
	Username     string   `json:"username"`      // 用户名
	Roles        []string `json:"roles"`         // 用户角色列表
	ExpiresAt    int64    `json:"expires_at"`    // 令牌过期时间
}

// User 用户领域模型
type User struct {
	ID        int64  `json:"id"`         // 用户ID
	UUID      string `json:"uuid"`       // 用户唯一标识
	Username  string `json:"username"`   // 用户名
	Email     string `json:"email"`      // 邮箱
	Phone     string `json:"phone"`      // 手机号
	Name      string `json:"name"`       // 真实姓名
	Status    string `json:"status"`     // 状态：active/inactive/locked
	Roles     []Role `json:"roles"`      // 用户角色
	CreatedAt int64  `json:"created_at"` // 创建时间
	UpdatedAt int64  `json:"updated_at"` // 更新时间
}

// Role 角色领域模型
type Role struct {
	ID          int64        `json:"id"`           // 角色ID
	Name        string       `json:"name"`         // 角色名称
	DisplayName string       `json:"display_name"` // 显示名称
	Description string       `json:"description"`  // 角色描述
	IsActive    bool         `json:"is_active"`    // 是否激活
	Permissions []Permission `json:"permissions"`  // 角色权限
	CreatedAt   int64        `json:"created_at"`   // 创建时间
}

// Permission 权限领域模型
type Permission struct {
	ID     int64  `json:"id"`     // 权限ID
	Role   string `json:"role"`   // 角色名称
	Path   string `json:"path"`   // 资源路径
	Method string `json:"method"` // HTTP方法
	Action string `json:"action"` // 操作类型
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username        string   `json:"username"`         // 用户名
	Email           string   `json:"email"`            // 邮箱
	Phone           string   `json:"phone"`            // 手机号
	Name            string   `json:"name"`             // 真实姓名
	Password        string   `json:"password"`         // 密码
	ConfirmPassword string   `json:"confirm_password"` // 确认密码
	Roles           []string `json:"roles"`            // 分配的角色
	ManagerUUID     string   `json:"manager_uuid"`     // 管理者UUID
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email       *string  `json:"email,omitempty"`        // 邮箱
	Phone       *string  `json:"phone,omitempty"`        // 手机号
	Name        *string  `json:"name,omitempty"`         // 真实姓名
	Status      *string  `json:"status,omitempty"`       // 状态
	Roles       []string `json:"roles,omitempty"`        // 角色列表
	ManagerUUID *string  `json:"manager_uuid,omitempty"` // 管理者UUID
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name"`         // 角色名称
	DisplayName string `json:"display_name"` // 显示名称
	Description string `json:"description"`  // 角色描述
	IsActive    bool   `json:"is_active"`    // 是否激活
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	DisplayName *string `json:"display_name,omitempty"` // 显示名称
	Description *string `json:"description,omitempty"`  // 角色描述
	IsActive    *bool   `json:"is_active,omitempty"`    // 是否激活
}

// PermissionRequest 权限请求
type PermissionRequest struct {
	Role   string `json:"role"`   // 角色名称
	Path   string `json:"path"`   // 资源路径
	Method string `json:"method"` // HTTP方法
}

// AuthService 身份认证服务接口
// 负责用户认证、token管理、会话控制
type AuthService interface {
	// Authenticate 用户认证
	// 验证用户名密码，返回访问令牌和用户信息
	Authenticate(ctx context.Context, req AuthenticateRequest) (*AuthenticateResponse, error)

	// RefreshToken 刷新访问令牌
	// 使用刷新令牌获取新的访问令牌
	RefreshToken(ctx context.Context, refreshToken string) (*AuthenticateResponse, error)

	// Logout 用户登出
	// 将令牌加入黑名单，禁止后续访问
	Logout(ctx context.Context, token string) error

	// ValidateToken 验证令牌有效性
	// 检查令牌是否有效且未过期
	ValidateToken(ctx context.Context, token string) (*User, error)

	// ResetPassword 重置密码
	// 生成密码重置令牌并发送邮件
	ResetPassword(ctx context.Context, email string) error

	// ConfirmPasswordReset 确认密码重置
	// 使用重置令牌设置新密码
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}

// UserService 用户管理服务接口
// 负责用户生命周期管理、角色分配
type UserService interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)

	// GetUser 获取用户详情
	GetUser(ctx context.Context, userID string) (*User, error)

	// GetUserByUUID 根据UUID获取用户
	GetUserByUUID(ctx context.Context, uuid string) (*User, error)

	// ListUsers 分页查询用户列表
	ListUsers(ctx context.Context, page, pageSize int, status string) ([]User, int64, error)

	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) error

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error

	// DeactivateUser 停用用户
	DeactivateUser(ctx context.Context, userID string) error

	// ActivateUser 激活用户
	ActivateUser(ctx context.Context, userID string) error

	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, userID string) error

	// BatchGetUsers 批量获取用户
	BatchGetUsers(ctx context.Context, uuids []string) ([]User, error)
}

// RoleService 角色管理服务接口
// 负责角色定义、权限分配
type RoleService interface {
	// CreateRole 创建角色
	CreateRole(ctx context.Context, req CreateRoleRequest) (*Role, error)

	// GetRole 获取角色详情
	GetRole(ctx context.Context, roleID string) (*Role, error)

	// ListRoles 查询角色列表
	ListRoles(ctx context.Context) ([]Role, error)

	// UpdateRole 更新角色信息
	UpdateRole(ctx context.Context, roleID string, req UpdateRoleRequest) error

	// DeleteRole 删除角色
	DeleteRole(ctx context.Context, roleID string) error

	// AssignRoleToUser 给用户分配角色
	AssignRoleToUser(ctx context.Context, userID string, roleNames []string) error

	// RemoveRoleFromUser 移除用户角色
	RemoveRoleFromUser(ctx context.Context, userID string, roleName string) error
}

// PermissionService 权限管理服务接口
// 负责基于角色的访问控制(RBAC)
type PermissionService interface {
	// AddPermission 添加权限
	AddPermission(ctx context.Context, req PermissionRequest) error

	// RemovePermission 移除权限
	RemovePermission(ctx context.Context, req PermissionRequest) error

	// ListRolePermissions 查询角色权限
	ListRolePermissions(ctx context.Context, role string) ([]Permission, error)

	// CheckPermission 检查权限
	// 验证用户是否有权限访问指定资源
	CheckPermission(ctx context.Context, userRoles []string, path, method string) (bool, error)

	// ListAllPermissions 查询所有权限策略
	ListAllPermissions(ctx context.Context) ([]Permission, error)

	// UpdateRolePermissions 批量更新角色权限
	UpdateRolePermissions(ctx context.Context, role string, permissions []PermissionRequest) error

	// GetRolesForUser 获取用户的所有角色
	GetRolesForUser(ctx context.Context, userID string) ([]string, error)

	// GetUsersForRole 获取角色的所有用户
	GetUsersForRole(ctx context.Context, role string) ([]string, error)
}

// HierarchyService 组织架构层级服务接口
// 负责用户上下级关系管理和权限控制
type HierarchyService interface {
	// GetSubordinates 获取下属用户列表(含多级)
	// 返回指定管理者的所有下属用户ID，包括间接下属
	GetSubordinates(ctx context.Context, managerID int64) ([]int64, error)

	// CanAccessCustomer 检查用户是否可以访问指定客户
	// 基于客户分配关系和组织层级关系判断访问权限
	CanAccessCustomer(ctx context.Context, operatorID int64, customerID int64) (bool, error)

	// GetDirectReports 获取直接下属
	// 只返回直接汇报的下属，不包括间接下属
	GetDirectReports(ctx context.Context, managerID int64) ([]int64, error)

	// GetManagerChain 获取管理链
	// 返回用户的完整管理链，从直接上级到最高级别
	GetManagerChain(ctx context.Context, userID int64) ([]int64, error)
}

// Service 身份认证域统一服务接口
// 整合认证、用户、角色、权限的完整功能
type Service interface {
	AuthService
	UserService
	RoleService
	PermissionService
	HierarchyService

	// 控制器接口 - 兼容现有控制器
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req *RegisterRequest) error
	UpdateProfile(ctx context.Context, userID int64, req *UpdateProfileRequest) error
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	RealName string `json:"real_name" binding:"required"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}
