package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

type RoleService struct {
	q        *query.Query
	resource *resource.Manager
}

func NewRoleService(resManager *resource.Manager) *RoleService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for RoleService: " + err.Error())
	}
	return &RoleService{
		q:        query.Use(db.DB),
		resource: resManager,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, req *dto.RoleCreateRequest) (*dto.RoleResponse, error) {
	// 检查 name 唯一性（排除软删除的记录）
	count, err := s.q.Role.WithContext(ctx).
		Where(s.q.Role.Name.Eq(req.Name)).
		Where(s.q.Role.DeletedAt.IsNull()).
		Count()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrRoleNameAlreadyExists
	}

	role := &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsActive:    req.IsActive,
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
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return nil, ErrRoleNotFound
	}
	role, err := s.q.Role.WithContext(ctx).Where(s.q.Role.ID.Eq(idNum)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return s.toDTO(role), nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id string, req *dto.RoleUpdateRequest) (*dto.RoleResponse, error) {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return nil, ErrRoleNotFound
	}
	r := s.q.Role
	updates := make(map[string]interface{})

	// 检查 display_name 是否需要更新以及唯一性（虽然数据库中没有强制唯一约束，但业务上建议保持唯一）
	if req.DisplayName != "" {
		// 检查 display_name 是否已存在（排除当前角色）
		count, err := r.WithContext(ctx).Where(r.DisplayName.Eq(req.DisplayName), r.ID.Neq(idNum)).Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrRoleNameAlreadyExists
		}
		updates["display_name"] = req.DisplayName
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		result, err := r.WithContext(ctx).Where(r.ID.Eq(idNum)).Updates(updates)
		if err != nil {
			return nil, err
		}
		if result.RowsAffected == 0 {
			return nil, ErrRoleNotFound
		}
	}

	// 返回更新后的角色信息
	return s.GetRoleByID(ctx, id)
}

func (s *RoleService) DeleteRole(ctx context.Context, id string) error {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return ErrRoleNotFound
	}
	r := s.q.Role
	result, err := r.WithContext(ctx).Where(r.ID.Eq(idNum)).Delete()
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
