package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	q        *query.Query
	resource *resource.Manager // 添加 resource manager
}

func NewUserService(resManager *resource.Manager) *UserService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for UserService: " + err.Error())
	}
	return &UserService{
		q:        query.Use(db.DB),
		resource: resManager, // 初始化 resource manager
	}
}

// GetUserByUUID 根据 UUID 获取单个用户详细信息
func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	// 1. 查询用户基本信息
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(uuid)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 2. 查询用户角色
	var roles []string
	err = s.q.Role.WithContext(ctx).
		Select(s.q.Role.Name).
		LeftJoin(s.q.AdminUserRole, s.q.AdminUserRole.RoleID.EqCol(s.q.Role.ID)).
		Where(s.q.AdminUserRole.AdminUserID.Eq(user.UUID)).
		Scan(&roles)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// 3. 组装成 DTO 返回
	return &dto.UserResponse{
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		RealName:  user.RealName,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		IsActive:  user.IsActive,
		Roles:     roles,
		CreatedAt: utils.FormatTime(user.CreatedAt),
	}, nil
}

// CreateUserByAdmin 管理员创建用户
func (s *UserService) CreateUserByAdmin(ctx context.Context, req *dto.AdminCreateUserRequest) (*dto.UserResponse, error) {
	// 1. 检查用户名、邮箱或手机号是否已存在（排除软删除的记录）
	var existingUser *model.AdminUser
	existingUser, _ = s.q.AdminUser.WithContext(ctx).
		Where(s.q.AdminUser.Username.Eq(req.Username)).
		Or(s.q.AdminUser.Email.Eq(req.Email)).
		Or(s.q.AdminUser.Phone.Eq(req.Phone)).
		Where(s.q.AdminUser.DeletedAt.IsNull()).
		First()
	if existingUser != nil {
		if existingUser.Username == req.Username {
			return nil, ErrUserAlreadyExists
		}
		if existingUser.Email == req.Email {
			return nil, ErrEmailAlreadyExists
		}
		if existingUser.Phone == req.Phone {
			return nil, ErrPhoneAlreadyExists
		}
	}

	// 2. 哈希密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. 在事务中创建用户并关联角色
	newUser := &model.AdminUser{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
		RealName:     req.RealName,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
	}
	if req.IsActive != nil {
		newUser.IsActive = *req.IsActive
	} else {
		newUser.IsActive = true // 默认激活
	}

	err = s.q.Transaction(func(tx *query.Query) error {
		// 创建用户
		if err := tx.AdminUser.WithContext(ctx).Create(newUser); err != nil {
			return err
		}

		// 关联角色
		if len(req.RoleIDs) > 0 {
			roles := make([]*model.AdminUserRole, len(req.RoleIDs))
			for i, roleID := range req.RoleIDs {
				roles[i] = &model.AdminUserRole{
					ID:          uuid.New().String(),
					AdminUserID: newUser.UUID,
					RoleID:      roleID,
				}
			}
			if err := tx.AdminUserRole.WithContext(ctx).Create(roles...); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user in transaction: %w", err)
	}

	return s.GetUserByUUID(ctx, newUser.UUID)
}

// ListUsers 获取用户列表（带分页和筛选）
func (s *UserService) ListUsers(ctx context.Context, req *dto.UserListRequest) (*dto.UserListResponse, error) {
	q := s.q.AdminUser.WithContext(ctx)

	// 构建查询条件
	if req.Username != "" {
		q = q.Where(s.q.AdminUser.Username.Like("%" + req.Username + "%"))
	}
	if req.Email != "" {
		q = q.Where(s.q.AdminUser.Email.Eq(req.Email))
	}
	if req.IsActive != nil {
		q = q.Where(s.q.AdminUser.IsActive.Is(*req.IsActive))
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// 分页查询
	users, err := q.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find()
	if err != nil {
		return nil, err
	}

	// 组装响应
	userResponses := make([]*dto.UserResponse, 0, len(users))
	for _, u := range users {
		// 复用 GetUserByUUID 以获取完整的用户信息（包含角色）
		fullUser, err := s.GetUserByUUID(ctx, u.UUID)
		if err != nil {
			// 如果某个用户查询失败，可以记录日志并跳过
			// log.Printf("Failed to get full details for user %s: %v", u.UUID, err)
			continue
		}
		userResponses = append(userResponses, fullUser)
	}

	return &dto.UserListResponse{
		Total: total,
		Users: userResponses,
	}, nil
}

// UpdateUserByAdmin 管理员更新用户信息
func (s *UserService) UpdateUserByAdmin(ctx context.Context, uuid_str string, req *dto.AdminUpdateUserRequest) (*dto.UserResponse, error) {
	err := s.q.Transaction(func(tx *query.Query) error {
		// 1. 更新用户信息
		updates := make(map[string]interface{})
		if req.Email != "" {
			updates["email"] = req.Email
		}
		if req.RealName != "" {
			updates["real_name"] = req.RealName
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.Avatar != "" {
			updates["avatar"] = req.Avatar
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}

		// 2. 检查 email 和 phone 的唯一性（排除软删除的记录）
		if req.Email != "" {
			count, err := tx.AdminUser.WithContext(ctx).
				Where(tx.AdminUser.Email.Eq(req.Email), tx.AdminUser.UUID.Neq(uuid_str)).
				Where(tx.AdminUser.DeletedAt.IsNull()).
				Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return ErrEmailAlreadyExists
			}
			updates["email"] = req.Email
		}
		if req.Phone != "" {
			count, err := tx.AdminUser.WithContext(ctx).
				Where(tx.AdminUser.Phone.Eq(req.Phone), tx.AdminUser.UUID.Neq(uuid_str)).
				Where(tx.AdminUser.DeletedAt.IsNull()).
				Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return ErrPhoneAlreadyExists
			}
			updates["phone"] = req.Phone
		}

		if len(updates) > 0 {
			if _, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(uuid_str)).Updates(updates); err != nil {
				return err
			}
		}

		// 2. 更新角色关联 - 改为调用 Casbin
		// 注意: Casbin 的操作不是事务性的，如果这里失败，前面的用户信息更新不会回滚。
		// 这是一个设计权衡，或者需要引入更复杂的补偿事务逻辑（如 two-phase commit）。
		// 当前我们接受这种权衡，因为角色分配失败的概率远低于用户信息更新失败。
		casbinRes, err := resource.Get[*resource.CasbinResource](s.resource, resource.CasbinServiceKey)
		if err != nil {
			return fmt.Errorf("failed to get casbin resource for role update: %w", err)
		}
		enforcer := casbinRes.GetEnforcer()

		// 先删除用户的所有旧角色
		if _, err := enforcer.DeleteRolesForUser(uuid_str); err != nil {
			return fmt.Errorf("failed to delete old roles for user: %w", err)
		}

		// 再添加新角色
		if len(req.RoleIDs) > 0 {
			// 将 role_id 转换为 role_name
			role_models, err := s.q.Role.WithContext(ctx).Where(s.q.Role.ID.In(req.RoleIDs...)).Find()
			if err != nil {
				return fmt.Errorf("failed to find roles by ids: %w", err)
			}
			if len(role_models) != len(req.RoleIDs) {
				return ErrRoleNotFound // 如果提供的某些RoleID无效
			}

			var roles_to_add []string
			for _, r_model := range role_models {
				roles_to_add = append(roles_to_add, r_model.Name)
			}

			if len(roles_to_add) > 0 {
				if _, err := enforcer.AddRolesForUser(uuid_str, roles_to_add); err != nil {
					return fmt.Errorf("failed to add new roles for user: %w", err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user in transaction: %w", err)
	}

	return s.GetUserByUUID(ctx, uuid_str)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, uuid_str string) error {
	// 在删除用户时，也应该清理其在 Casbin 中的角色和权限信息
	casbinRes, err := resource.Get[*resource.CasbinResource](s.resource, resource.CasbinServiceKey)
	if err != nil {
		return fmt.Errorf("failed to get casbin resource for user deletion: %w", err)
	}
	enforcer := casbinRes.GetEnforcer()

	// 开启事务
	return s.q.Transaction(func(tx *query.Query) error {
		// 1. 从数据库中删除用户
		if _, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(uuid_str)).Delete(); err != nil {
			return err
		}

		// 2. 从 Casbin 中删除该用户的所有角色关联 (g policies)
		if _, err := enforcer.DeleteRolesForUser(uuid_str); err != nil {
			// 注意：这里如果失败，数据库的用户记录已删除。
			// 这是一个需要权衡的地方。可以选择忽略这个错误，或者记录日志。
			log.Printf("Warn: failed to delete user roles from casbin for user %s, but user was deleted from db. manual cleanup may be needed. Error: %v", uuid_str, err)
		}

		// 3. (可选) 删除以该用户为主体的权限策略 (p policies)
		// 通常用户是通过角色获得权限的，直接赋予用户的权限较少。
		// 如果有这种场景，需要执行 enforcer.DeletePermissionsForUser(uuid_str)
		// 当前设计中，我们假设权限都与角色挂钩，所以这一步省略。

		return nil
	})
}
