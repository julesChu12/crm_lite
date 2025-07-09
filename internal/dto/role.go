package dto

// Role DTOs

type RoleCreateRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description"`
}

type RoleUpdateRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type RoleResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}
