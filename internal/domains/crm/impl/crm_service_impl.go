package impl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/crm"
	"crm_lite/pkg/utils"

	"gorm.io/gorm"
)

// 错误定义
var (
	ErrCustomerNotFound            = errors.New("customer not found")
	ErrContactNotFound             = errors.New("contact not found")
	ErrPhoneAlreadyExists          = errors.New("phone already exists")
	ErrPrimaryContactAlreadyExists = errors.New("primary contact already exists")
	ErrContactPhoneAlreadyExists   = errors.New("contact phone already exists")
	ErrContactEmailAlreadyExists   = errors.New("contact email already exists")
)

// CRMServiceImpl CRM域服务实现
type CRMServiceImpl struct {
	q         *query.Query
	walletSvc WalletPort
}

// WalletPort 钱包服务端口接口 - 最小化依赖
type WalletPort interface {
	CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error)
	GetWalletByCustomerID(ctx context.Context, customerID int64) (balance int64, err error)
}

// NewCRMService 创建 CRM 服务实例
func NewCRMService(q *query.Query, walletSvc WalletPort) *CRMServiceImpl {
	return &CRMServiceImpl{
		q:         q,
		walletSvc: walletSvc,
	}
}

// 辅助方法：模型转换
func (s *CRMServiceImpl) toCustomerResponse(customer *model.Customer) *crm.CustomerResponse {
	birthday := ""
	if customer != nil && !customer.Birthday.IsZero() {
		birthday = customer.Birthday.Format("2006-01-02")
	}

	var tags []string
	if customer.Tags != "" {
		_ = json.Unmarshal([]byte(customer.Tags), &tags)
	}

	return &crm.CustomerResponse{
		ID:         customer.ID,
		Name:       customer.Name,
		Phone:      customer.Phone,
		Email:      customer.Email,
		Gender:     customer.Gender,
		Birthday:   birthday,
		Level:      customer.Level,
		Tags:       tags,
		Note:       customer.Note,
		Source:     customer.Source,
		AssignedTo: customer.AssignedTo,
		CreatedAt:  utils.FormatTime(customer.CreatedAt),
		UpdatedAt:  utils.FormatTime(customer.UpdatedAt),
	}
}

func (s *CRMServiceImpl) toContactResponse(c *model.Contact) *crm.ContactResponse {
	return &crm.ContactResponse{
		ID:         c.ID,
		CustomerID: c.CustomerID,
		Name:       c.Name,
		Phone:      c.Phone,
		Email:      c.Email,
		Position:   c.Position,
		IsPrimary:  c.IsPrimary,
		Note:       c.Note,
		CreatedAt:  utils.FormatTime(c.CreatedAt),
		UpdatedAt:  utils.FormatTime(c.UpdatedAt),
	}
}

// 核心域接口实现（新接口）

// CreateCustomer 创建客户
func (s *CRMServiceImpl) CreateCustomer(ctx context.Context, req crm.CreateCustomerReq) (*crm.Customer, error) {
	// TODO: 实现域模型的 CreateCustomer
	return nil, errors.New("domain CreateCustomer not implemented yet")
}

// GetCustomer 获取客户
func (s *CRMServiceImpl) GetCustomer(ctx context.Context, customerID int64) (*crm.Customer, error) {
	// TODO: 实现域模型的 GetCustomer
	return nil, errors.New("domain GetCustomer not implemented yet")
}

// GetCustomerByPhone 根据手机号获取客户
func (s *CRMServiceImpl) GetCustomerByPhone(ctx context.Context, phone string) (*crm.Customer, error) {
	// TODO: 实现域模型的 GetCustomerByPhone
	return nil, errors.New("domain GetCustomerByPhone not implemented yet")
}

// UpdateCustomer 更新客户
func (s *CRMServiceImpl) UpdateCustomer(ctx context.Context, req crm.UpdateCustomerReq) (*crm.Customer, error) {
	// TODO: 实现域模型的 UpdateCustomer
	return nil, errors.New("domain UpdateCustomer not implemented yet")
}

// ListCustomers 客户列表
func (s *CRMServiceImpl) ListCustomers(ctx context.Context, assignedTo int64, level, keyword string, page, pageSize int) ([]crm.Customer, error) {
	// TODO: 实现域模型的 ListCustomers
	return []crm.Customer{}, errors.New("domain ListCustomers not implemented yet")
}

