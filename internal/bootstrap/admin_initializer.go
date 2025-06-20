package bootstrap

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	dao_model "crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"embed"
	"errors"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//go:embed model.conf
var casbinModel embed.FS

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
	if adminCfg.Role == "" {
		adminCfg.Role = "super_admin"
	}

	ctx := context.Background()

	admin, err := q.AdminUser.WithContext(ctx).Where(q.AdminUser.Username.Eq(adminCfg.Username)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("Failed to query super admin", zap.Error(err))
		return err
	}

	// 如果超级管理员不存在，则创建
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 使用配置的加密成本
		cost := opts.Auth.BCryptCost
		if cost == 0 {
			cost = bcrypt.DefaultCost // fallback
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(adminCfg.Password), cost)
		if err != nil {
			log.Error("Failed to hash password for super admin", zap.Error(err))
			return err
		}

		admin = &dao_model.AdminUser{
			UUID:         uuid.New().String(),
			Username:     adminCfg.Username,
			PasswordHash: string(hashed),
			Email:        adminCfg.Email,
			RealName:     "SuperAdmin",
			IsActive:     true,
		}
		if err := q.AdminUser.WithContext(ctx).Create(admin); err != nil {
			log.Error("Failed to create super admin account", zap.Error(err))
			return err
		}
		log.Info("Super admin account created", zap.String("username", adminCfg.Username))
	} else {
		log.Info("Super admin account already exists, skipping creation.", zap.String("username", admin.Username))
	}

	// 初始化 Casbin，使该用户拥有超级管理员角色
	// 使用 NewAdapterByDBUseTableName 指定表名，表名和前缀均从配置和模型常量获取
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, opts.Database.TablePrefix, dao_model.TableNameCasbinRule)
	if err != nil {
		log.Error("Failed to create casbin adapter", zap.Error(err))
		return err
	}
	// 从嵌入的文件系统创建 casbin model
	m, err := casbinmodel.NewModelFromString(string(mustLoadFile(casbinModel, "model.conf")))
	if err != nil {
		log.Error("Failed to create casbin model from string", zap.Error(err))
		return err
	}
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Error("Failed to create casbin enforcer", zap.Error(err))
		return err
	}
	if err := enforcer.LoadPolicy(); err != nil {
		log.Error("Failed to load policy from storage", zap.Error(err))
		return err
	}

	// 设定角色
	hasPolicy := enforcer.HasGroupingPolicy(admin.UUID, adminCfg.Role)
	if !hasPolicy {
		if _, err := enforcer.AddGroupingPolicy(admin.UUID, adminCfg.Role); err != nil {
			log.Error("Failed to add grouping policy for super admin", zap.Error(err))
			return err
		}
		if err := enforcer.SavePolicy(); err != nil {
			log.Error("Failed to save policy after adding role for super admin", zap.Error(err))
			return err
		}
		log.Info("Successfully granted role to super admin",
			zap.String("user_uuid", admin.UUID),
			zap.String("role", adminCfg.Role),
		)
	}

	return nil
}

func mustLoadFile(fs embed.FS, path string) []byte {
	b, err := fs.ReadFile(path)
	if err != nil {
		panic("cannot read embedded file: " + path)
	}
	return b
}
