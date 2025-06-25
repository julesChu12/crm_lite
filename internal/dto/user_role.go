package dto

// UserRoleRequest 用于将用户分配给角色或从角色中移除用户的请求体
type UserRoleRequest struct {
	UserID string `json:"user_id" binding:"required"` // 用户ID
	Role   string `json:"role" binding:"required"`    // 角色名称
}
