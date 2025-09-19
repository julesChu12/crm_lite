package service

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockAuthRepo is a mock type for the IAuthRepo interface
type MockAuthRepo struct{ mock.Mock }

func (m *MockAuthRepo) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	args := m.Called(ctx, username)
	var r0 *model.AdminUser
	if args.Get(0) != nil {
		r0 = args.Get(0).(*model.AdminUser)
	}
	return r0, args.Error(1)
}
func (m *MockAuthRepo) Create(ctx context.Context, user *model.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// ... Add other mocked methods as needed ...
func (m *MockAuthRepo) FindByEmail(ctx context.Context, email string) (*model.AdminUser, error) {
	args := m.Called(ctx, email)
	var r0 *model.AdminUser
	if args.Get(0) != nil {
		r0 = args.Get(0).(*model.AdminUser)
	}
	return r0, args.Error(1)
}
func (m *MockAuthRepo) FindByUUID(ctx context.Context, uuid string) (*model.AdminUser, error) {
	args := m.Called(ctx, uuid)
	var r0 *model.AdminUser
	if args.Get(0) != nil {
		r0 = args.Get(0).(*model.AdminUser)
	}
	return r0, args.Error(1)
}
func (m *MockAuthRepo) UpdatePassword(ctx context.Context, userID, hashedPassword string) (int64, error) {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockAuthRepo) Save(ctx context.Context, user *model.AdminUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockAuthCache is a mock type for the IAuthCache interface
type MockAuthCache struct{ mock.Mock }

func (m *MockAuthCache) SetBlacklist(ctx context.Context, jti string, ttl time.Duration) error {
	args := m.Called(ctx, jti, ttl)
	return args.Error(0)
}
func (m *MockAuthCache) SetResetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	args := m.Called(ctx, token, userID, ttl)
	return args.Error(0)
}
func (m *MockAuthCache) GetResetToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}
func (m *MockAuthCache) DeleteResetToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// MockEmailService is a mock type for the IEmailService interface
type MockEmailService struct{ mock.Mock }

func (m *MockEmailService) SendMail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

// SetupAuthServiceForTest is a helper to setup AuthService with mocks
func setupAuthServiceForTest() (*AuthService, *MockAuthRepo, *MockAuthCache, *MockEmailService) {
	mockRepo := new(MockAuthRepo)
	mockCache := new(MockAuthCache)
	mockEmailSvc := new(MockEmailService)

	// Setup in-memory casbin for testing
	m, _ := casbinmodel.NewModelFromString(`
	[request_definition]
	r = sub, obj, act
	
	[policy_definition]
	p = sub, obj, act
	
	[role_definition]
	g = _, _
	
	[policy_effect]
	e = some(where (p.eft == allow))
	
	[matchers]
	m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
	`)
	// 使用内存 SQLite DB 以便 Casbin Adapter 正常工作
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	adapter, _ := gormadapter.NewAdapterByDB(db)
	enforcer, _ := casbin.NewEnforcer(m, adapter)

	resManager := resource.NewManager()
	casbinRes := resource.NewCasbinResource(resManager, config.RbacOptions{
		ModelFile: "../../bootstrap/model.conf",
	})
	casbinRes.Enforcer = enforcer
	_ = resManager.Register(resource.CasbinServiceKey, casbinRes)

	opts := &config.Options{Auth: config.AuthOptions{JWTOptions: config.JWTOptions{
		AccessTokenExpire: time.Hour,
		Secret:            "test-secret", // Use a consistent secret for tests
	}}}
	// Make sure the config singleton is updated for utils.ParseToken
	config.SetInstanceForTest(opts)

	authService := NewAuthService(mockRepo, mockCache, mockEmailSvc, casbinRes, opts)
	return authService, mockRepo, mockCache, mockEmailSvc
}

func TestLogin_Success(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &model.AdminUser{
		UUID:         "user-uuid-123",
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}

	// Grant a role to the user for testing
	_, _ = authService.casbin.GetEnforcer().AddRoleForUser(user.UUID, "test_role")

	// Expectations
	mockRepo.On("FindByUsername", ctx, "testuser").Return(user, nil)

	// Execute
	req := &dto.LoginRequest{Username: "testuser", Password: password}
	resp, err := authService.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()

	mockRepo.On("FindByUsername", ctx, "unknownuser").Return(nil, gorm.ErrRecordNotFound)

	req := &dto.LoginRequest{Username: "unknownuser", Password: "password"}
	_, err := authService.Login(ctx, req)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserNotFound))
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &model.AdminUser{Username: "testuser", PasswordHash: string(hashedPassword)}

	mockRepo.On("FindByUsername", ctx, "testuser").Return(user, nil)

	req := &dto.LoginRequest{Username: "testuser", Password: "wrongpassword"}
	_, err := authService.Login(ctx, req)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPassword))
	mockRepo.AssertExpectations(t)
}

