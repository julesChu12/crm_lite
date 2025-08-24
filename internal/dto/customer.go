package dto

// CustomerCreateRequest 创建客户的请求
type CustomerCreateRequest struct {
	Name       string   `json:"name" binding:"required"`
	Phone      string   `json:"phone" binding:"required"` // 使用e164格式校验手机号
	Email      string   `json:"email" binding:"omitempty,email"`
	Gender     string   `json:"gender"`                                           // 性别: male, female, unknown
	Birthday   string   `json:"birthday" binding:"omitempty,datetime=2006-01-02"` // 生日，格式：YYYY-MM-DD
	Level      string   `json:"level"`                                            // 客户等级: 普通, 银牌, 金牌, 铂金
	Tags       []string `json:"tags"`                                             // 标签，逗号分隔
	Note       string   `json:"note"`                                             // 备注
	Source     string   `json:"source"`                                           // 客户来源：manual, referral, marketing, etc.
	AssignedTo int64    `json:"assigned_to"`                                      // 分配给哪个员工
}

// CustomerUpdateRequest 更新客户的请求
type CustomerUpdateRequest struct {
	Name       string   `json:"name"`
	Phone      string   `json:"phone" binding:"omitempty,e164"`
	Email      string   `json:"email" binding:"omitempty,email"`
	Gender     string   `json:"gender"`                                           // 性别: male, female, unknown
	Birthday   string   `json:"birthday" binding:"omitempty,datetime=2006-01-02"` // 生日，格式：YYYY-MM-DD
	Level      string   `json:"level"`                                            // 客户等级: 普通, 银牌, 金牌, 铂金
	Tags       []string `json:"tags"`                                             // 标签，逗号分隔
	Note       string   `json:"note"`                                             // 备注
	Source     string   `json:"source"`                                           // 客户来源：manual, referral, marketing, etc.
	AssignedTo int64    `json:"assigned_to"`                                      // 分配给哪个员工
}

// CustomerListRequest 获取客户列表的请求参数
type CustomerListRequest struct {
	Page     int     `form:"page,default=1"`
	PageSize int     `form:"page_size,default=10"`
	Name     string  `form:"name"`     // 按姓名模糊搜索
	Phone    string  `form:"phone"`    // 按手机号精确搜索
	Email    string  `form:"email"`    // 按邮箱精确搜索
	OrderBy  string  `form:"order_by"` // 排序字段, e.g., created_at_desc
	IDs      []int64 `form:"ids"`      // 新增: 用于根据ID批量查询
}

// CustomerBatchGetRequest 批量获取客户的请求体
type CustomerBatchGetRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// CustomerResponse 单个客户的响应数据
type CustomerResponse struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Phone         string   `json:"phone"`
	Email         string   `json:"email"`
	Address       string   `json:"address"`
	Gender        string   `json:"gender"`
	Birthday      string   `json:"birthday,omitempty"`
	Level         string   `json:"level"`
	Tags          []string `json:"tags"` // 标签列表
	Note          string   `json:"note"`
	Source        string   `json:"source"`
	AssignedTo    int64    `json:"assigned_to"`
	WalletBalance float64  `json:"wallet_balance,omitempty"` // 兼容测试字段
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// SourceMap defines the mapping for customer sources.
var SourceMap = map[string]string{
	"manual":    "手动创建",
	"referral":  "客户推荐",
	"marketing": "营销活动",
}

// GenderMap defines the mapping for gender.
var GenderMap = map[string]string{
	"male":    "男",
	"female":  "女",
	"unknown": "未知",
}

// CustomerListResponse 客户列表的响应
type CustomerListResponse struct {
	Total     int64               `json:"total"`
	Customers []*CustomerResponse `json:"customers"`
}
