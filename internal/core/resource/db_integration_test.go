//go:build integration

package resource

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/dao/model"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// 测试配置获取函数
func getTestConfig() config.Options {
	// 方案1: 检查环境变量指定的配置文件
	if configPath := os.Getenv("TEST_CONFIG_FILE"); configPath != "" {
		if err := config.InitOptions(configPath); err == nil {
			return *config.GetInstance()
		}
	}

	// 方案2: 尝试自动发现项目根目录
	if opts, err := loadConfigFromProjectRoot(); err == nil {
		return opts
	}

	// 方案3: 使用内存中的测试配置（最后的备选方案）
	return createInMemoryTestConfig()
}

// 自动发现项目根目录并加载配置
func loadConfigFromProjectRoot() (config.Options, error) {
	// 获取当前文件的目录
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return config.Options{}, fmt.Errorf("unable to get current file path")
	}

	// 从当前文件向上查找项目根目录（包含go.mod的目录）
	dir := filepath.Dir(filename)
	for {
		// 检查是否存在go.mod文件
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// 找到项目根目录，尝试加载测试配置
			testConfigPaths := []string{
				filepath.Join(dir, "config", "app.test.yaml"),
				filepath.Join(dir, "config", "test.yaml"),
				filepath.Join(dir, "configs", "test.yaml"),
				filepath.Join(dir, "test.yaml"),
			}

			for _, path := range testConfigPaths {
				if _, err := os.Stat(path); err == nil {
					if err := config.InitOptions(path); err == nil {
						return *config.GetInstance(), nil
					}
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// 已经到达文件系统根目录
			break
		}
		dir = parent
	}

	return config.Options{}, fmt.Errorf("unable to find project root or test config file")
}

// 创建内存中的测试配置
func createInMemoryTestConfig() config.Options {
	return config.Options{
		Logger: config.LogOptions{
			Level:    0, // Debug level
			Filename: "test.log",
			LineNum:  true,
		},
		Database: config.DBOptions{
			Driver:          getEnvOrDefault("TEST_DB_DRIVER", "mysql"),
			Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:            getIntEnvOrDefault("TEST_DB_PORT", 3306),
			User:            getEnvOrDefault("TEST_DB_USER", "root"),
			Password:        getEnvOrDefault("TEST_DB_PASSWORD", ""),
			DBName:          getEnvOrDefault("TEST_DB_NAME", "crm_lite_test"),
			TablePrefix:     getEnvOrDefault("TEST_DB_PREFIX", "test_"),
			SSLMode:         "disable",
			TimeZone:        "Asia/Shanghai",
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: time.Hour,
		},
	}
}

// 辅助函数：获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 辅助函数：获取整数类型的环境变量或默认值
func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// DBResourceIntegrationSuite 集成测试套件 - 测试与真实数据库的交互
type DBResourceIntegrationSuite struct {
	suite.Suite
	dbResource *DBResource
	opts       config.Options
}

// SetupSuite 在测试套件开始前运行
func (s *DBResourceIntegrationSuite) SetupSuite() {
	// 智能加载测试配置
	s.opts = getTestConfig()

	// 初始化全局Logger
	logger.InitGlobalLogger(&s.opts.Logger)

	s.dbResource = NewDBResource(s.opts.Database)

	// 初始化资源
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.dbResource.Initialize(ctx)
	s.Require().NoError(err, "DBResource Initialize() failed. Please check your database configuration.\n"+
		"You can set database config via environment variables:\n"+
		"  TEST_DB_HOST, TEST_DB_PORT, TEST_DB_USER, TEST_DB_PASSWORD, TEST_DB_NAME\n"+
		"Or provide a test config file via TEST_CONFIG_FILE environment variable.")

	// 确保测试表存在
	err = s.dbResource.AutoMigrate(&model.AdminUser{})
	s.Require().NoError(err, "Failed to auto-migrate AdminUser table")
}

