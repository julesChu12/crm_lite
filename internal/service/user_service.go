package service

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	q *query.Query
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		q: query.Use(db),
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
	// 1. 检查用户名或邮箱是否已存在
	user, _ := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Username.Eq(req.Username)).Or(s.q.AdminUser.Email.Eq(req.Email)).First()
	if user != nil {
		return nil, ErrUserAlreadyExists
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

		if len(updates) > 0 {
			if _, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(uuid_str)).Updates(updates); err != nil {
				return err
			}
		}

		// 2. 更新角色关联
		// 先删除旧的关联
		if _, err := tx.AdminUserRole.WithContext(ctx).Where(tx.AdminUserRole.AdminUserID.Eq(uuid_str)).Delete(); err != nil {
			return err
		}
		// 再创建新的关联
		if len(req.RoleIDs) > 0 {
			roles := make([]*model.AdminUserRole, len(req.RoleIDs))
			for i, roleID := range req.RoleIDs {
				roles[i] = &model.AdminUserRole{
					ID:          uuid.New().String(),
					AdminUserID: uuid_str,
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
		return nil, fmt.Errorf("failed to update user in transaction: %w", err)
	}

	return s.GetUserByUUID(ctx, uuid_str)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, uuid string) error {
	// 推荐使用事务确保原子性
	return s.q.Transaction(func(tx *query.Query) error {
		// 1. 删除用户与角色的关联
		if _, err := tx.AdminUserRole.WithContext(ctx).Where(tx.AdminUserRole.AdminUserID.Eq(uuid)).Delete(); err != nil {
			return err
		}
		// 2. 删除用户自身
		if _, err := tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.UUID.Eq(uuid)).Delete(); err != nil {
			return err
		}
		return nil
	})
}
