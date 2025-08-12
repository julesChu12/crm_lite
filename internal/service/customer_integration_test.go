package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dto"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// CustomerIntegrationTestSuite 是客户服务集成测试的测试套件
type CustomerIntegrationTestSuite struct {
	suite.Suite
	resManager *resource.Manager
	service    *CustomerService
	db         *resource.DBResource
}

// SetupSuite 在测试套件开始时运行
func (s *CustomerIntegrationTestSuite) SetupSuite() {
	// 使用在 main_test.go 中初始化并存储在全局变量中的资源
	if testResManager == nil {
		s.T().Fatal("Test resource manager was not initialized in TestMain")
	}
	s.resManager = testResManager

	// 获取数据库资源
	db, err := resource.Get[*resource.DBResource](s.resManager, resource.DBServiceKey)
	if err != nil {
		s.T().Fatalf("Failed to get db resource for test suite: %v", err)
	}
	s.db = db

	// 初始化 CustomerService
	customerRepo := NewCustomerRepo(s.db.DB)
	walletSvc := NewWalletService(s.resManager)
	s.service = NewCustomerService(customerRepo, walletSvc)
}

// TearDownSuite 在测试套件结束时运行
func (s *CustomerIntegrationTestSuite) TearDownSuite() {
	// TearDown 由 TestMain 统一处理，这里不需要做任何事
}

// BeforeTest 在每个测试方法运行前运行，用于清理数据库
func (s *CustomerIntegrationTestSuite) BeforeTest(suiteName, testName string) {
	// 清理数据，确保测试独立性
	s.db.DB.Exec("DELETE FROM customers")
	s.db.DB.Exec("DELETE FROM wallets")
	s.db.DB.Exec("DELETE FROM wallet_transactions")
}

// TestRunner 是启动测试套件的入口
func TestCustomerIntegration(t *testing.T) {
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("Skipping integration tests because RUN_DB_TESTS is not set")
		return
	}
	suite.Run(t, new(CustomerIntegrationTestSuite))
}

// TestCreateCustomerIntegration 是一个端到端的集成测试用例
func (s *CustomerIntegrationTestSuite) TestCreateCustomerIntegration() {
	ctx := context.Background()
	req := &dto.CustomerCreateRequest{
		Name:  "Integration Test User",
		Phone: "9876543210",
		Email: "integration@example.com",
		Tags:  "[]", // 提供一个有效的JSON数组作为默认值
	}

	// 1. 调用服务创建客户
	createdCustomer, err := s.service.CreateCustomer(ctx, req)

	// 断言服务调用没有错误
	s.NoError(err)
	s.NotNil(createdCustomer)
	s.Equal("Integration Test User", createdCustomer.Name)

	// 2. 直接从数据库验证数据
	var dbCustomer model.Customer
	result := s.db.DB.Where("phone = ?", "9876543210").First(&dbCustomer)

	// 断言数据库查询没有错误，并且数据被正确写入
	s.NoError(result.Error)
	s.Equal(createdCustomer.ID, dbCustomer.ID)
	s.Equal("Integration Test User", dbCustomer.Name)
	s.Equal("integration@example.com", dbCustomer.Email)

	// 3. 验证钱包是否也已创建
	var dbWallet model.Wallet
	walletResult := s.db.DB.Where("customer_id = ?", dbCustomer.ID).First(&dbWallet)
	s.NoError(walletResult.Error)
	s.Equal("balance", dbWallet.Type)
}