// TearDownSuite 在测试套件结束后运行
func (s *DBResourceIntegrationSuite) TearDownSuite() {
	// 关闭资源
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.dbResource.Close(ctx)
	s.Require().NoError(err, "DBResource Close() failed")
}

// TestDatabaseConnection 测试数据库连接
func (s *DBResourceIntegrationSuite) TestDatabaseConnection() {
	t := s.T()

	// 测试数据库是否可以正常连接
	assert.NotNil(t, s.dbResource.DB, "Database connection should be established")

	// 测试Ping操作
	sqlDB, err := s.dbResource.DB.DB()
	assert.NoError(t, err, "Should be able to get underlying sql.DB")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = sqlDB.PingContext(ctx)
	assert.NoError(t, err, "Database should be pingable")
}

// TestConnectionPoolConfiguration 测试连接池配置
func (s *DBResourceIntegrationSuite) TestConnectionPoolConfiguration() {
	t := s.T()

	sqlDB, err := s.dbResource.DB.DB()
	assert.NoError(t, err, "Should be able to get underlying sql.DB")

	// 验证连接池配置是否生效
	stats := sqlDB.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1, "MaxOpenConnections should be configured")

	// 测试配置值是否正确应用
	expectedMaxOpen := int(s.opts.Database.MaxOpenConns)
	if expectedMaxOpen > 0 {
		assert.Equal(t, expectedMaxOpen, stats.MaxOpenConnections, "MaxOpenConnections should match configuration")
	}
}

// TestAutoMigrate 测试数据库迁移功能
func (s *DBResourceIntegrationSuite) TestAutoMigrate() {
	t := s.T()

	// 测试迁移一个简单的模型
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	err := s.dbResource.AutoMigrate(&TestModel{})
	assert.NoError(t, err, "AutoMigrate should succeed")

	// 验证表是否存在
	var exists bool
	tableName := "test_models"
	if s.opts.Database.TablePrefix != "" {
		tableName = s.opts.Database.TablePrefix + tableName
	}

	query := "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = ? AND table_name = ?)"
	err = s.dbResource.Raw(query, s.opts.Database.DBName, tableName).Scan(&exists).Error
	assert.NoError(t, err, "Should be able to check table existence")
	assert.True(t, exists, "Table should exist after migration")

	// 清理测试表
	err = s.dbResource.Migrator().DropTable(&TestModel{})
	assert.NoError(t, err, "Should be able to drop test table")
}

// TestCRUDOperations 测试基本的CRUD操作
func (s *DBResourceIntegrationSuite) TestCRUDOperations() {
	t := s.T()
	ctx := context.Background()

	testUserUUID := "integration-test-user-123"

	// 清理函数
	t.Cleanup(func() {
		s.dbResource.WithContext(ctx).Unscoped().Delete(&model.AdminUser{}, "uuid = ?", testUserUUID)
	})

	// 1. Create - 创建测试用户
	testUser := &model.AdminUser{
		UUID:         testUserUUID,
		Username:     "integration_test_user",
		Email:        "integration@test.com",
		PasswordHash: "test-hash",
		RealName:     "Integration Tester",
	}

	err := s.dbResource.WithContext(ctx).Create(testUser).Error
	assert.NoError(t, err, "Should be able to create user")

	// 2. Read - 读取用户
	var foundUser model.AdminUser
	err = s.dbResource.WithContext(ctx).Where("username = ?", "integration_test_user").First(&foundUser).Error
	assert.NoError(t, err, "Should be able to find user")
	assert.Equal(t, "integration_test_user", foundUser.Username)
	assert.Equal(t, "Integration Tester", foundUser.RealName)

	// 3. Update - 更新用户
	err = s.dbResource.WithContext(ctx).Model(&foundUser).Update("real_name", "Updated Tester").Error
	assert.NoError(t, err, "Should be able to update user")

	// 验证更新
	var updatedUser model.AdminUser
	err = s.dbResource.WithContext(ctx).Where("uuid = ?", testUserUUID).First(&updatedUser).Error
	assert.NoError(t, err, "Should be able to find updated user")
	assert.Equal(t, "Updated Tester", updatedUser.RealName)

	// 4. Delete - 删除用户（软删除）
	err = s.dbResource.WithContext(ctx).Delete(&updatedUser).Error
	assert.NoError(t, err, "Should be able to soft delete user")

	// 验证软删除
	var deletedUser model.AdminUser
	err = s.dbResource.WithContext(ctx).Where("uuid = ?", testUserUUID).First(&deletedUser).Error
	assert.Error(t, err, "Should not find soft deleted user")

	// 验证使用Unscoped可以找到软删除的记录
	var unscopedUser model.AdminUser
	err = s.dbResource.WithContext(ctx).Unscoped().Where("uuid = ?", testUserUUID).First(&unscopedUser).Error
	assert.NoError(t, err, "Should find soft deleted user with Unscoped")
}

