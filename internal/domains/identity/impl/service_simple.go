package impl

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"crm_lite/internal/common"
	"crm_lite/internal/core/config"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/identity"
	"crm_lite/pkg/utils"

	"github.com/casbin/casbin/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SimpleIdentityService Identity域简化实现
// 先实现基本功能，保证编译通过
type SimpleIdentityService struct {
	db       *gorm.DB
	q        *query.Query
	enforcer *casbin.Enforcer
}

// 错误常量
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrInvalidPassword   = errors.New("密码错误")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrInvalidToken      = errors.New("无效令牌")
)

// Login 实现控制器接口 - 登录
func (s *SimpleIdentityService) Login(ctx context.Context, req *identity.LoginRequest) (*identity.LoginResponse, error) {
	// 查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Username.Eq(req.Username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, ErrUserNotFound
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := utils.GenerateTokens(user.UUID, user.Username, []string{"user"})
	if err != nil {
		return nil, err
	}

	// 获取用户角色（简化实现）
	roles := []identity.Role{
		{
			ID:          1,
			Name:        "user",
			DisplayName: "普通用户",
			IsActive:    true,
		},
	}

	// 转换为域模型
	userModel := &identity.User{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.RealName,
		Status:    "active",
		Roles:     roles,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	return &identity.LoginResponse{
		User:         userModel,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 3600, // 24小时
	}, nil
}

// Register 实现控制器接口 - 注册
func (s *SimpleIdentityService) Register(ctx context.Context, req *identity.RegisterRequest) error {
	// 检查用户名是否已存在
	existingUser, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Username.Eq(req.Username)).First()
	if err == nil && existingUser != nil {
		return ErrUserAlreadyExists
	}

	// 检查邮箱是否已存在
	existingUser, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Email.Eq(req.Email)).First()
	if err == nil && existingUser != nil {
		return ErrUserAlreadyExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建用户
	user := &model.AdminUser{
		UUID:         fmt.Sprintf("user-%d", time.Now().Unix()),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		RealName:     req.RealName,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.q.AdminUser.WithContext(ctx).Create(user); err != nil {
		return err
	}

	return nil
}

// UpdateProfile 实现控制器接口 - 更新资料
func (s *SimpleIdentityService) UpdateProfile(ctx context.Context, userID int64, req *identity.UpdateProfileRequest) error {
	// 查找用户
	_, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.ID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	updates["updated_at"] = time.Now()

	// 执行更新
	_, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.ID.Eq(userID)).Updates(updates)
	if err != nil {
		return err
	}

	return nil
}

// NewSimpleIdentityService 创建简化Identity服务实现
func NewSimpleIdentityService(db *gorm.DB, enforcer *casbin.Enforcer) *SimpleIdentityService {
	return &SimpleIdentityService{
		db:       db,
		q:        query.Use(db),
		enforcer: enforcer,
	}
}

// ===== AuthService 接口实现 =====

// Authenticate 用户认证
func (s *SimpleIdentityService) Authenticate(ctx context.Context, req identity.AuthenticateRequest) (*identity.AuthenticateResponse, error) {
	// 查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Username.Eq(req.Username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户名或密码错误")
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户名或密码错误")
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户已被禁用")
	}

	// 获取用户角色
	roles, err := s.enforcer.GetRolesForUser(user.UUID)
	if err != nil {
		roles = []string{"user"} // 默认角色
	}

	// 生成JWT令牌
	accessToken, refreshToken, err := utils.GenerateTokens(user.UUID, user.Username, roles)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	return &identity.AuthenticateResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		UserID:       user.UUID,
		Username:     user.Username,
		Roles:        roles,
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *SimpleIdentityService) RefreshToken(ctx context.Context, refreshToken string) (*identity.AuthenticateResponse, error) {
	// 解析refresh token获取claims
	opts := config.GetInstance().Auth.JWTOptions
	claims, err := utils.ParseToken(refreshToken, opts)
	if err != nil {
		return nil, common.NewBusinessError(common.ErrCodeUnauthorized, "无效的刷新令牌")
	}

	// 查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(claims.UserID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return nil, err
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户已被禁用")
	}

	// 获取用户角色
	roles, err := s.enforcer.GetRolesForUser(user.UUID)
	if err != nil {
		roles = []string{"user"} // 默认角色
	}

	// 生成新的JWT令牌
	accessToken, newRefreshToken, err := utils.GenerateTokens(user.UUID, user.Username, roles)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	return &identity.AuthenticateResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		UserID:       user.UUID,
		Username:     user.Username,
		Roles:        roles,
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

