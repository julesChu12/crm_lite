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

// IAuthRepo 定义了认证相关的数据库操作
type IAuthRepo interface {
	FindByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	FindByEmail(ctx context.Context, email string) (*model.AdminUser, error)
	FindByUUID(ctx context.Context, uuid string) (*model.AdminUser, error)
	Create(ctx context.Context, user *model.AdminUser) error
	UpdatePassword(ctx context.Context, userID, hashedPassword string) (int64, error)
	Save(ctx context.Context, user *model.AdminUser) error
}

// IAuthCache 定义了认证相关的缓存操作
type IAuthCache interface {
	SetBlacklist(ctx context.Context, jti string, ttl time.Duration) error
	SetResetToken(ctx context.Context, token, userID string, ttl time.Duration) error
	GetResetToken(ctx context.Context, token string) (string, error)
	DeleteResetToken(ctx context.Context, token string) error
}

// authRepo 实现了 IAuthRepo
type authRepo struct {
	q *query.Query
}

func NewAuthRepo(db *gorm.DB) IAuthRepo {
	return &authRepo{q: query.Use(db)}
}

func (r *authRepo) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	return r.q.AdminUser.WithContext(ctx).Where(r.q.AdminUser.Username.Eq(username)).First()
}
func (r *authRepo) FindByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	return r.q.AdminUser.WithContext(ctx).Where(r.q.AdminUser.Email.Eq(email)).First()
}
func (r *authRepo) FindByUUID(ctx context.Context, uuid string) (*model.AdminUser, error) {
	return r.q.AdminUser.WithContext(ctx).Where(r.q.AdminUser.UUID.Eq(uuid)).First()
}
func (r *authRepo) Create(ctx context.Context, user *model.AdminUser) error {
	return r.q.AdminUser.WithContext(ctx).Create(user)
}
func (r *authRepo) UpdatePassword(ctx context.Context, userID, hashedPassword string) (int64, error) {
	result, err := r.q.AdminUser.WithContext(ctx).Where(r.q.AdminUser.UUID.Eq(userID)).Update(r.q.AdminUser.PasswordHash, hashedPassword)
	return result.RowsAffected, err
}
func (r *authRepo) Save(ctx context.Context, user *model.AdminUser) error {
	return r.q.AdminUser.WithContext(ctx).Save(user)
}

// authCache 实现了 IAuthCache
type authCache struct {
	client *resource.CacheResource
}

func NewAuthCache(client *resource.CacheResource) IAuthCache {
	return &authCache{client: client}
}
func (c *authCache) SetBlacklist(ctx context.Context, jti string, ttl time.Duration) error {
	return c.client.Client.Set(ctx, "jti:"+jti, "blacklisted", ttl).Err()
}
func (c *authCache) SetResetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return c.client.Client.Set(ctx, "reset_token:"+token, userID, ttl).Err()
}
func (c *authCache) GetResetToken(ctx context.Context, token string) (string, error) {
	return c.client.Client.Get(ctx, "reset_token:"+token).Result()
}
func (c *authCache) DeleteResetToken(ctx context.Context, token string) error {
	return c.client.Client.Del(ctx, "reset_token:"+token).Err()
}

type AuthService struct {
	repo     IAuthRepo
	cache    IAuthCache
	emailSvc IEmailService // 假设 EmailService 也有一个接口
	casbin   *resource.CasbinResource
	opts     *config.Options
}

func NewAuthService(repo IAuthRepo, cache IAuthCache, emailSvc IEmailService, casbin *resource.CasbinResource, opts *config.Options) *AuthService {
	return &AuthService{
		repo:     repo,
		cache:    cache,
		emailSvc: emailSvc,
		casbin:   casbin,
		opts:     opts,
	}
}

// Logout 将JWT加入黑名单
func (s *AuthService) Logout(ctx context.Context, claims *utils.CustomClaims) error {
	expiresAt := claims.ExpiresAt.Time
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return s.cache.SetBlacklist(ctx, claims.ID, ttl)
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	claims, err := utils.ParseToken(req.RefreshToken, s.opts.Auth.JWTOptions)
	if err != nil {
		return nil, ErrInvalidToken
	}
	user, err := s.repo.FindByUUID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	enforcer := s.casbin.GetEnforcer()
	roles, err := enforcer.GetRolesForUser(user.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles from casbin: %w", err)
	}

	accessToken, refreshToken, err := utils.GenerateTokens(user.UUID, user.Username, roles)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.opts.Auth.JWTOptions.AccessTokenExpire.Seconds()),
	}, nil
}

// ForgotPassword 处理忘记密码请求
func (s *AuthService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("Password reset requested for non-existent email", zap.String("email", req.Email))
			return nil
		}
		return err
	}

	resetToken := uuid.New().String()
	ttl := 15 * time.Minute
	if err := s.cache.SetResetToken(ctx, resetToken, user.UUID, ttl); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	if s.emailSvc == nil {
		logger.Error("Email service is not configured, cannot send password reset email.")
		return nil
	}

	subject := "Reset Your Password"
	body := fmt.Sprintf(`<p>Your password reset token is: <b>%s</b></p>`, resetToken)

	if err := s.emailSvc.SendMail(req.Email, subject, body); err != nil {
		logger.Error("Failed to send password reset email", zap.Error(err), zap.String("email", req.Email))
		return nil
	}
	return nil
}

// ResetPassword 使用令牌重置密码
func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	userID, err := s.cache.GetResetToken(ctx, req.Token)
	if err != nil {
		return ErrInvalidToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	rowsAffected, err := s.repo.UpdatePassword(ctx, userID, string(hashed))
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	_ = s.cache.DeleteResetToken(ctx, req.Token)
	return nil
}

// Login 处理用户登录逻辑
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	enforcer := s.casbin.GetEnforcer()
	roles, err := enforcer.GetRolesForUser(user.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles from casbin: %w", err)
	}

	accessToken, refreshToken, err := utils.GenerateTokens(user.UUID, user.Username, roles)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.opts.Auth.JWTOptions.AccessTokenExpire.Seconds()),
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	_, err := s.repo.FindByUsername(ctx, req.Username)
	if err == nil {
		return ErrUserAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.AdminUser{
		UUID:         uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
		RealName:     req.RealName,
		IsActive:     true,
	}

	return s.repo.Create(ctx, user)
}

// UpdateProfile 更新用户信息
func (s *AuthService) UpdateProfile(ctx context.Context, req *dto.UpdateUserRequest) error {
	userID, ok := utils.GetUserID(ctx)
	if !ok || userID == "" {
		return ErrUserNotFound
	}

	user, err := s.repo.FindByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	return s.repo.Save(ctx, user)
}

// GetProfile 获取当前登录用户的个人资料
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	enforcer := s.casbin.GetEnforcer()
	roles, err := enforcer.GetRolesForUser(user.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles from casbin: %w", err)
	}

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
	user, err := s.repo.FindByUUID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrInvalidPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	rowsAffected, err := s.repo.UpdatePassword(ctx, userID, string(hashed))
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
