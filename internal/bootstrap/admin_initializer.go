package bootstrap

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"errors"

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

	return nil
}