// Logout 用户登出
func (s *SimpleIdentityService) Logout(ctx context.Context, token string) error {
	return nil
}

// ValidateToken 验证令牌有效性
func (s *SimpleIdentityService) ValidateToken(ctx context.Context, token string) (*identity.User, error) {
	// 解析token获取claims
	opts := config.GetInstance().Auth.JWTOptions
	claims, err := utils.ParseToken(token, opts)
	if err != nil {
		return nil, common.NewBusinessError(common.ErrCodeUnauthorized, "无效的访问令牌")
	}

	// 查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(claims.UserID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return nil, err
	}

	// 检查用户状态
	if !user.IsActive {
		return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户已被禁用")
	}

	// 转换为域模型
	userModel := &identity.User{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.RealName,
		Status:    "active",
		Roles:     []identity.Role{}, // 简化实现，不包含详细角色信息
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	return userModel, nil
}

// ResetPassword 重置密码
func (s *SimpleIdentityService) ResetPassword(ctx context.Context, email string) error {
	// 查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Email.Eq(email)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 为安全起见，不明确提示用户是否存在
			return nil
		}
		return err
	}

	// 生成重置令牌（简化实现，实际应该存储到缓存）
	resetToken := fmt.Sprintf("reset-%d", time.Now().Unix())

	// 这里应该发送邮件，但简化实现只记录日志
	fmt.Printf("Password reset token for user %s: %s\n", user.Username, resetToken)

	return nil
}

// ConfirmPasswordReset 确认密码重置
func (s *SimpleIdentityService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	// 简化实现，实际应该验证token有效性
	if len(token) < 10 {
		return common.NewBusinessError(common.ErrCodeInvalidParam, "无效的重置令牌")
	}

	// 这里应该从缓存中获取用户ID，简化实现
	// 实际应该根据token查找对应的用户
	return fmt.Errorf("密码重置确认功能需要完善实现")
}

// ===== UserService 接口实现 =====

// CreateUser 创建用户
func (s *SimpleIdentityService) CreateUser(ctx context.Context, req identity.CreateUserRequest) (*identity.User, error) {
	// 检查用户名是否已存在
	existingUser, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Username.Eq(req.Username)).First()
	if err == nil && existingUser != nil {
		return nil, common.NewBusinessError(common.ErrCodeResourceExists, "用户名已存在")
	}

	// 检查邮箱是否已存在
	existingUser, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Email.Eq(req.Email)).First()
	if err == nil && existingUser != nil {
		return nil, common.NewBusinessError(common.ErrCodeResourceExists, "邮箱已存在")
	}

	// 验证密码
	if req.Password != req.ConfirmPassword {
		return nil, common.NewBusinessError(common.ErrCodeInvalidParam, "密码确认不匹配")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &model.AdminUser{
		UUID:         fmt.Sprintf("user-%d", time.Now().Unix()),
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hashedPassword),
		RealName:     req.Name,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.q.AdminUser.WithContext(ctx).Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 转换为域模型
	userModel := &identity.User{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.RealName,
		Status:    "active",
		Roles:     []identity.Role{},
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	return userModel, nil
}

// GetUser 获取用户详情
func (s *SimpleIdentityService) GetUser(ctx context.Context, userID string) (*identity.User, error) {
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return nil, err
	}

	// 转换为域模型
	userModel := &identity.User{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.RealName,
		Status:    fmt.Sprintf("%t", user.IsActive),
		Roles:     []identity.Role{}, // 简化实现
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	return userModel, nil
}

// GetUserByUUID 根据UUID获取用户
func (s *SimpleIdentityService) GetUserByUUID(ctx context.Context, uuid string) (*identity.User, error) {
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(uuid)).First()
	if err != nil {
		return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
	}

	return &identity.User{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.RealName,
		Status:    fmt.Sprintf("%t", user.IsActive),
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}, nil
}