// DeleteCustomer 删除客户
func (s *CRMServiceImpl) DeleteCustomer(ctx context.Context, customerID int64) error {
	// TODO: 实现域模型的 DeleteCustomer
	return errors.New("domain DeleteCustomer not implemented yet")
}

// CreateContact 创建联系人
func (s *CRMServiceImpl) CreateContact(ctx context.Context, contact *crm.Contact) (*crm.Contact, error) {
	// TODO: 实现域模型的 CreateContact
	return nil, errors.New("domain CreateContact not implemented yet")
}

// GetContact 获取联系人
func (s *CRMServiceImpl) GetContact(ctx context.Context, contactID int64) (*crm.Contact, error) {
	// TODO: 实现域模型的 GetContact
	return nil, errors.New("domain GetContact not implemented yet")
}

// ListContactsByCustomer 获取客户联系人
func (s *CRMServiceImpl) ListContactsByCustomer(ctx context.Context, customerID int64) ([]crm.Contact, error) {
	// TODO: 实现域模型的 ListContactsByCustomer
	return []crm.Contact{}, errors.New("domain ListContactsByCustomer not implemented yet")
}

// UpdateContact 更新联系人
func (s *CRMServiceImpl) UpdateContact(ctx context.Context, contact *crm.Contact) (*crm.Contact, error) {
	// TODO: 实现域模型的 UpdateContact
	return nil, errors.New("domain UpdateContact not implemented yet")
}

// DeleteContact 删除联系人
func (s *CRMServiceImpl) DeleteContact(ctx context.Context, contactID int64) error {
	// TODO: 实现域模型的 DeleteContact
	return errors.New("domain DeleteContact not implemented yet")
}

// SetPrimaryContact 设置主联系人
func (s *CRMServiceImpl) SetPrimaryContact(ctx context.Context, customerID, contactID int64) error {
	// TODO: 实现域模型的 SetPrimaryContact
	return errors.New("domain SetPrimaryContact not implemented yet")
}

// Legacy 兼容接口实现（与现有控制器兼容）

// CreateCustomerLegacy 创建客户 - 兼容现有实现
func (s *CRMServiceImpl) CreateCustomerLegacy(ctx context.Context, req *crm.CustomerCreateRequest) (*crm.CustomerResponse, error) {
	var customerToReturn *model.Customer
	isNewCreation := false

	if req.Phone != "" {
		existingCustomer, err := s.q.Customer.WithContext(ctx).Unscoped().Where(s.q.Customer.Phone.Eq(req.Phone)).First()
		if err == nil {
			if existingCustomer.DeletedAt.Valid {
				// 恢复并更新
				existingCustomer.Name = req.Name
				existingCustomer.Email = req.Email
				existingCustomer.Gender = req.Gender
				existingCustomer.Level = req.Level

				tagsJSON, err := json.Marshal(req.Tags)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tags: %w", err)
				}
				existingCustomer.Tags = string(tagsJSON)

				existingCustomer.Note = req.Note
				existingCustomer.Source = req.Source
				existingCustomer.AssignedTo = req.AssignedTo

				if req.Birthday != "" {
					birthday, err := time.Parse("2006-01-02", req.Birthday)
					if err != nil {
						return nil, fmt.Errorf("invalid birthday format: %w", err)
					}
					existingCustomer.Birthday = birthday
				}

				// 使用 map 来构建更新
				updates := map[string]interface{}{
					"name":        existingCustomer.Name,
					"email":       existingCustomer.Email,
					"gender":      existingCustomer.Gender,
					"level":       existingCustomer.Level,
					"tags":        existingCustomer.Tags,
					"note":        existingCustomer.Note,
					"source":      existingCustomer.Source,
					"assigned_to": existingCustomer.AssignedTo,
					"birthday":    existingCustomer.Birthday,
					"deleted_at":  nil,
				}
				if _, err := s.q.Customer.WithContext(ctx).Unscoped().Where(s.q.Customer.ID.Eq(existingCustomer.ID)).Updates(updates); err != nil {
					return nil, err
				}
				customerToReturn = existingCustomer
			} else {
				return nil, ErrPhoneAlreadyExists
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if customerToReturn == nil {
		customer := &model.Customer{
			Name:       req.Name,
			Phone:      req.Phone,
			Email:      req.Email,
			Gender:     req.Gender,
			Level:      req.Level,
			Note:       req.Note,
			Source:     req.Source,
			AssignedTo: req.AssignedTo,
		}

		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		customer.Tags = string(tagsJSON)

		if req.Birthday != "" {
			birthday, err := time.Parse("2006-01-02", req.Birthday)
			if err != nil {
				return nil, fmt.Errorf("invalid birthday format: %w", err)
			}
			customer.Birthday = birthday
		}
		if err := s.q.Customer.WithContext(ctx).Create(customer); err != nil {
			return nil, err
		}
		customerToReturn = customer
		isNewCreation = true
	}

	// 创建钱包
	if _, err := s.walletSvc.CreateWallet(ctx, customerToReturn.ID, "balance"); err != nil {
		if isNewCreation {
			return nil, fmt.Errorf("failed to create wallet for new customer: %w", err)
		}
	}

	return s.toCustomerResponse(customerToReturn), nil
}

// GetCustomerByIDLegacy 根据 ID 获取客户
func (s *CRMServiceImpl) GetCustomerByIDLegacy(ctx context.Context, id string) (*crm.CustomerResponse, error) {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return nil, ErrCustomerNotFound
	}
	customer, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(idNum), s.q.Customer.DeletedAt.IsNull()).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	resp := s.toCustomerResponse(customer)
	if balance, errW := s.walletSvc.GetWalletByCustomerID(ctx, customer.ID); errW == nil {
		resp.WalletBalance = balance
	}
	return resp, nil
}

