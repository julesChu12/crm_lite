package testutil

import (
	"bytes"
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ControllerTestSuite 控制器测试套件
type ControllerTestSuite struct {
	DB      *gorm.DB
	Query   *query.Query
	Router  *gin.Engine
	Manager *resource.Manager
	cleanup func()
}

// SetupControllerTestSuite 设置控制器测试套件
func SetupControllerTestSuite(t *testing.T) *ControllerTestSuite {
	// 跳过集成测试，除非设置了环境变量
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("Skipping integration test - set RUN_DB_TESTS=1 to run")
	}

	// 设置测试数据库
	db, cleanup, err := SetupTestDatabase()
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.Customer{},
		&model.Product{},
		&model.Wallet{},
		&model.WalletTransaction{},
		&model.Order{},
		&model.OrderItem{},
		&model.AdminUser{},
	)
	require.NoError(t, err)

	// 创建query实例
	q := query.Use(db)

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建资源管理器mock
	// 注意：这里需要根据实际的resource.Manager实现来调整
	// 由于我们无法看到resource.Manager的完整实现，这里提供一个框架
	manager := &resource.Manager{}

	suite := &ControllerTestSuite{
		DB:      db,
		Query:   q,
		Router:  router,
		Manager: manager,
		cleanup: cleanup,
	}

	return suite
}

// Cleanup 清理测试资源
func (suite *ControllerTestSuite) Cleanup() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// CreateTestCustomer 创建测试客户
func (suite *ControllerTestSuite) CreateTestCustomer(id int64, name string) *model.Customer {
	customer := &model.Customer{
		ID:         id,
		Name:       name,
		Phone:      fmt.Sprintf("139%08d", id),
		Email:      fmt.Sprintf("test_%d@example.com", id),
		Gender:     "unknown",
		Birthday:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Level:      "普通",
		Tags:       "[]",
		Source:     "manual",
		AssignedTo: 1,
	}
	err := suite.Query.Customer.WithContext(context.Background()).Create(customer)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test customer: %v", err))
	}
	return customer
}

// CreateTestProduct 创建测试商品
func (suite *ControllerTestSuite) CreateTestProduct(id int64, name string, price float64) *model.Product {
	product := &model.Product{
		ID:            id,
		Name:          name,
		Description:   fmt.Sprintf("测试商品 %s", name),
		Type:          "product",
		Category:      "测试分类",
		Price:         price,
		Cost:          price * 0.5,
		StockQuantity: 100,
		Unit:          "件",
		IsActive:      true,
	}
	err := suite.Query.Product.WithContext(context.Background()).Create(product)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test product: %v", err))
	}
	return product
}

// CreateTestWallet 创建测试钱包
func (suite *ControllerTestSuite) CreateTestWallet(customerID int64, balance int64) *model.Wallet {
	wallet := &model.Wallet{
		CustomerID: customerID,
		Balance:    balance,
		Status:     1, // 正常状态
		UpdatedAt:  time.Now().Unix(),
	}
	err := suite.Query.Wallet.WithContext(context.Background()).Create(wallet)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test wallet: %v", err))
	}
	return wallet
}

// CreateTestAdmin 创建测试管理员
func (suite *ControllerTestSuite) CreateTestAdmin(id int64, username string) *model.AdminUser {
	admin := &model.AdminUser{
		ID:           id,
		UUID:         fmt.Sprintf("test-admin-%d", id),
		Username:     username,
		Email:        fmt.Sprintf("%s@test.com", username),
		PasswordHash: "$2a$10$test.hash", // Mock密码hash
		RealName:     fmt.Sprintf("测试管理员%d", id),
		Phone:        fmt.Sprintf("138%08d", id),
		IsActive:     true,
		LastLoginAt:  time.Now(),
	}
	err := suite.Query.AdminUser.WithContext(context.Background()).Create(admin)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test admin: %v", err))
	}
	return admin
}

// MakeJSONRequest 创建JSON请求
func (suite *ControllerTestSuite) MakeJSONRequest(method, path string, body interface{}) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("Failed to marshal request body: %v", err))
		}
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, path, bodyReader)
	if err != nil {
		panic(fmt.Sprintf("Failed to create request: %v", err))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

// ExecuteRequest 执行HTTP请求
func (suite *ControllerTestSuite) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	suite.Router.ServeHTTP(w, req)
	return w
}

// CleanupTestData 清理测试数据
func (suite *ControllerTestSuite) CleanupTestData(tables ...string) {
	err := CleanupTestData(suite.DB, tables...)
	if err != nil {
		panic(fmt.Sprintf("Failed to cleanup test data: %v", err))
	}
}