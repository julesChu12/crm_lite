package domains

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"crm_lite/internal/constants"
)

// DomainError 域错误基类
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Field, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 预定义的错误代码
const (
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeBusinessRule     = "BUSINESS_RULE_VIOLATION"
	ErrCodeResourceNotFound = "RESOURCE_NOT_FOUND"
	ErrCodeConcurrency      = "CONCURRENCY_CONFLICT"
)

// CustomerDomain 客户域模型
type CustomerDomain struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Gender     string    `json:"gender"`
	Birthday   time.Time `json:"birthday"`
	Level      string    `json:"level"`
	Tags       []string  `json:"tags"`
	Source     string    `json:"source"`
	AssignedTo int64     `json:"assigned_to"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate 验证客户数据
func (c *CustomerDomain) Validate() error {
	var validationErrors []error

	// 姓名验证
	if c.Name == "" {
		validationErrors = append(validationErrors, DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "name",
			Message: "客户姓名不能为空",
		})
	}
	if len(c.Name) > 50 {
		validationErrors = append(validationErrors, DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "name",
			Message: "客户姓名长度不能超过50个字符",
		})
	}

	// 手机号验证
	if err := c.validatePhone(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// 邮箱验证
	if err := c.validateEmail(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// 性别验证
	if err := c.validateGender(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// 客户等级验证
	if err := c.validateLevel(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if len(validationErrors) > 0 {
		return combineErrors(validationErrors)
	}

	return nil
}

// validatePhone 验证手机号
func (c *CustomerDomain) validatePhone() error {
	if c.Phone == "" {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "phone",
			Message: "手机号不能为空",
		}
	}

	// 中国手机号正则
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(c.Phone) {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "phone",
			Message: "手机号格式不正确",
		}
	}

	return nil
}

// validateEmail 验证邮箱
func (c *CustomerDomain) validateEmail() error {
	if c.Email == "" {
		return nil // 邮箱可选
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(c.Email) {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "email",
			Message: "邮箱格式不正确",
		}
	}

	return nil
}

// validateGender 验证性别
func (c *CustomerDomain) validateGender() error {
	validGenders := constants.ValidCustomerGenders()
	for _, valid := range validGenders {
		if c.Gender == valid {
			return nil
		}
	}

	return DomainError{
		Code:    ErrCodeInvalidInput,
		Field:   "gender",
		Message: "性别必须是 male、female 或 unknown",
	}
}

// validateLevel 验证客户等级
func (c *CustomerDomain) validateLevel() error {
	validLevels := constants.ValidCustomerLevels()
	for _, valid := range validLevels {
		if c.Level == valid {
			return nil
		}
	}

	return DomainError{
		Code:    ErrCodeInvalidInput,
		Field:   "level",
		Message: "客户等级必须是：普通、银牌、金牌、铂金之一",
	}
}

// ProductDomain 产品域模型
type ProductDomain struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Type          string    `json:"type"`
	Category      string    `json:"category"`
	Price         float64   `json:"price"`         // 单价（元）
	Cost          float64   `json:"cost"`          // 成本（元）
	StockQuantity int32     `json:"stock_quantity"`
	MinStockLevel int32     `json:"min_stock_level"`
	Unit          string    `json:"unit"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate 验证产品数据
func (p *ProductDomain) Validate() error {
	var validationErrors []error

	// 产品名称验证
	if p.Name == "" {
		validationErrors = append(validationErrors, DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "name",
			Message: "产品名称不能为空",
		})
	}
	if len(p.Name) > 100 {
		validationErrors = append(validationErrors, DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "name",
			Message: "产品名称长度不能超过100个字符",
		})
	}

	// 产品类型验证
	if err := p.validateType(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// 价格验证
	if err := p.validatePrice(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// 库存验证
	if err := p.validateStock(); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if len(validationErrors) > 0 {
		return combineErrors(validationErrors)
	}

	return nil
}

// validateType 验证产品类型
func (p *ProductDomain) validateType() error {
	validTypes := constants.ValidProductTypes()
	for _, valid := range validTypes {
		if p.Type == valid {
			return nil
		}
	}

	return DomainError{
		Code:    ErrCodeInvalidInput,
		Field:   "type",
		Message: "产品类型必须是 product 或 service",
	}
}

// validatePrice 验证价格
func (p *ProductDomain) validatePrice() error {
	if p.Price < 0 {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "price",
			Message: "产品价格不能为负数",
		}
	}

	if p.Cost < 0 {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "cost",
			Message: "产品成本不能为负数",
		}
	}

	if p.Cost > p.Price {
		return DomainError{
			Code:    ErrCodeBusinessRule,
			Field:   "cost",
			Message: "产品成本不能高于销售价格",
		}
	}

	return nil
}

