package service

import (
	"context"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
)

type AuthService struct {
	userQuery *query.Query
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		userQuery: query.Use(db),
	}
}

// Login 处理用户登录逻辑
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. 根据用户名查找用户
	user, err := s.userQuery.AdminUser.WithContext(ctx).Where(s.userQuery.AdminUser.Username.Eq(req.Username)).First()
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 2. 验证密码 - 注意字段名是 PasswordHash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	// 3. 生成JWT
	accessToken, refreshToken, err := utils.GenerateTokens(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 4. 返回响应
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 应当从配置中读取
	}, nil
}
