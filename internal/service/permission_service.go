package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"fmt"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
)

type PermissionService struct {
	enforcer *casbin.Enforcer
	q        *query.Query
}

func NewPermissionService(rm *resource.Manager) (*PermissionService, error) {
	casbinRes, err := resource.Get[*resource.CasbinResource](rm, resource.CasbinServiceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get casbin resource: %w", err)
	}
	dbRes, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get db resource: %w", err)
	}

	return &PermissionService{
		enforcer: casbinRes.GetEnforcer(),
		q:        query.Use(dbRes.DB),
	}, nil
}

// AddPermission 添加权限 (p, role, path, method)
func (s *PermissionService) AddPermission(ctx context.Context, req *dto.PermissionRequest) error {
	_, err := s.enforcer.AddPolicy(req.Role, req.Path, req.Method)
	if err != nil {
		return fmt.Errorf("failed to add permission policy: %w", err)
	}
	// 策略会自动定期保存或在其他操作中保存，此处无需手动保存
	return nil
}

// RemovePermission 移除权限
func (s *PermissionService) RemovePermission(ctx context.Context, req *dto.PermissionRequest) error {
	removed, err := s.enforcer.RemovePolicy(req.Role, req.Path, req.Method)
	if err != nil {
		return fmt.Errorf("failed to remove permission policy: %w", err)
	}
	if !removed {
		return ErrPermissionNotFound
	}
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
	return s.q.Transaction(func(tx *query.Query) error {
		// 1. 根据 UUID 查找用户
		user, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(req.UserID)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrUserNotFound
			}
			return fmt.Errorf("failed to find user by uuid: %w", err)
		}

		// 2. 根据角色名称查找角色
		role, err := tx.Role.WithContext(ctx).Where(tx.Role.Name.Eq(req.Role)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrRoleNotFound
			}
			return fmt.Errorf("failed to find role by name: %w", err)
		}

		// 3. 在 admin_user_roles 中创建关联关系 (如果不存在)
		_, err = tx.AdminUserRole.WithContext(ctx).
			Where(tx.AdminUserRole.AdminUserID.Eq(user.ID), tx.AdminUserRole.RoleID.Eq(role.ID)).
			FirstOrCreate()
		if err != nil {
			return fmt.Errorf("failed to create user-role association: %w", err)
		}

		// 4. 更新 Casbin 策略
		// 使用 AddGroupingPolicy 来添加用户和角色之间的关联
		_, err = s.enforcer.AddGroupingPolicy(req.UserID, req.Role)
		if err != nil {
			return fmt.Errorf("failed to add grouping policy: %w", err)
		}

		return nil
	})
}

// RemoveRoleFromUser 移除用户的角色
func (s *PermissionService) RemoveRoleFromUser(ctx context.Context, req *dto.UserRoleRequest) error {
	return s.q.Transaction(func(tx *query.Query) error {
		// 1. 根据 UUID 查找用户
		user, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(req.UserID)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrUserNotFound
			}
			return fmt.Errorf("failed to find user by uuid: %w", err)
		}

		// 2. 根据角色名称查找角色
		role, err := tx.Role.WithContext(ctx).Where(tx.Role.Name.Eq(req.Role)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrRoleNotFound
			}
			return fmt.Errorf("failed to find role by name: %w", err)
		}

		// 3. 在 admin_user_roles 中删除关联关系
		result, err := tx.AdminUserRole.WithContext(ctx).
			Where(tx.AdminUserRole.AdminUserID.Eq(user.ID), tx.AdminUserRole.RoleID.Eq(role.ID)).
			Delete()
		if err != nil {
			return fmt.Errorf("failed to delete user-role association: %w", err)
		}
		if result.RowsAffected == 0 {
			return ErrUserRoleNotFound
		}

		// 4. 更新 Casbin 策略
		removed, err := s.enforcer.RemoveGroupingPolicy(req.UserID, req.Role)
		if err != nil {
			return fmt.Errorf("failed to remove grouping policy: %w", err)
		}
		if !removed {
			return ErrUserRoleNotFound // Casbin中未找到，可能数据不一致
		}
		return nil
	})
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