// ListCustomersLegacy 获取客户列表
func (s *CRMServiceImpl) ListCustomersLegacy(ctx context.Context, req *crm.CustomerListRequest) (*crm.CustomerListResponse, error) {
	q := s.q.Customer.WithContext(ctx).Where(s.q.Customer.DeletedAt.IsNull())

	if len(req.IDs) > 0 {
		q = q.Where(s.q.Customer.ID.In(req.IDs...))
	} else {
		if req.Name != "" {
			q = q.Where(s.q.Customer.Name.Like("%" + req.Name + "%"))
		}
		if req.Phone != "" {
			q = q.Where(s.q.Customer.Phone.Eq(req.Phone))
		}
		if req.Email != "" {
			q = q.Where(s.q.Customer.Email.Eq(req.Email))
		}
	}

	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			field := parts[0]
			order := parts[1]
			if col, ok := s.q.Customer.GetFieldByName(field); ok {
				if order == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
		q = q.Order(s.q.Customer.CreatedAt.Desc())
	}

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	var customers []*model.Customer
	if len(req.IDs) == 0 && req.PageSize > 0 {
		customers, err = q.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find()
	} else {
		customers, err = q.Find()
	}
	if err != nil {
		return nil, err
	}

	customerResponses := make([]*crm.CustomerResponse, 0, len(customers))
	for _, c := range customers {
		resp := s.toCustomerResponse(c)
		if balance, errW := s.walletSvc.GetWalletByCustomerID(ctx, c.ID); errW == nil {
			resp.WalletBalance = balance
		}
		customerResponses = append(customerResponses, resp)
	}

	return &crm.CustomerListResponse{
		Total:     total,
		Customers: customerResponses,
	}, nil
}

// UpdateCustomerLegacy 更新客户
func (s *CRMServiceImpl) UpdateCustomerLegacy(ctx context.Context, id string, req *crm.CustomerUpdateRequest) error {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return ErrCustomerNotFound
	}

	if req.Phone != "" {
		count, err := s.q.Customer.WithContext(ctx).
			Where(s.q.Customer.Phone.Eq(req.Phone), s.q.Customer.ID.Neq(idNum)).
			Where(s.q.Customer.DeletedAt.IsNull()).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrPhoneAlreadyExists
		}
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Gender != "" {
		updates["gender"] = req.Gender
	}
	if req.Level != "" {
		updates["level"] = req.Level
	}
	if req.Tags != nil {
		tagsJSON, err := json.Marshal(*req.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshal tags: %w", err)
		}
		updates["tags"] = string(tagsJSON)
	}
	if req.Note != "" {
		updates["note"] = req.Note
	}
	if req.Source != "" {
		updates["source"] = req.Source
	}
	if req.AssignedTo != 0 {
		updates["assigned_to"] = req.AssignedTo
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return fmt.Errorf("invalid birthday format: %w", err)
		}
		updates["birthday"] = birthday
	}

	if len(updates) == 0 {
		return nil
	}

	result, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(idNum)).Updates(updates)
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

// DeleteCustomerLegacy 删除客户
func (s *CRMServiceImpl) DeleteCustomerLegacy(ctx context.Context, id string) error {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return ErrCustomerNotFound
	}
	result, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(idNum)).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

