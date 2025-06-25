package service

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"errors"

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

func (s *RoleService) GetRoleByID(ctx context.Context, id string) (*dto.RoleResponse, error) {
	role, err := s.q.Role.WithContext(ctx).Where(s.q.Role.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return s.toDTO(role), nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id string, req *dto.RoleUpdateRequest) (*dto.RoleResponse, error) {
	r := s.q.Role
	updates := make(map[string]interface{})
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	result, err := r.WithContext(ctx).Where(r.ID.Eq(id)).Updates(updates)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, ErrRoleNotFound
	}

	// 返回更新后的角色信息
	return s.GetRoleByID(ctx, id)
}

func (s *RoleService) DeleteRole(ctx context.Context, id string) error {
	r := s.q.Role
	result, err := r.WithContext(ctx).Where(r.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrRoleNotFound
	}
	return nil
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
