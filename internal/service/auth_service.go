package service

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	q        *query.Query
	resource *resource.Manager
	opts     *config.Options
}

func NewAuthService(resManager *resource.Manager) *AuthService {
	// 从资源管理器获取数据库连接
	dbResource, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for AuthService: " + err.Error())
	}

	return &AuthService{
		q:        query.Use(dbResource.DB),
		resource: resManager,
		opts:     config.GetInstance(),
	}
}

// Logout 将JWT加入黑名单
func (s *AuthService) Logout(ctx context.Context, claims *utils.CustomClaims) error {
	cache, err := resource.Get[*resource.CacheResource](s.resource, resource.CacheServiceKey)
	if err != nil {
		return fmt.Errorf("failed to get cache resource: %w", err)
	}

	// 计算剩余过期时间
	expiresAt := claims.ExpiresAt.Time
	ttl := time.Until(expiresAt)

	// 如果 token 已经过期，则无需操作
	if ttl <= 0 {
		return nil
	}

	// JTI（JWT ID）是令牌的唯一标识符
	// 我们用 "jti:" 作为前缀，便于管理
	err = cache.Client.Set(ctx, "jti:"+claims.ID, "blacklisted", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}
	return nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	// 1. 解析 Refresh Token
	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 2. 检查黑名单（如果需要，登出时也可以将 RefreshToken 加入黑名单）
	// 此处简化，不检查 refreshToken 的黑名单

	// 3. 验证用户是否存在
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(claims.UserID)).First()
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 4. 生成新的 Access Token 和 Refresh Token
	// 为安全起见，通常刷新操作也会轮换 Refresh Token
	accessToken, refreshToken, err := utils.GenerateTokens(user.UUID, user.Username, claims.Roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// 5. 返回响应
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.opts.Auth.JWTOptions.AccessTokenExpire.Seconds()),
	}, nil
}

// ForgotPassword 处理忘记密码请求
func (s *AuthService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	log := logger.GetGlobalLogger()
	// 1. 查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Email.Eq(req.Email)).First()
	if err != nil {
		// 即使找不到用户，也返回 nil，避免泄露用户信息
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("Password reset requested for non-existent email", zap.String("email", req.Email))
			return nil
		}
		return err
	}

	// 2. 生成一个安全的重置令牌
	resetToken := uuid.New().String()
	cacheKey := "reset_token:" + resetToken

	// 3. 将令牌存入 Redis，并设置一个较短的过期时间（例如 15 分钟）
	cache, err := resource.Get[*resource.CacheResource](s.resource, resource.CacheServiceKey)
	if err != nil {
		return fmt.Errorf("failed to get cache resource: %w", err)
	}
	ttl := 15 * time.Minute
	err = cache.Client.Set(ctx, cacheKey, user.UUID, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// 4. 发送邮件（此处用日志模拟）
	// 在真实项目中，这里会调用邮件服务
	log.Info("Password reset token generated",
		zap.String("email", req.Email),
		zap.String("token", resetToken),
	)

	return nil
}

// ResetPassword 使用令牌重置密码
func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	// 1. 验证重置令牌
	cacheKey := "reset_token:" + req.Token
	cache, err := resource.Get[*resource.CacheResource](s.resource, resource.CacheServiceKey)
	if err != nil {
		return fmt.Errorf("failed to get cache resource: %w", err)
	}

	userID, err := cache.Client.Get(ctx, cacheKey).Result()
	if err != nil {
		return ErrInvalidToken // 令牌不存在或已过期
	}

	// 2. 哈希新密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 3. 更新数据库中的密码
	result, err := s.q.AdminUser.WithContext(ctx).
		Where(s.q.AdminUser.UUID.Eq(userID)).
		Update(s.q.AdminUser.PasswordHash, string(hashed))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	// 4. 删除已使用的令牌
	_ = cache.Client.Del(ctx, cacheKey).Err()

	return nil
}

// Login 处理用户登录逻辑
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. 根据用户名查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Username.Eq(req.Username)).First()
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

	// 3. 查询用户角色
	var roles []string
	q := s.q
	err = q.Role.WithContext(ctx).
		Select(q.Role.Name).
		LeftJoin(q.AdminUserRole, q.AdminUserRole.RoleID.EqCol(q.Role.ID)).
		Where(q.AdminUserRole.AdminUserID.Eq(user.UUID)).
		Scan(&roles)

	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// 4. 生成JWT
	accessToken, refreshToken, err := utils.GenerateTokens(user.UUID, user.Username, roles)
	if err != nil {
		return nil, err
	}

	// 5. 返回响应
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.opts.Auth.JWTOptions.AccessTokenExpire.Seconds()),
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	// 1. 检查用户名是否存在
	_, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Username.Eq(req.Username)).First()
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

	if err := s.q.AdminUser.WithContext(ctx).Create(user); err != nil {
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
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
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

	return s.q.AdminUser.WithContext(ctx).Save(user)
}

// GetProfile 获取当前登录用户的个人资料
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
	// 1. 查询用户基本信息
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 2. 查询用户角色
	var roles []string
	q := s.q
	err = q.Role.WithContext(ctx).
		Select(q.Role.Name).
		LeftJoin(q.AdminUserRole, q.AdminUserRole.RoleID.EqCol(q.Role.ID)).
		Where(q.AdminUserRole.AdminUserID.Eq(user.UUID)).
		Scan(&roles)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// 3. 组装 DTO
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

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	// 1. 查询用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		return ErrUserNotFound
	}

	// 2. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// 3. 哈希新密码并更新
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	result, err := s.q.AdminUser.WithContext(ctx).
		Where(s.q.AdminUser.UUID.Eq(userID)).
		Update(s.q.AdminUser.PasswordHash, string(hashed))

	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