// ListUsers 分页查询用户列表
func (s *SimpleIdentityService) ListUsers(ctx context.Context, page, pageSize int, status string) ([]identity.User, int64, error) {
	// 构建查询条件
	query := s.q.AdminUser.WithContext(ctx)

	// 如果指定了状态，添加状态过滤
	if status != "" {
		if status == "active" {
			query = query.Where(s.q.AdminUser.IsActive.Is(true))
		} else if status == "inactive" {
			query = query.Where(s.q.AdminUser.IsActive.Is(false))
		}
	}

	// 计算总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	users, err := query.Offset(offset).Limit(pageSize).Find()
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	// 转换为域模型
	result := make([]identity.User, len(users))
	for i, user := range users {
		result[i] = identity.User{
			ID:        user.ID,
			UUID:      user.UUID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Name:      user.RealName,
			Status:    fmt.Sprintf("%t", user.IsActive),
			Roles:     []identity.Role{}, // 简化实现
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		}
	}

	return result, total, nil
}

// UpdateUser 更新用户信息
func (s *SimpleIdentityService) UpdateUser(ctx context.Context, userID string, req identity.UpdateUserRequest) error {
	// 查找用户
	_, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return err
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Email != nil {
		// 检查邮箱是否已被其他用户使用
		existingUser, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Email.Eq(*req.Email)).First()
		if err == nil && existingUser.UUID != userID {
			return common.NewBusinessError(common.ErrCodeResourceExists, "邮箱已被其他用户使用")
		}
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		// 检查手机号是否已被其他用户使用
		existingUser, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.Phone.Eq(*req.Phone)).First()
		if err == nil && existingUser.UUID != userID {
			return common.NewBusinessError(common.ErrCodeResourceExists, "手机号已被其他用户使用")
		}
		updates["phone"] = *req.Phone
	}
	if req.Name != nil {
		updates["real_name"] = *req.Name
	}
	if req.Status != nil {
		updates["is_active"] = *req.Status == "active"
	}
	updates["updated_at"] = time.Now()

	// 执行更新
	_, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).Updates(updates)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// ChangePassword 修改密码
func (s *SimpleIdentityService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// 查找用户
	user, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return common.NewBusinessError(common.ErrCodeInvalidParam, "原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	_, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).Update(s.q.AdminUser.PasswordHash, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// DeactivateUser 停用用户
func (s *SimpleIdentityService) DeactivateUser(ctx context.Context, userID string) error {
	// 查找用户
	_, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return err
	}

	// 更新用户状态为停用
	_, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).Update(s.q.AdminUser.IsActive, false)
	if err != nil {
		return fmt.Errorf("停用用户失败: %w", err)
	}

	return nil
}

// ActivateUser 激活用户
func (s *SimpleIdentityService) ActivateUser(ctx context.Context, userID string) error {
	// 查找用户
	_, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return err
	}

	// 更新用户状态为激活
	_, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).Update(s.q.AdminUser.IsActive, true)
	if err != nil {
		return fmt.Errorf("激活用户失败: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (s *SimpleIdentityService) DeleteUser(ctx context.Context, userID string) error {
	// 查找用户
	_, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewBusinessError(common.ErrCodeResourceNotFound, "用户不存在")
		}
		return err
	}

	// 软删除用户（设置deleted_at字段）
	_, err = s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.Eq(userID)).Update(s.q.AdminUser.DeletedAt, time.Now())
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// BatchGetUsers 批量获取用户
func (s *SimpleIdentityService) BatchGetUsers(ctx context.Context, uuids []string) ([]identity.User, error) {
	if len(uuids) == 0 {
		return []identity.User{}, nil
	}

	// 批量查询用户
	users, err := s.q.AdminUser.WithContext(ctx).Where(s.q.AdminUser.UUID.In(uuids...)).Find()
	if err != nil {
		return nil, fmt.Errorf("批量查询用户失败: %w", err)
	}

	// 转换为域模型
	result := make([]identity.User, len(users))
	for i, user := range users {
		result[i] = identity.User{
			ID:        user.ID,
			UUID:      user.UUID,
			Username:  user.Username,
			Email:     user.Email,
			Phone:     user.Phone,
			Name:      user.RealName,
			Status:    fmt.Sprintf("%t", user.IsActive),
			Roles:     []identity.Role{}, // 简化实现
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		}
	}

	return result, nil
}

