package bootstrap

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"errors"

	"github.com/casbin/casbin/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// initSuperAdmin 确保系统启动时存在一个超级管理员账号，并给其授予角色
func initSuperAdmin(rm *resource.Manager) error {
	opts := config.GetInstance()

	// 1. 获取数据库连接
	dbRes, err := resource.Get[*resource.DBResource](rm, resource.DBServiceKey)
	if err != nil {
		return err
	}
	db := dbRes.DB

	// 2. 获取 Casbin Enforcer（已由 CasbinResource 初始化）
	casRes, err := resource.Get[*resource.CasbinResource](rm, resource.CasbinServiceKey)
	if err != nil {
		return err
	}
	enforcer := casRes.GetEnforcer()

	// 3. 准备查询
	q := query.Use(db)

	adminCfg := opts.Auth.SuperAdmin
	if adminCfg.Username == "" {
		adminCfg.Username = "admin"
	}
	if adminCfg.Password == "" {
		adminCfg.Password = "admin123"
	}
	if adminCfg.Role == "" {
		adminCfg.Role = "super_admin"
	}

	ctx := context.Background()

	// 使用事务确保所有操作的原子性
	err = q.Transaction(func(tx *query.Query) error {
		// 1. 查找或创建超级管理员角色
		var role *model.Role
		role, err := tx.Role.WithContext(ctx).Where(tx.Role.Name.Eq(adminCfg.Role)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = &model.Role{
				Name:        adminCfg.Role,
				DisplayName: "超级管理员",
				Description: "系统超级管理员",
				IsActive:    true,
			}
			if err = tx.Role.WithContext(ctx).Create(role); err != nil {
				return err
			}
			logger.Info("Super admin role created", zap.String("role", adminCfg.Role))
		}

		// 2. 查找或创建超级管理员用户
		var admin *model.AdminUser
		admin, err = tx.AdminUser.WithContext(ctx).Where(tx.AdminUser.Username.Eq(adminCfg.Username)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cost := opts.Auth.BCryptCost
			if cost == 0 {
				cost = bcrypt.DefaultCost
			}
			hashed, err_hash := bcrypt.GenerateFromPassword([]byte(adminCfg.Password), cost)
			if err_hash != nil {
				return err_hash
			}
			admin = &model.AdminUser{
				UUID:         uuid.New().String(),
				Username:     adminCfg.Username,
				PasswordHash: string(hashed),
				Email:        adminCfg.Email,
				RealName:     "SuperAdmin",
				IsActive:     true,
			}
			if err = tx.AdminUser.WithContext(ctx).Create(admin); err != nil {
				return err
			}
			logger.Info("Super admin account created", zap.String("username", adminCfg.Username))
		} else {
			logger.Info("Super admin account already exists", zap.String("username", admin.Username))
		}

		// 3. 关联用户和角色
		_, err = tx.AdminUserRole.WithContext(ctx).Where(tx.AdminUserRole.AdminUserID.Eq(admin.ID), tx.AdminUserRole.RoleID.Eq(role.ID)).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userRole := &model.AdminUserRole{
				AdminUserID: admin.ID,
				RoleID:      role.ID,
			}
			if err = tx.AdminUserRole.WithContext(ctx).Create(userRole); err != nil {
				return err
			}
			logger.Info("Associated super admin user with role in database.")
		}

		// 4. 更新 Casbin 中的策略
		has, err_casbin := enforcer.HasGroupingPolicy(admin.UUID, role.Name)
		if err_casbin != nil {
			return err_casbin
		}
		if !has {
			if _, err_add := enforcer.AddGroupingPolicy(admin.UUID, role.Name); err_add != nil {
				return err_add
			}
			logger.Info("Granted role to super admin in Casbin", zap.String("role", role.Name))
		}
		return nil
	})
	if err != nil {
		logger.Error("Failed to initialize super admin", zap.Error(err))
		return err
	}

	// 5. 初始化并同步默认角色权限
	if err := initDefaultRole(enforcer, opts.Auth.DefaultRole); err != nil {
		logger.Error("Failed to initialize default role", zap.Error(err))
		return err
	}

	// 6. 为超级管理员角色同步所有API权限
	if err := syncSuperAdminPermissions(enforcer, adminCfg.Role); err != nil {
		logger.Error("Failed to sync permissions for super admin role", zap.Error(err), zap.String("role", adminCfg.Role))
		return err
	}

	return nil
}