// Contact Legacy 实现

// ListContactsLegacy 获取客户联系人列表
func (s *CRMServiceImpl) ListContactsLegacy(ctx context.Context, customerID int64) ([]*crm.ContactResponse, error) {
	// 先检查客户是否存在
	count, err := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.ID.Eq(customerID)).
		Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrCustomerNotFound
	}

	contacts, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.CustomerID.Eq(customerID)).Find()
	if err != nil {
		return nil, err
	}
	res := make([]*crm.ContactResponse, 0, len(contacts))
	for _, c := range contacts {
		res = append(res, s.toContactResponse(c))
	}
	return res, nil
}

// GetContactByIDLegacy 获取单个联系人详情
func (s *CRMServiceImpl) GetContactByIDLegacy(ctx context.Context, id int64) (*crm.ContactResponse, error) {
	contact, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, err
	}
	return s.toContactResponse(contact), nil
}

// CreateContactLegacy 创建联系人
func (s *CRMServiceImpl) CreateContactLegacy(ctx context.Context, customerID int64, req *crm.ContactCreateRequest) (*crm.ContactResponse, error) {
	// 先检查客户是否存在
	count, err := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.ID.Eq(customerID)).
		Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrCustomerNotFound
	}

	// 检查主要联系人唯一性
	if req.IsPrimary {
		count, err := s.q.Contact.WithContext(ctx).
			Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.IsPrimary.Is(true)).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrPrimaryContactAlreadyExists
		}
	}

	// 检查手机号唯一性（同一客户下）
	if req.Phone != "" {
		count, err := s.q.Contact.WithContext(ctx).
			Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.Phone.Eq(req.Phone)).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrContactPhoneAlreadyExists
		}
	}

	// 检查邮箱唯一性（同一客户下）
	if req.Email != "" {
		count, err := s.q.Contact.WithContext(ctx).
			Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.Email.Eq(req.Email)).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrContactEmailAlreadyExists
		}
	}

	contact := &model.Contact{
		CustomerID: customerID,
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      req.Email,
		Position:   req.Position,
		IsPrimary:  req.IsPrimary,
		Note:       req.Note,
	}
	if err := s.q.Contact.WithContext(ctx).Create(contact); err != nil {
		return nil, err
	}
	return s.toContactResponse(contact), nil
}

// UpdateContactLegacy 更新联系人
func (s *CRMServiceImpl) UpdateContactLegacy(ctx context.Context, id int64, req *crm.ContactUpdateRequest) error {
	// 先检查联系人是否存在并获取当前信息
	existingContact, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	// 检查主要联系人唯一性
	if req.IsPrimary != nil && *req.IsPrimary {
		count, err := s.q.Contact.WithContext(ctx).
			Where(s.q.Contact.CustomerID.Eq(existingContact.CustomerID), s.q.Contact.IsPrimary.Is(true)).
			Where(s.q.Contact.ID.Neq(id)).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrPrimaryContactAlreadyExists
		}
	}

	// 检查手机号唯一性（同一客户下）
	if req.Phone != "" && req.Phone != existingContact.Phone {
		count, err := s.q.Contact.WithContext(ctx).
			Where(s.q.Contact.CustomerID.Eq(existingContact.CustomerID), s.q.Contact.Phone.Eq(req.Phone)).
			Where(s.q.Contact.ID.Neq(id)).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrContactPhoneAlreadyExists
		}
	}

	// 检查邮箱唯一性（同一客户下）
	if req.Email != "" && req.Email != existingContact.Email {
		count, err := s.q.Contact.WithContext(ctx).
			Where(s.q.Contact.CustomerID.Eq(existingContact.CustomerID), s.q.Contact.Email.Eq(req.Email)).
			Where(s.q.Contact.ID.Neq(id)).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrContactEmailAlreadyExists
		}
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Position != "" {
		updates["position"] = req.Position
	}
	if req.IsPrimary != nil {
		updates["is_primary"] = *req.IsPrimary
	}
	if req.Note != "" {
		updates["note"] = req.Note
	}

	if len(updates) == 0 {
		return nil
	}

	res, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrContactNotFound
	}
	return nil
}

// DeleteContactLegacy 删除联系人
func (s *CRMServiceImpl) DeleteContactLegacy(ctx context.Context, id int64) error {
	res, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrContactNotFound
	}
	return nil
}

// 断言接口实现
var _ crm.Service = (*CRMServiceImpl)(nil)