// validateStock 验证库存
func (p *ProductDomain) validateStock() error {
	if p.StockQuantity < 0 {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "stock_quantity",
			Message: "库存数量不能为负数",
		}
	}

	if p.MinStockLevel < 0 {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "min_stock_level",
			Message: "最小库存预警不能为负数",
		}
	}

	return nil
}

// IsLowStock 检查是否库存不足
func (p *ProductDomain) IsLowStock() bool {
	return p.StockQuantity <= p.MinStockLevel
}

// CanSell 检查是否可以销售
func (p *ProductDomain) CanSell(quantity int32) error {
	if !p.IsActive {
		return DomainError{
			Code:    ErrCodeBusinessRule,
			Message: "产品已停用，无法销售",
		}
	}

	if p.Type == "product" && p.StockQuantity < quantity {
		return DomainError{
			Code:    ErrCodeBusinessRule,
			Message: fmt.Sprintf("库存不足，当前库存：%d，请求数量：%d", p.StockQuantity, quantity),
		}
	}

	return nil
}

// WalletDomain 钱包域模型
type WalletDomain struct {
	ID         int64     `json:"id"`
	CustomerID int64     `json:"customer_id"`
	Balance    int64     `json:"balance"`    // 余额（分）
	Status     int32     `json:"status"`     // 状态：1-正常，0-冻结
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  int64     `json:"updated_at"` // 乐观锁版本号
}

// Validate 验证钱包数据
func (w *WalletDomain) Validate() error {
	var validationErrors []error

	if w.CustomerID <= 0 {
		validationErrors = append(validationErrors, DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "customer_id",
			Message: "客户ID必须大于0",
		})
	}

	if w.Balance < 0 {
		validationErrors = append(validationErrors, DomainError{
			Code:    ErrCodeBusinessRule,
			Field:   "balance",
			Message: "钱包余额不能为负数",
		})
	}

	if w.Status != 0 && w.Status != 1 {
		validationErrors = append(validationErrors, DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "status",
			Message: "钱包状态必须是0（冻结）或1（正常）",
		})
	}

	if len(validationErrors) > 0 {
		return combineErrors(validationErrors)
	}

	return nil
}

// IsActive 检查钱包是否可用
func (w *WalletDomain) IsActive() bool {
	return w.Status == 1
}

// CanDebit 检查是否可以扣款
func (w *WalletDomain) CanDebit(amount int64) error {
	if !w.IsActive() {
		return DomainError{
			Code:    ErrCodeBusinessRule,
			Message: "钱包已冻结，无法扣款",
		}
	}

	if amount <= 0 {
		return DomainError{
			Code:    ErrCodeInvalidInput,
			Field:   "amount",
			Message: "扣款金额必须大于0",
		}
	}

	if w.Balance < amount {
		return DomainError{
			Code:    ErrCodeBusinessRule,
			Message: fmt.Sprintf("余额不足，当前余额：%.2f元，请求扣款：%.2f元",
				float64(w.Balance)/100, float64(amount)/100),
		}
	}

	return nil
}

// GetBalanceYuan 获取余额（元）
func (w *WalletDomain) GetBalanceYuan() float64 {
	return float64(w.Balance) / 100.0
}

// combineErrors 合并多个错误
func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return errs[0]
	}

	// 创建组合错误信息
	message := "验证失败："
	for i, err := range errs {
		if i > 0 {
			message += "; "
		}
		message += err.Error()
	}

	return errors.New(message)
}