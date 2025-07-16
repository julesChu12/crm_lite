package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"

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

// Create
func (s *ContactService) Create(ctx context.Context, customerID int64, req *dto.ContactCreateRequest) (*dto.ContactResponse, error) {
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

	res, err := s.q.Contact.WithContext(ctx).Where(s.q.Contact.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
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
		return gorm.ErrRecordNotFound
	}
	return nil
}
