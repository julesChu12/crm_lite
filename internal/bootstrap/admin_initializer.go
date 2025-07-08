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

	admin, err := q.AdminUser.WithContext(ctx).Where(q.AdminUser.Username.Eq(adminCfg.Username)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("Failed to query super admin", zap.Error(err))
		return err
	}

	// 4. 如不存在则创建超级管理员用户
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cost := opts.Auth.BCryptCost
		if cost == 0 {
			cost = bcrypt.DefaultCost
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(adminCfg.Password), cost)
		if err != nil {
			logger.Error("Failed to hash password for super admin", zap.Error(err))
			return err
		}
		admin = &model.AdminUser{
			UUID:         uuid.New().String(),
			Username:     adminCfg.Username,
			PasswordHash: string(hashed),
			Email:        adminCfg.Email,
			RealName:     "SuperAdmin",
			IsActive:     true,
		}
		if err := q.AdminUser.WithContext(ctx).Create(admin); err != nil {
			logger.Error("Failed to create super admin", zap.Error(err))
			return err
		}
		logger.Info("Super admin account created", zap.String("username", adminCfg.Username))
	} else {
		logger.Info("Super admin account already exists", zap.String("username", admin.Username))
	}

	// 5. 给予角色（Grouping policy）
	has, err := enforcer.HasGroupingPolicy(admin.UUID, adminCfg.Role)
	if err != nil {
		return err
	}
	if !has {
		if _, err := enforcer.AddGroupingPolicy(admin.UUID, adminCfg.Role); err != nil {
			return err
		}
		if err := enforcer.SavePolicy(); err != nil {
			return err
		}
		logger.Info("Granted role to super admin", zap.String("role", adminCfg.Role))
	}

	// 6. 为超级管理员角色同步所有API权限
	if err := syncSuperAdminPermissions(enforcer, adminCfg.Role); err != nil {
		logger.Error("Failed to sync permissions for super admin role", zap.Error(err), zap.String("role", adminCfg.Role))
		return err
	}

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