// TestTransactionSupport 测试事务支持
func (s *DBResourceIntegrationSuite) TestTransactionSupport() {
	t := s.T()
	ctx := context.Background()

	testUserUUID1 := "tx-test-user-1"
	testUserUUID2 := "tx-test-user-2"

	// 清理函数
	t.Cleanup(func() {
		s.dbResource.WithContext(ctx).Unscoped().Delete(&model.AdminUser{}, "uuid IN ?", []string{testUserUUID1, testUserUUID2})
	})

	// 测试事务回滚
	err := s.dbResource.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建第一个用户
		user1 := &model.AdminUser{
			UUID:         testUserUUID1,
			Username:     "tx_user_1",
			Email:        "tx1@test.com",
			PasswordHash: "hash1",
		}
		if err := tx.Create(user1).Error; err != nil {
			return err
		}

		// 创建第二个用户，但是故意让它失败（重复用户名）
		user2 := &model.AdminUser{
			UUID:         testUserUUID2,
			Username:     "tx_user_1", // 重复的用户名
			Email:        "tx2@test.com",
			PasswordHash: "hash2",
		}
		return tx.Create(user2).Error // 这应该会失败
	})

	assert.Error(t, err, "Transaction should fail due to duplicate username")

	// 验证事务回滚，两个用户都不应该存在
	var count int64
	err = s.dbResource.WithContext(ctx).Model(&model.AdminUser{}).Where("uuid IN ?", []string{testUserUUID1, testUserUUID2}).Count(&count).Error
	assert.NoError(t, err, "Should be able to count users")
	assert.Equal(t, int64(0), count, "No users should exist after transaction rollback")
}

// TestConcurrentConnections 测试并发连接
func (s *DBResourceIntegrationSuite) TestConcurrentConnections() {
	t := s.T()
	ctx := context.Background()

	// 启动多个goroutine同时进行数据库操作
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			testUserUUID := fmt.Sprintf("concurrent-test-user-%d", id)

			// 创建用户
			user := &model.AdminUser{
				UUID:         testUserUUID,
				Username:     fmt.Sprintf("concurrent_user_%d", id),
				Email:        fmt.Sprintf("concurrent%d@test.com", id),
				PasswordHash: "hash",
			}

			err := s.dbResource.WithContext(ctx).Create(user).Error
			if err != nil {
				done <- err
				return
			}

			// 读取用户
			var foundUser model.AdminUser
			err = s.dbResource.WithContext(ctx).Where("uuid = ?", testUserUUID).First(&foundUser).Error
			if err != nil {
				done <- err
				return
			}

			// 删除用户
			err = s.dbResource.WithContext(ctx).Delete(&foundUser).Error
			done <- err
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err, "Concurrent operation should succeed")
	}
}

// TestDBResourceIntegrationSuite 启动集成测试套件
func TestDBResourceIntegration(t *testing.T) {
	// 只有在明确指定运行集成测试时才执行
	// 可以通过 go test -tags=integration 来运行
	suite.Run(t, new(DBResourceIntegrationSuite))
}
