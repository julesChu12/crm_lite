package impl

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/utils"
	"errors"

	"gorm.io/gorm"
)

type ContactService struct{ q *query.Query }

func NewContactServiceWithQuery(q *query.Query) *ContactService { return &ContactService{q: q} }

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

func (s *ContactService) List(ctx context.Context, customerID int64) ([]*dto.ContactResponse, error) {
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

func (s *ContactService) GetContactByID(ctx context.Context, id int64) (*dto.ContactResponse, error) {
	contact, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrContactNotFound
		}
		return nil, err
	}
	return s.toResponse(contact), nil
}

func (s *ContactService) Create(ctx context.Context, customerID int64, req *dto.ContactCreateRequest) (*dto.ContactResponse, error) {
	if err := s.validateCustomerExists(ctx, customerID); err != nil {
		return nil, err
	}
	if req.IsPrimary {
		if err := s.validatePrimaryContactUniqueness(ctx, customerID, 0); err != nil {
			return nil, err
		}
	}
	if req.Phone != "" {
		if err := s.validatePhoneUniqueness(ctx, customerID, req.Phone, 0); err != nil {
			return nil, err
		}
	}
	if req.Email != "" {
		if err := s.validateEmailUniqueness(ctx, customerID, req.Email, 0); err != nil {
			return nil, err
		}
	}
	contact := &model.Contact{CustomerID: customerID, Name: req.Name, Phone: req.Phone, Email: req.Email, Position: req.Position, IsPrimary: req.IsPrimary, Note: req.Note}
	if err := s.q.Contact.WithContext(ctx).Create(contact); err != nil {
		return nil, err
	}
	return s.toResponse(contact), nil
}

func (s *ContactService) Update(ctx context.Context, id int64, req *dto.ContactUpdateRequest) error {
	existing, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return service.ErrContactNotFound
		}
		return err
	}
	if req.IsPrimary != nil && *req.IsPrimary {
		if err := s.validatePrimaryContactUniqueness(ctx, existing.CustomerID, id); err != nil {
			return err
		}
	}
	if req.Phone != "" && req.Phone != existing.Phone {
		if err := s.validatePhoneUniqueness(ctx, existing.CustomerID, req.Phone, id); err != nil {
			return err
		}
	}
	if req.Email != "" && req.Email != existing.Email {
		if err := s.validateEmailUniqueness(ctx, existing.CustomerID, req.Email, id); err != nil {
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
		return nil
	}
	res, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return service.ErrContactNotFound
	}
	return nil
}

func (s *ContactService) Delete(ctx context.Context, id int64) error {
	res, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return service.ErrContactNotFound
	}
	return nil
}

// 校验辅助
func (s *ContactService) validateCustomerExists(ctx context.Context, customerID int64) error {
	count, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(customerID)).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return service.ErrCustomerNotFound
	}
	return nil
}

func (s *ContactService) validatePrimaryContactUniqueness(ctx context.Context, customerID int64, excludeID int64) error {
	q := s.q.Contact.WithContext(ctx).Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.IsPrimary.Is(true))
	if excludeID > 0 {
		q = q.Where(s.q.Contact.ID.Neq(excludeID))
	}
	count, err := q.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return service.ErrPrimaryContactAlreadyExists
	}
	return nil
}

func (s *ContactService) validatePhoneUniqueness(ctx context.Context, customerID int64, phone string, excludeID int64) error {
	q := s.q.Contact.WithContext(ctx).Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.Phone.Eq(phone))
	if excludeID > 0 {
		q = q.Where(s.q.Contact.ID.Neq(excludeID))
	}
	count, err := q.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return service.ErrContactPhoneAlreadyExists
	}
	return nil
}

func (s *ContactService) validateEmailUniqueness(ctx context.Context, customerID int64, email string, excludeID int64) error {
	q := s.q.Contact.WithContext(ctx).Where(s.q.Contact.CustomerID.Eq(customerID), s.q.Contact.Email.Eq(email))
	if excludeID > 0 {
		q = q.Where(s.q.Contact.ID.Neq(excludeID))
	}
	count, err := q.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return service.ErrContactEmailAlreadyExists
	}
	return nil
}
