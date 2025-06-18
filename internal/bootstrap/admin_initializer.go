package bootstrap

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"errors"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// initSuperAdmin 确保系统启动时存在一个超级管理员账号，并通过 Casbin 授予对应权限。
func initSuperAdmin(db *gorm.DB) error {
	log := logger.GetGlobalLogger()
	opts := config.GetInstance()

	q := query.Use(db)

	// 超级管理员用户名密码从配置读取（或使用默认值）
	adminCfg := opts.Auth.SuperAdmin // 假设配置项存在
	if adminCfg.Username == "" {
		adminCfg.Username = "admin"
	}
	if adminCfg.Password == "" {
		adminCfg.Password = "admin123"
	}

	ctx := context.Background()

	admin, err := q.AdminUser.WithContext(ctx).Where(q.AdminUser.Username.Eq(adminCfg.Username)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(adminCfg.Password), bcrypt.DefaultCost)
		admin = &model.AdminUser{
			ID:           uuid.New().String(),
			Username:     adminCfg.Username,
			PasswordHash: string(hashed),
			Email:        adminCfg.Email,
			RealName:     "SuperAdmin",
			IsActive:     true,
		}
		if err := q.AdminUser.WithContext(ctx).Create(admin); err != nil {
			return err
		}
		log.Info("Super admin account created", zap.String("username", adminCfg.Username))
	}

	// 初始化 Casbin，使该用户拥有超级管理员角色
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &model.CasbinRule{})
	if err != nil {
		return err
	}
	enforcer, err := casbin.NewEnforcer("./config/rbac_model.conf", adapter)
	if err != nil {
		return err
	}
	enforcer.LoadPolicy()

	// 设定角色，p ruls maybe: g, userID, role::super
	_, _ = enforcer.AddGroupingPolicy(admin.ID, "super_admin")
	enforcer.SavePolicy()

	return nil
}
