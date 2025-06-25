package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"fmt"

	"github.com/casbin/casbin/v2"
)

type PermissionService struct {
	enforcer *casbin.Enforcer
}

func NewPermissionService(rm *resource.Manager) (*PermissionService, error) {
	casbinRes, err := resource.Get[*resource.CasbinResource](rm, resource.CasbinServiceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get casbin resource: %w", err)
	}
	return &PermissionService{enforcer: casbinRes.GetEnforcer()}, nil
}

// AddPermission 添加权限 (p, role, path, method)
func (s *PermissionService) AddPermission(ctx context.Context, req *dto.PermissionRequest) error {
	_, err := s.enforcer.AddPolicy(req.Role, req.Path, req.Method)
	if err != nil {
		return fmt.Errorf("failed to add permission policy: %w", err)
	}
	// 在分布式场景或需要立即生效时，需要调用 SavePolicy()
	// enforcer.SavePolicy()
	return nil
}

// RemovePermission 移除权限
func (s *PermissionService) RemovePermission(ctx context.Context, req *dto.PermissionRequest) error {
	_, err := s.enforcer.RemovePolicy(req.Role, req.Path, req.Method)
	if err != nil {
		return fmt.Errorf("failed to remove permission policy: %w", err)
	}
	// enforcer.SavePolicy()
	return nil
}

// ListPermissionsByRole 获取指定角色的所有权限
func (s *PermissionService) ListPermissionsByRole(ctx context.Context, role string) ([][]string, error) {
	permissions, err := s.enforcer.GetFilteredPolicy(0, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered policy: %w", err)
	}
	return permissions, nil
}

// AssignRoleToUser 给用户分配角色 (g, user_id, role)
func (s *PermissionService) AssignRoleToUser(ctx context.Context, req *dto.UserRoleRequest) error {
	// 使用 AddGroupingPolicy 来添加用户和角色之间的关联
	_, err := s.enforcer.AddGroupingPolicy(req.UserID, req.Role)
	if err != nil {
		return fmt.Errorf("failed to add grouping policy: %w", err)
	}
	return nil
}

// RemoveRoleFromUser 移除用户的角色
func (s *PermissionService) RemoveRoleFromUser(ctx context.Context, req *dto.UserRoleRequest) error {
	_, err := s.enforcer.RemoveGroupingPolicy(req.UserID, req.Role)
	if err != nil {
		return fmt.Errorf("failed to remove grouping policy: %w", err)
	}
	return nil
}

// GetRolesForUser 获取用户的所有角色
func (s *PermissionService) GetRolesForUser(ctx context.Context, userID string) ([]string, error) {
	roles, err := s.enforcer.GetRolesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}
	return roles, nil
}

// GetUsersForRole 获取拥有指定角色的所有用户
func (s *PermissionService) GetUsersForRole(ctx context.Context, role string) ([]string, error) {
	users, err := s.enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}
	return users, nil
}