// ===== RoleService 接口实现 =====

// CreateRole 创建角色
func (s *SimpleIdentityService) CreateRole(ctx context.Context, req identity.CreateRoleRequest) (*identity.Role, error) {
	// 检查角色名称是否已存在
	existingRole, err := s.q.Role.WithContext(ctx).Where(s.q.Role.Name.Eq(req.Name)).First()
	if err == nil && existingRole != nil {
		return nil, common.NewBusinessError(common.ErrCodeResourceExists, "角色名称已存在")
	}

	// 创建角色
	role := &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsActive:    req.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.q.Role.WithContext(ctx).Create(role); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	// 转换为域模型
	roleModel := &identity.Role{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsActive:    role.IsActive,
		CreatedAt:   role.CreatedAt.Unix(),
	}

	return roleModel, nil
}

// GetRole 获取角色详情
func (s *SimpleIdentityService) GetRole(ctx context.Context, roleID string) (*identity.Role, error) {
	// 转换roleID为int64
	id, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return nil, common.NewBusinessError(common.ErrCodeInvalidParam, "无效的角色ID")
	}

	role, err := s.q.Role.WithContext(ctx).Where(s.q.Role.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewBusinessError(common.ErrCodeResourceNotFound, "角色不存在")
		}
		return nil, err
	}

	// 转换为域模型
	roleModel := &identity.Role{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsActive:    role.IsActive,
		CreatedAt:   role.CreatedAt.Unix(),
	}

	return roleModel, nil
}

// ListRoles 查询角色列表
func (s *SimpleIdentityService) ListRoles(ctx context.Context) ([]identity.Role, error) {
	roles, err := s.q.Role.WithContext(ctx).Find()
	if err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %w", err)
	}

	result := make([]identity.Role, len(roles))
	for i, role := range roles {
		result[i] = identity.Role{
			ID:          role.ID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			IsActive:    role.IsActive,
			CreatedAt:   role.CreatedAt.Unix(),
		}
	}
	return result, nil
}

// UpdateRole 更新角色信息
func (s *SimpleIdentityService) UpdateRole(ctx context.Context, roleID string, req identity.UpdateRoleRequest) error {
	// 转换roleID为int64
	id, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return common.NewBusinessError(common.ErrCodeInvalidParam, "无效的角色ID")
	}

	// 查找角色
	_, err = s.q.Role.WithContext(ctx).Where(s.q.Role.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewBusinessError(common.ErrCodeResourceNotFound, "角色不存在")
		}
		return err
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["updated_at"] = time.Now()

	// 执行更新
	_, err = s.q.Role.WithContext(ctx).Where(s.q.Role.ID.Eq(id)).Updates(updates)
	if err != nil {
		return fmt.Errorf("更新角色失败: %w", err)
	}

	return nil
}

// DeleteRole 删除角色
func (s *SimpleIdentityService) DeleteRole(ctx context.Context, roleID string) error {
	// 转换roleID为int64
	id, err := strconv.ParseInt(roleID, 10, 64)
	if err != nil {
		return common.NewBusinessError(common.ErrCodeInvalidParam, "无效的角色ID")
	}

	// 查找角色
	_, err = s.q.Role.WithContext(ctx).Where(s.q.Role.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewBusinessError(common.ErrCodeResourceNotFound, "角色不存在")
		}
		return err
	}

	// 软删除角色（设置deleted_at字段）
	_, err = s.q.Role.WithContext(ctx).Where(s.q.Role.ID.Eq(id)).Update(s.q.Role.DeletedAt, time.Now())
	if err != nil {
		return fmt.Errorf("删除角色失败: %w", err)
	}

	return nil
}

