package dto

// ContactCreateRequest 创建联系人
type ContactCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Phone     string `json:"phone" binding:"omitempty,e164"`
	Email     string `json:"email" binding:"omitempty,email"`
	Position  string `json:"position"`
	IsPrimary bool   `json:"is_primary"`
	Note      string `json:"note"`
}

// ContactUpdateRequest 更新联系人
type ContactUpdateRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone" binding:"omitempty,e164"`
	Email     string `json:"email" binding:"omitempty,email"`
	Position  string `json:"position"`
	IsPrimary *bool  `json:"is_primary"`
	Note      string `json:"note"`
}

// ContactResponse 联系人响应
type ContactResponse struct {
	ID         int64  `json:"id"`
	CustomerID int64  `json:"customer_id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Position   string `json:"position"`
	IsPrimary  bool   `json:"is_primary"`
	Note       string `json:"note"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ContactListResponse 列表响应
type ContactListResponse struct {
	Total    int64              `json:"total"`
	Contacts []*ContactResponse `json:"contacts"`
}
