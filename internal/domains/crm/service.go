// Package crm 客户关系管理域服务接口
// 职责：客户信息管理、联系人管理、客户关系维护
package crm

import "context"

// Customer 客户领域模型
// 表示CRM系统中的客户核心信息
type Customer struct {
	ID         int64    `json:"id"`          // 客户ID
	Name       string   `json:"name"`        // 客户姓名
	Phone      string   `json:"phone"`       // 手机号（唯一标识）
	Email      string   `json:"email"`       // 邮箱
	Gender     string   `json:"gender"`      // 性别：male/female/unknown
	Birthday   string   `json:"birthday"`    // 生日（YYYY-MM-DD格式）
	Level      string   `json:"level"`       // 客户等级：普通/银牌/金牌/铂金
	Tags       []string `json:"tags"`        // 客户标签列表
	Note       string   `json:"note"`        // 备注信息
	Source     string   `json:"source"`      // 客户来源：manual/referral/marketing
	AssignedTo int64    `json:"assigned_to"` // 分配的销售员工ID
	CreatedAt  int64    `json:"created_at"`  // 创建时间
	UpdatedAt  int64    `json:"updated_at"`  // 更新时间
}

// Contact 联系人领域模型
// 表示客户的联系人信息，一个客户可以有多个联系人
type Contact struct {
	ID         int64  `json:"id"`          // 联系人ID
	CustomerID int64  `json:"customer_id"` // 所属客户ID
	Name       string `json:"name"`        // 联系人姓名
	Phone      string `json:"phone"`       // 联系人电话
	Email      string `json:"email"`       // 联系人邮箱
	Position   string `json:"position"`    // 职位
	IsPrimary  bool   `json:"is_primary"`  // 是否主联系人
	Note       string `json:"note"`        // 备注
	CreatedAt  int64  `json:"created_at"`  // 创建时间
	UpdatedAt  int64  `json:"updated_at"`  // 更新时间
}

// CreateCustomerReq 创建客户请求
// 用于封装创建客户所需的信息
type CreateCustomerReq struct {
	Name       string   `json:"name" binding:"required"`  // 客户姓名（必填）
	Phone      string   `json:"phone" binding:"required"` // 手机号（必填，唯一）
	Email      string   `json:"email"`                    // 邮箱（可选）
	Gender     string   `json:"gender"`                   // 性别（可选）
	Birthday   string   `json:"birthday"`                 // 生日（可选）
	Level      string   `json:"level"`                    // 客户等级（可选）
	Tags       []string `json:"tags"`                     // 标签（可选）
	Note       string   `json:"note"`                     // 备注（可选）
	Source     string   `json:"source"`                   // 来源（可选）
	AssignedTo int64    `json:"assigned_to"`              // 分配员工（可选）
}

// UpdateCustomerReq 更新客户请求
// 用于封装更新客户信息的请求
type UpdateCustomerReq struct {
	ID         int64    `json:"id" binding:"required"` // 客户ID（必填）
	Name       string   `json:"name"`                  // 客户姓名
	Phone      string   `json:"phone"`                 // 手机号
	Email      string   `json:"email"`                 // 邮箱
	Gender     string   `json:"gender"`                // 性别
	Birthday   string   `json:"birthday"`              // 生日
	Level      string   `json:"level"`                 // 客户等级
	Tags       []string `json:"tags"`                  // 标签
	Note       string   `json:"note"`                  // 备注
	AssignedTo int64    `json:"assigned_to"`           // 分配员工
}

// CustomerCreateRequest 创建客户请求 - 兼容现有 DTO
type CustomerCreateRequest struct {
	Name       string   `json:"name" binding:"required"`
	Phone      string   `json:"phone" binding:"required"`
	Email      string   `json:"email"`
	Gender     string   `json:"gender"`
	Birthday   string   `json:"birthday"`
	Level      string   `json:"level"`
	Tags       []string `json:"tags"`
	Note       string   `json:"note"`
	Source     string   `json:"source"`
	AssignedTo int64    `json:"assigned_to"`
}

// CustomerUpdateRequest 更新客户请求 - 兼容现有 DTO
type CustomerUpdateRequest struct {
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Gender     string    `json:"gender"`
	Birthday   string    `json:"birthday"`
	Level      string    `json:"level"`
	Tags       *[]string `json:"tags"` // 指针类型，区分未提供和空数组
	Note       string    `json:"note"`
	Source     string    `json:"source"`
	AssignedTo int64     `json:"assigned_to"`
}

// CustomerListRequest 客户列表请求 - 兼容现有 DTO
type CustomerListRequest struct {
	Page     int     `json:"page" binding:"min=1"`
	PageSize int     `json:"page_size" binding:"min=1,max=100"`
	IDs      []int64 `json:"ids"`
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	OrderBy  string  `json:"order_by"`
}