func TestRegister_Success(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()
	req := &dto.RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
	}

	// Expectations
	mockRepo.On("FindByUsername", ctx, "newuser").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.AdminUser")).Return(nil)

	// Execute
	err := authService.Register(ctx, req)

	// Assert
	assert.NoError(t, err)

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()
	existingUser := &model.AdminUser{
		Username: "existinguser",
		Email:    "existinguser@example.com",
	}
	req := &dto.RegisterRequest{
		Username: "existinguser",
		Email:    "existinguser@example.com",
		Password: "password123",
	}

	// Expectations
	mockRepo.On("FindByUsername", ctx, "existinguser").Return(existingUser, nil)

	// Execute
	err := authService.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserAlreadyExists))

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestLogout_Success(t *testing.T) {
	// Setup
	authService, _, mockCache, _ := setupAuthServiceForTest()
	ctx := context.Background()
	claims := &utils.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "test-jti",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// Expectations
	mockCache.On("SetBlacklist", ctx, "test-jti", mock.AnythingOfType("time.Duration")).Return(nil)

	// Execute
	err := authService.Logout(ctx, claims)

	// Assert
	assert.NoError(t, err)

	// Verify
	mockCache.AssertExpectations(t)
}

func TestRefreshToken_Success(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()
	userID := "user-uuid-123"
	username := "testuser"

	// Mock user and roles
	user := &model.AdminUser{UUID: userID, Username: username}
	_, _ = authService.casbin.GetEnforcer().AddRoleForUser(userID, "test_role")

	// Generate a valid refresh token for testing
	// In a real scenario, this would be generated by your app
	// For testing, we can manually create claims and sign a token
	claims := &utils.CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// IMPORTANT: use the same secret key as in your token generation logic
	refreshToken, err := token.SignedString([]byte("test-secret")) // Replace with your actual secret key
	assert.NoError(t, err)

	// Expectations
	mockRepo.On("FindByUUID", ctx, userID).Return(user, nil)

	// Execute
	req := &dto.RefreshTokenRequest{RefreshToken: refreshToken}
	resp, err := authService.RefreshToken(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_CasbinError(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()
	userID := "user-uuid-casbin-error"
	user := &model.AdminUser{UUID: userID}
	mockRepo.On("FindByUUID", ctx, userID).Return(user, nil)

	// To trigger a predictable error, set the enforcer to nil.
	// This will cause a panic, which is a form of error.
	authService.casbin.Enforcer = nil

	// Generate a valid token
	claims := &utils.CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	// Execute and expect a panic/error
	assert.Panics(t, func() {
		_, _ = authService.RefreshToken(ctx, &dto.RefreshTokenRequest{RefreshToken: refreshToken})
	}, "The code did not panic as expected")

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	// Setup
	authService, _, _, _ := setupAuthServiceForTest()
	ctx := context.Background()

	// Execute
	req := &dto.RefreshTokenRequest{RefreshToken: "invalid-token"}
	_, err := authService.RefreshToken(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidToken))
}

func TestForgotPassword_Success(t *testing.T) {
	// Setup
	authService, mockRepo, mockCache, mockEmailSvc := setupAuthServiceForTest()
	ctx := context.Background()
	req := &dto.ForgotPasswordRequest{Email: "user@example.com"}
	user := &model.AdminUser{UUID: "user-uuid", Email: "user@example.com"}

	// Expectations
	mockRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
	mockCache.On("SetResetToken", ctx, mock.AnythingOfType("string"), user.UUID, 15*time.Minute).Return(nil)
	mockEmailSvc.On("SendMail", req.Email, "Reset Your Password", mock.AnythingOfType("string")).Return(nil)

	// Execute
	err := authService.ForgotPassword(ctx, req)

	// Assert
	assert.NoError(t, err)

	// Verify
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockEmailSvc.AssertExpectations(t)
}

func TestForgotPassword_EmailNotFound(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	ctx := context.Background()
	req := &dto.ForgotPasswordRequest{Email: "notfound@example.com"}

	// Expectations
	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	err := authService.ForgotPassword(ctx, req)

	// Assert
	assert.NoError(t, err) // Should not return an error to prevent email enumeration

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestResetPassword_Success(t *testing.T) {
	// Setup
	authService, mockRepo, mockCache, _ := setupAuthServiceForTest()
	ctx := context.Background()
	userID := "user-uuid-for-reset"
	token := "valid-reset-token"
	req := &dto.ResetPasswordRequest{
		Token:       token,
		NewPassword: "newStrongPassword",
	}

	// Expectations
	mockCache.On("GetResetToken", ctx, token).Return(userID, nil)
	mockRepo.On("UpdatePassword", ctx, userID, mock.AnythingOfType("string")).Return(int64(1), nil)
	mockCache.On("DeleteResetToken", ctx, token).Return(nil)

	// Execute
	err := authService.ResetPassword(ctx, req)

	// Assert
	assert.NoError(t, err)

	// Verify
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestResetPassword_InvalidToken(t *testing.T) {
	// Setup
	authService, _, mockCache, _ := setupAuthServiceForTest()
	ctx := context.Background()
	token := "invalid-token"
	req := &dto.ResetPasswordRequest{Token: token, NewPassword: "password"}

	// Expectations
	mockCache.On("GetResetToken", ctx, token).Return("", errors.New("token not found"))

	// Execute
	err := authService.ResetPassword(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidToken))

	// Verify
	mockCache.AssertExpectations(t)
}

func TestResetPassword_UserNotFound(t *testing.T) {
	// Setup
	authService, mockRepo, mockCache, _ := setupAuthServiceForTest()
	ctx := context.Background()
	userID := "user-does-not-exist"
	token := "valid-token-for-non-existent-user"
	req := &dto.ResetPasswordRequest{Token: token, NewPassword: "password"}

	// Expectations
	mockCache.On("GetResetToken", ctx, token).Return(userID, nil)
	mockRepo.On("UpdatePassword", ctx, userID, mock.AnythingOfType("string")).Return(int64(0), nil) // 0 rows affected
	mockCache.On("DeleteResetToken", ctx, token).Return(nil)

	// Execute
	err := authService.ResetPassword(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrUserNotFound))

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_Success(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	userID := "user-to-update"
	ctx := utils.WithUser(context.Background(), userID, "testuser")
	req := &dto.UpdateUserRequest{
		RealName: "Updated Name",
		Phone:    "1234567890",
		Avatar:   "new-avatar-url",
	}
	user := &model.AdminUser{UUID: userID}

	// Expectations
	mockRepo.On("FindByUUID", ctx, userID).Return(user, nil)
	mockRepo.On("Save", ctx, user).Return(nil)

	// Execute
	err := authService.UpdateProfile(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", user.RealName)
	assert.Equal(t, "1234567890", user.Phone)
	assert.Equal(t, "new-avatar-url", user.Avatar)

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_Success(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	userID := "profile-user-id"
	ctx := context.Background()
	user := &model.AdminUser{
		UUID:     userID,
		Username: "profileuser",
		Email:    "profile@example.com",
	}
	roles := []string{"viewer"}

	// Expectations
	mockRepo.On("FindByUUID", ctx, userID).Return(user, nil)
	_, _ = authService.casbin.GetEnforcer().AddRoleForUser(userID, roles[0])

	// Execute
	resp, err := authService.GetProfile(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userID, resp.UUID)
	assert.Equal(t, "profileuser", resp.Username)
	assert.Equal(t, roles, resp.Roles)

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestChangePassword_Success(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	userID := "user-changing-password"
	ctx := context.Background()
	oldPassword := "oldPassword123"
	newPassword := "newPassword456"
	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	user := &model.AdminUser{UUID: userID, PasswordHash: string(hashedOldPassword)}
	req := &dto.ChangePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	// Expectations
	mockRepo.On("FindByUUID", ctx, userID).Return(user, nil)
	mockRepo.On("UpdatePassword", ctx, userID, mock.AnythingOfType("string")).Return(int64(1), nil)

	// Execute
	err := authService.ChangePassword(ctx, userID, req)

	// Assert
	assert.NoError(t, err)

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestChangePassword_InvalidOldPassword(t *testing.T) {
	// Setup
	authService, mockRepo, _, _ := setupAuthServiceForTest()
	userID := "user-invalid-old-pass"
	ctx := context.Background()
	correctOldPassword := "correctOldPassword"
	wrongOldPassword := "wrongOldPassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctOldPassword), bcrypt.DefaultCost)
	user := &model.AdminUser{UUID: userID, PasswordHash: string(hashedPassword)}
	req := &dto.ChangePasswordRequest{
		OldPassword: wrongOldPassword,
		NewPassword: "new-password",
	}

	// Expectations
	mockRepo.On("FindByUUID", ctx, userID).Return(user, nil)

	// Execute
	err := authService.ChangePassword(ctx, userID, req)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPassword))

	// Verify
	mockRepo.AssertExpectations(t)
}