// AssignRoleToUser 给用户分配角色
func (s *SimpleIdentityService) AssignRoleToUser(ctx context.Context, userID string, roleNames []string) error {
	return fmt.Errorf("分配角色功能待实现")
}

// RemoveRoleFromUser 移除用户角色
func (s *SimpleIdentityService) RemoveRoleFromUser(ctx context.Context, userID string, roleName string) error {
	return fmt.Errorf("移除角色功能待实现")
}

// ===== PermissionService 接口实现 =====

// AddPermission 添加权限
func (s *SimpleIdentityService) AddPermission(ctx context.Context, req identity.PermissionRequest) error {
	_, err := s.enforcer.AddPolicy(req.Role, req.Path, req.Method)
	return err
}

// RemovePermission 移除权限
func (s *SimpleIdentityService) RemovePermission(ctx context.Context, req identity.PermissionRequest) error {
	_, err := s.enforcer.RemovePolicy(req.Role, req.Path, req.Method)
	return err
}

// ListRolePermissions 查询角色权限
func (s *SimpleIdentityService) ListRolePermissions(ctx context.Context, role string) ([]identity.Permission, error) {
	policies, err := s.enforcer.GetFilteredPolicy(0, role)
	if err != nil {
		return nil, err
	}

	permissions := make([]identity.Permission, len(policies))
	for i, policy := range policies {
		if len(policy) >= 3 {
			permissions[i] = identity.Permission{
				Role:   policy[0],
				Path:   policy[1],
				Method: policy[2],
				Action: fmt.Sprintf("%s:%s", policy[2], policy[1]),
			}
		}
	}

	return permissions, nil
}

// CheckPermission 检查权限
func (s *SimpleIdentityService) CheckPermission(ctx context.Context, userRoles []string, path, method string) (bool, error) {
	for _, role := range userRoles {
		allowed, err := s.enforcer.Enforce(role, path, method)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// ListAllPermissions 查询所有权限策略
func (s *SimpleIdentityService) ListAllPermissions(ctx context.Context) ([]identity.Permission, error) {
	policies, err := s.enforcer.GetPolicy()
	if err != nil {
		return nil, err
	}

	permissions := make([]identity.Permission, len(policies))
	for i, policy := range policies {
		if len(policy) >= 3 {
			permissions[i] = identity.Permission{
				Role:   policy[0],
				Path:   policy[1],
				Method: policy[2],
				Action: fmt.Sprintf("%s:%s", policy[2], policy[1]),
			}
		}
	}

	return permissions, nil
}

// UpdateRolePermissions 批量更新角色权限
func (s *SimpleIdentityService) UpdateRolePermissions(ctx context.Context, role string, permissions []identity.PermissionRequest) error {
	// 删除角色的所有权限
	_, err := s.enforcer.RemoveFilteredPolicy(0, role)
	if err != nil {
		return err
	}

	// 添加新权限
	for _, perm := range permissions {
		_, err = s.enforcer.AddPolicy(role, perm.Path, perm.Method)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetRolesForUser 获取用户的所有角色
func (s *SimpleIdentityService) GetRolesForUser(ctx context.Context, userID string) ([]string, error) {
	roles, err := s.enforcer.GetRolesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}
	return roles, nil
}

// GetUsersForRole 获取角色的所有用户
func (s *SimpleIdentityService) GetUsersForRole(ctx context.Context, role string) ([]string, error) {
	users, err := s.enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("获取角色用户失败: %w", err)
	}
	return users, nil
}

// 确保实现了所有接口
var _ identity.Service = (*SimpleIdentityService)(nil)
