package service

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleService struct {
	q *query.Query
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{q: query.Use(db)}
}

func (s *RoleService) CreateRole(ctx context.Context, req *dto.RoleCreateRequest) (*dto.RoleResponse, error) {
	role := &model.Role{
		ID:          uuid.New().String(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsActive:    true,
	}
	if err := s.q.Role.WithContext(ctx).Create(role); err != nil {
		return nil, err
	}
	return s.toDTO(role), nil
}

func (s *RoleService) ListRoles(ctx context.Context) ([]*dto.RoleResponse, error) {
	roles, err := s.q.Role.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	res := make([]*dto.RoleResponse, 0, len(roles))
	for _, r := range roles {
		res = append(res, s.toDTO(r))
	}
	return res, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id string, req *dto.RoleUpdateRequest) error {
	r := s.q.Role
	updates := map[string]interface{}{
		"display_name": req.DisplayName,
		"description":  req.Description,
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	_, err := r.WithContext(ctx).Where(r.ID.Eq(id)).Updates(updates)
	return err
}

func (s *RoleService) DeleteRole(ctx context.Context, id string) error {
	r := s.q.Role
	_, err := r.WithContext(ctx).Where(r.ID.Eq(id)).Delete()
	return err
}

func (s *RoleService) toDTO(role *model.Role) *dto.RoleResponse {
	return &dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsActive:    role.IsActive,
	}
}
