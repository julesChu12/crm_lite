//go:build integration

package resource

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/dao/model"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DBResourceIntegrationSuite 是用于DBResource的集成测试套件
type DBResourceIntegrationSuite struct {
	suite.Suite
	dbResource *DBResource
	opts       config.Options
}

// SetupSuite 在测试套件开始前运行
func (s *DBResourceIntegrationSuite) SetupSuite() {
	// 加载测试配置
	// 假设测试从项目根目录运行
	configFile := "../../../config/app.test.yaml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		s.T().Fatalf("Test config file not found: %s. Make sure you are running tests from the project root.", configFile)
	}

	err := config.InitOptions(configFile)
	s.Require().NoError(err, "Failed to initialize configuration")

	s.opts = *config.GetInstance()

	// 初始化全局Logger，因为DBResource依赖它
	logger.InitGlobalLogger(&s.opts.Logger)

	s.dbResource = NewDBResource(s.opts.Database)

	// 初始化资源
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = s.dbResource.Initialize(ctx)
	s.Require().NoError(err, "DBResource Initialize() failed")

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

// TestDBResourceLifecycle 测试DBResource的完整生命周期（写入、读取、删除）
func (s *DBResourceIntegrationSuite) TestDBResourceLifecycle() {
	t := s.T()
	ctx := context.Background()

	testUserID := "test-user-id-123"

	// 使用 t.Cleanup 确保测试数据在测试结束后被清理
	t.Cleanup(func() {
		// 使用 Unscoped.Delete 物理删除记录，无论它是否被软删除
		// 这样可以确保下次测试运行环境是干净的
		s.dbResource.WithContext(ctx).Unscoped().Delete(&model.AdminUser{}, "id = ?", testUserID)
	})

	// 1. 创建一个测试用户
	testUser := &model.AdminUser{
		ID:           testUserID,
		Username:     "integration_test_user",
		Email:        "test@example.com",
		PasswordHash: "some-hash",
		RealName:     "Tester",
	}

	// 2. 写入数据
	err := s.dbResource.WithContext(ctx).Create(testUser).Error
	assert.NoError(t, err, "Failed to create test user")

	// 3. 读取并验证数据
	var foundUser model.AdminUser
	err = s.dbResource.WithContext(ctx).Where("username = ?", "integration_test_user").First(&foundUser).Error
	assert.NoError(t, err, "Failed to find test user")
	assert.Equal(t, "integration_test_user", foundUser.Username, "Username does not match")
	assert.Equal(t, "Tester", foundUser.RealName, "RealName does not match")

	// 4. 删除数据 (这里我们执行软删除)
	err = s.dbResource.WithContext(ctx).Delete(&foundUser).Error
	assert.NoError(t, err, "Failed to delete test user")

	// 5. 验证删除
	var deletedUser model.AdminUser
	err = s.dbResource.WithContext(ctx).Where("username = ?", "integration_test_user").First(&deletedUser).Error
	assert.Error(t, err, "Test user should have been deleted, but was found")
}

// TestDBResourceIntegrationSuite 启动测试套件
func TestDBResourceIntegration(t *testing.T) {
	suite.Run(t, new(DBResourceIntegrationSuite))
}