// CustomerResponse 客户响应 - 兼容现有 DTO
type CustomerResponse struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Phone         string   `json:"phone"`
	Email         string   `json:"email"`
	Gender        string   `json:"gender"`
	Birthday      string   `json:"birthday"`
	Level         string   `json:"level"`
	Tags          []string `json:"tags"`
	Note          string   `json:"note"`
	Source        string   `json:"source"`
	AssignedTo    int64    `json:"assigned_to"`
	WalletBalance int64    `json:"wallet_balance"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// CustomerListResponse 客户列表响应 - 兼容现有 DTO
type CustomerListResponse struct {
	Total     int64               `json:"total"`
	Customers []*CustomerResponse `json:"customers"`
}

// ContactCreateRequest 创建联系人请求 - 兼容现有 DTO
type ContactCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Position  string `json:"position"`
	IsPrimary bool   `json:"is_primary"`
	Note      string `json:"note"`
}

// ContactUpdateRequest 更新联系人请求 - 兼容现有 DTO
type ContactUpdateRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Position  string `json:"position"`
	IsPrimary *bool  `json:"is_primary"`
	Note      string `json:"note"`
}

// ContactResponse 联系人响应 - 兼容现有 DTO
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

// Service 客户关系管理域服务接口
// 提供客户和联系人相关的核心业务操作
type Service interface {
	// Customer相关操作

	// CreateCustomer 创建客户
	// 创建新客户并自动创建对应的钱包
	CreateCustomer(ctx context.Context, req CreateCustomerReq) (*Customer, error)

	// GetCustomer 根据ID获取客户信息
	GetCustomer(ctx context.Context, customerID int64) (*Customer, error)

	// GetCustomerByPhone 根据手机号获取客户信息
	// 手机号是客户的唯一标识，常用于验证和查找
	GetCustomerByPhone(ctx context.Context, phone string) (*Customer, error)

	// UpdateCustomer 更新客户信息
	UpdateCustomer(ctx context.Context, req UpdateCustomerReq) (*Customer, error)

	// ListCustomers 分页查询客户列表
	// 支持按姓名、电话、等级、分配员工等条件筛选
	ListCustomers(ctx context.Context, assignedTo int64, level, keyword string, page, pageSize int) ([]Customer, error)

	// DeleteCustomer 软删除客户
	// 不直接删除数据，而是标记删除状态
	DeleteCustomer(ctx context.Context, customerID int64) error

	// Contact相关操作

	// CreateContact 创建联系人
	CreateContact(ctx context.Context, contact *Contact) (*Contact, error)

	// GetContact 根据ID获取联系人信息
	GetContact(ctx context.Context, contactID int64) (*Contact, error)

	// ListContactsByCustomer 获取客户的所有联系人
	ListContactsByCustomer(ctx context.Context, customerID int64) ([]Contact, error)

	// UpdateContact 更新联系人信息
	UpdateContact(ctx context.Context, contact *Contact) (*Contact, error)

	// DeleteContact 删除联系人
	DeleteContact(ctx context.Context, contactID int64) error

	// SetPrimaryContact 设置主联系人
	// 确保一个客户只有一个主联系人
	SetPrimaryContact(ctx context.Context, customerID, contactID int64) error

	// Legacy 兼容现有控制器接口
	CreateCustomerLegacy(ctx context.Context, req *CustomerCreateRequest) (*CustomerResponse, error)
	GetCustomerByIDLegacy(ctx context.Context, id string) (*CustomerResponse, error)
	ListCustomersLegacy(ctx context.Context, req *CustomerListRequest) (*CustomerListResponse, error)
	UpdateCustomerLegacy(ctx context.Context, id string, req *CustomerUpdateRequest) error
	DeleteCustomerLegacy(ctx context.Context, id string) error

	// Contact Legacy 兼容现有控制器接口
	ListContactsLegacy(ctx context.Context, customerID int64) ([]*ContactResponse, error)
	GetContactByIDLegacy(ctx context.Context, id int64) (*ContactResponse, error)
	CreateContactLegacy(ctx context.Context, customerID int64, req *ContactCreateRequest) (*ContactResponse, error)
	UpdateContactLegacy(ctx context.Context, id int64, req *ContactUpdateRequest) error
	DeleteContactLegacy(ctx context.Context, id int64) error
}

// Repository 客户关系管理域数据访问接口
// 定义客户和联系人相关的数据持久化操作
type Repository interface {
	// Customer Repository
	CreateCustomer(ctx context.Context, customer *Customer) error
	GetCustomerByID(ctx context.Context, id int64) (*Customer, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*Customer, error)
	UpdateCustomer(ctx context.Context, customer *Customer) error
	ListCustomers(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]Customer, error)
	SoftDeleteCustomer(ctx context.Context, id int64) error

	// Contact Repository
	CreateContact(ctx context.Context, contact *Contact) error
	GetContactByID(ctx context.Context, id int64) (*Contact, error)
	ListContactsByCustomerID(ctx context.Context, customerID int64) ([]Contact, error)
	UpdateContact(ctx context.Context, contact *Contact) error
	DeleteContact(ctx context.Context, id int64) error
	UpdatePrimaryContact(ctx context.Context, customerID, contactID int64) error
}
