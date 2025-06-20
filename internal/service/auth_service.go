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

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
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
		fmt.Printf("User not found: %v\n", err)
		return nil, ErrUserNotFound
	}

	fmt.Printf("Login attempt details:\n")
	fmt.Printf("- Username: %s\n", user.Username)
	fmt.Printf("- Stored Hash: %s\n", user.PasswordHash)
	fmt.Printf("- Input Password: %s\n", req.Password)
	fmt.Printf("- Hash Format: %s (algorithm), %s (cost), %s (salt+hash)\n",
		user.PasswordHash[0:4], // $2a$
		user.PasswordHash[4:6], // 10
		user.PasswordHash[6:],  // 剩余部分
	)

	// 2. 验证密码 - 注意字段名是 PasswordHash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		fmt.Printf("Password verification failed: %v\n", err)
		return nil, ErrInvalidPassword
	}

	fmt.Printf("Password verification successful!\n")

	// 3. 生成JWT
	accessToken, refreshToken, err := utils.GenerateTokens(user.UUID, user.Username)
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

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	// 1. 检查用户名是否存在
	_, err := s.userQuery.AdminUser.WithContext(ctx).Where(s.userQuery.AdminUser.Username.Eq(req.Username)).First()
	if err == nil {
		return ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err // 其他数据库错误
	}

	// 2. 生成密码哈希
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 3. 创建用户
	user := &model.AdminUser{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
		RealName:     req.RealName,
		IsActive:     true,
	}

	if err := s.userQuery.AdminUser.WithContext(ctx).Create(user); err != nil {
		return err
	}
	return nil
}

// UpdateProfile 更新用户信息（示例）
func (s *AuthService) UpdateProfile(ctx context.Context, req *dto.UpdateUserRequest) error {
	// 1. 获取当前用户ID（应由上游中间件注入到 ctx）
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == "" {
		return ErrUserNotFound
	}

	// 2. 查询用户
	user, err := s.userQuery.AdminUser.WithContext(ctx).Where(s.userQuery.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 3. 更新字段
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	return s.userQuery.AdminUser.WithContext(ctx).Save(user)
}
