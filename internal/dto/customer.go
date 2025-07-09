package dto

import "time"

// CustomerCreateRequest 创建客户请求体
type CustomerCreateRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
	Email string `json:"email" binding:"omitempty,email"`
}

// CustomerUpdateRequest 更新客户请求体
type CustomerUpdateRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email" binding:"omitempty,email"`
}

// CustomerResponse 客户信息响应体
type CustomerResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
