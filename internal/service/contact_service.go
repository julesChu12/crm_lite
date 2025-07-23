package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"

	"gorm.io/gorm"
)

type ContactService struct {
	q *query.Query
}

func NewContactService(resManager *resource.Manager) *ContactService {
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic(err)
	}
	return &ContactService{q: query.Use(dbRes.DB)}
}

func (s *ContactService) toResponse(c *model.Contact) *dto.ContactResponse {
	return &dto.ContactResponse{
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

// List contacts for customer
func (s *ContactService) List(ctx context.Context, customerID int64) ([]*dto.ContactResponse, error) {
	// 先检查客户是否存在
	if err := s.validateCustomerExists(ctx, customerID); err != nil {
		return nil, err
	}

	contacts, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.CustomerID.Eq(customerID)).Find()
	if err != nil {
		return nil, err
	}
	res := make([]*dto.ContactResponse, 0, len(contacts))
	for _, c := range contacts {
		res = append(res, s.toResponse(c))
	}
	return res, nil
}

// GetContactByID 获取单个联系人详情
func (s *ContactService) GetContactByID(ctx context.Context, id int64) (*dto.ContactResponse, error) {
	contact, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContactNotFound
		}
		return nil, err
	}
	return s.toResponse(contact), nil
}

// Create
func (s *ContactService) Create(ctx context.Context, customerID int64, req *dto.ContactCreateRequest) (*dto.ContactResponse, error) {
	// 先检查客户是否存在
	if err := s.validateCustomerExists(ctx, customerID); err != nil {
		return nil, err
	}

	// 检查主要联系人唯一性
	if req.IsPrimary {
		if err := s.validatePrimaryContactUniqueness(ctx, customerID, 0); err != nil {
			return nil, err
		}
	}

	// 检查手机号唯一性（同一客户下）
	if req.Phone != "" {
		if err := s.validatePhoneUniqueness(ctx, customerID, req.Phone, 0); err != nil {
			return nil, err
		}
	}

	// 检查邮箱唯一性（同一客户下）
	if req.Email != "" {
		if err := s.validateEmailUniqueness(ctx, customerID, req.Email, 0); err != nil {
			return nil, err
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
	return s.toResponse(contact), nil
}

// Update
func (s *ContactService) Update(ctx context.Context, id int64, req *dto.ContactUpdateRequest) error {
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
		if err := s.validatePrimaryContactUniqueness(ctx, existingContact.CustomerID, id); err != nil {
			return err
		}
	}

	// 检查手机号唯一性（同一客户下）
	if req.Phone != "" && req.Phone != existingContact.Phone {
		if err := s.validatePhoneUniqueness(ctx, existingContact.CustomerID, req.Phone, id); err != nil {
			return err
		}
	}

	// 检查邮箱唯一性（同一客户下）
	if req.Email != "" && req.Email != existingContact.Email {
		if err := s.validateEmailUniqueness(ctx, existingContact.CustomerID, req.Email, id); err != nil {
			return err
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
		return nil // 没有需要更新的字段
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

// Delete
func (s *ContactService) Delete(ctx context.Context, id int64) error {
	res, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrContactNotFound
	}
	return nil
}

// 私有方法：业务逻辑校验

// validateCustomerExists 检查客户是否存在
func (s *ContactService) validateCustomerExists(ctx context.Context, customerID int64) error {
	count, err := s.q.Customer.WithContext(ctx).
		Where(s.q.Customer.ID.Eq(customerID)).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

// validatePrimaryContactUniqueness 检查主要联系人唯一性
func (s *ContactService) validatePrimaryContactUniqueness(ctx context.Context, customerID int64, excludeID int64) error {
	query := s.q.Contact.WithContext(ctx).
		Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.IsPrimary.Is(true))

	if excludeID > 0 {
		query = query.Where(s.q.Contact.ID.Neq(excludeID))
	}

	count, err := query.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPrimaryContactAlreadyExists
	}
	return nil
}

// validatePhoneUniqueness 检查手机号唯一性（同一客户下）
func (s *ContactService) validatePhoneUniqueness(ctx context.Context, customerID int64, phone string, excludeID int64) error {
	query := s.q.Contact.WithContext(ctx).
		Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.Phone.Eq(phone))

	if excludeID > 0 {
		query = query.Where(s.q.Contact.ID.Neq(excludeID))
	}

	count, err := query.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrContactPhoneAlreadyExists
	}
	return nil
}

// validateEmailUniqueness 检查邮箱唯一性（同一客户下）
func (s *ContactService) validateEmailUniqueness(ctx context.Context, customerID int64, email string, excludeID int64) error {
	query := s.q.Contact.WithContext(ctx).
		Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.Email.Eq(email))

	if excludeID > 0 {
		query = query.Where(s.q.Contact.ID.Neq(excludeID))
	}

	count, err := query.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrContactEmailAlreadyExists
	}
	return nil
}
