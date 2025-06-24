package service

import (
	"context"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"

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