// initDefaultRole 确保默认角色存在，并为其分配基础权限
func initDefaultRole(enforcer *casbin.Enforcer, defaultRole string) error {
	if defaultRole == "" {
		logger.Warn("Default role name is not configured, skipping initialization.")
		return nil
	}

	logger.Info("Start syncing permissions for default role...", zap.String("role", defaultRole))

	// 定义默认角色需要的基础权限
	defaultPermissions := [][]string{
		{"/api/v1/auth/profile", "GET"},
		{"/api/v1/auth/profile", "PUT"},
		{"/api/v1/auth/password", "PUT"},
		{"/api/v1/auth/logout", "POST"},
	}

	var policiesAdded bool
	for _, p := range defaultPermissions {
		path, method := p[0], p[1]
		has, err := enforcer.HasPolicy(defaultRole, path, method)
		if err != nil {
			return err
		}
		if !has {
			if _, err := enforcer.AddPolicy(defaultRole, path, method); err != nil {
				return err
			}
			policiesAdded = true
			logger.Info("Granted new permission to default role",
				zap.String("role", defaultRole),
				zap.String("path", path),
				zap.String("method", method),
			)
		}
	}

	if policiesAdded {
		logger.Info("New permissions granted to default role, saving policies...")
		return enforcer.SavePolicy()
	}

	logger.Info("Default role permissions are already up to date.")
	return nil
}

// virtualRoleForAllApis 是一个特殊的、不存在的角色，用作占位符
// 与 cmd/tools/permission/discover.go 中的定义保持一致
const virtualRoleForAllApis = "_all_apis_"

// syncSuperAdminPermissions 确保超级管理员角色拥有系统中所有已发现的API资源的权限
func syncSuperAdminPermissions(enforcer *casbin.Enforcer, superAdminRole string) error {
	logger.Info("Start syncing permissions for super admin role...", zap.String("role", superAdminRole))

	// 1. 获取 _all_apis_ 角色拥有的所有权限，这代表了系统中所有可被分配的资源
	allApiPolicies, err := enforcer.GetFilteredPolicy(0, virtualRoleForAllApis)
	if err != nil {
		return err
	}
	if len(allApiPolicies) == 0 {
		logger.Warn("No API resources found for the virtual role. Has the discover tool been run?", zap.String("virtualRole", virtualRoleForAllApis))
		return nil
	}

	// 2. 为 superAdminRole 角色过滤出它已有的权限，用于快速查找
	currentPolicies, err := enforcer.GetFilteredPolicy(0, superAdminRole)
	if err != nil {
		return err
	}
	policyMap := make(map[string]struct{})
	for _, p := range currentPolicies {
		key := p[1] + "::" + p[2] // e.g., /api/v1/users::GET
		policyMap[key] = struct{}{}
	}

	// 3. 遍历所有API资源，为 superAdminRole 添加其尚未拥有的权限
	var policiesAdded bool
	for _, policy := range allApiPolicies {
		path := policy[1]
		method := policy[2]
		key := path + "::" + method

		if _, exists := policyMap[key]; !exists {
			if _, err := enforcer.AddPolicy(superAdminRole, path, method); err != nil {
				return err
			}
			policiesAdded = true
			logger.Info("Granted new permission to super admin role",
				zap.String("role", superAdminRole),
				zap.String("path", path),
				zap.String("method", method),
			)
		}
	}

	// 4. 如果有新权限被添加，则保存到数据库
	if policiesAdded {
		logger.Info("New permissions granted to super admin role, saving policies...")
		if err := enforcer.SavePolicy(); err != nil {
			return err
		}
		logger.Info("Policies saved successfully.")
	} else {
		logger.Info("Super admin role permissions are already up to date.")
	}

	return nil
}
