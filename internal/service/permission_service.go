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
