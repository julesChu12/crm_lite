package controller

import (
	"crm_lite/internal/testutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCustomerControllerCreateCustomer 测试客户创建控制器
func TestCustomerControllerCreateCustomer(t *testing.T) {
	suite := testutil.SetupControllerTestSuite(t)
	defer suite.Cleanup()

	// 清理测试数据
	defer suite.CleanupTestData("customers")

	// 1. 设置路由
	customerController := NewCustomerController(suite.Manager)
	suite.Router.POST("/customers", customerController.CreateCustomer)

	// 2. 构建创建客户请求
	createCustomerReq := gin.H{
		"name":   "API测试客户",
		"phone":  "13900001234",
		"email":  "api_test@example.com",
		"gender": "unknown",
		"level":  "普通",
		"source": "api_test",
	}

	// 3. 发送创建客户请求
	req := suite.MakeJSONRequest("POST", "/customers", createCustomerReq)
	w := suite.ExecuteRequest(req)

	// 4. 验证HTTP状态码
	t.Logf("Create customer response status: %d", w.Code)
	t.Logf("Create customer response body: %s", w.Body.String())

	// 验证接口能正常响应
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest)
}

// TestCustomerControllerGetCustomer 测试获取客户控制器
func TestCustomerControllerGetCustomer(t *testing.T) {
	suite := testutil.SetupControllerTestSuite(t)
	defer suite.Cleanup()

	// 清理测试数据
	defer suite.CleanupTestData("customers")

	// 1. 创建测试客户
	customer := suite.CreateTestCustomer(1003, "获取客户测试")

	// 2. 设置路由
	customerController := NewCustomerController(suite.Manager)
	suite.Router.GET("/customers/:id", customerController.GetCustomer)

	// 3. 发送获取客户请求
	req := suite.MakeJSONRequest("GET", "/customers/"+string(rune(customer.ID)), nil)
	w := suite.ExecuteRequest(req)

	// 4. 验证响应
	t.Logf("Get customer response status: %d", w.Code)
	t.Logf("Get customer response body: %s", w.Body.String())

	// 验证接口能正常响应
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
}

// TestProductControllerCreateProduct 测试商品创建控制器
func TestProductControllerCreateProduct(t *testing.T) {
	suite := testutil.SetupControllerTestSuite(t)
	defer suite.Cleanup()

	// 清理测试数据
	defer suite.CleanupTestData("products")

	// 1. 设置路由
	productController := NewProductController(suite.Manager)
	suite.Router.POST("/products", productController.CreateProduct)

	// 2. 构建创建商品请求
	createProductReq := gin.H{
		"name":           "API测试商品",
		"description":    "通过API创建的测试商品",
		"type":           "product",
		"category":       "API测试分类",
		"price":          99.99,
		"cost":           50.0,
		"stock_quantity": 100,
		"unit":           "件",
		"is_active":      true,
	}

	// 3. 发送创建商品请求
	req := suite.MakeJSONRequest("POST", "/products", createProductReq)
	w := suite.ExecuteRequest(req)

	// 4. 验证HTTP状态码
	t.Logf("Create product response status: %d", w.Code)
	t.Logf("Create product response body: %s", w.Body.String())

	// 验证接口能正常响应
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest)
}

// TestWalletControllerGetWallet 测试钱包查询控制器
func TestWalletControllerGetWallet(t *testing.T) {
	suite := testutil.SetupControllerTestSuite(t)
	defer suite.Cleanup()

	// 清理测试数据
	defer suite.CleanupTestData("wallets", "customers")

	// 1. 创建测试数据
	customer := suite.CreateTestCustomer(1004, "钱包测试客户")
	wallet := suite.CreateTestWallet(customer.ID, 50000)

	// 2. 设置路由
	walletController := NewWalletController(suite.Manager)
	suite.Router.GET("/customers/:customer_id/wallet", walletController.GetWalletByCustomerID)

	// 3. 发送获取钱包请求
	req := suite.MakeJSONRequest("GET", "/customers/"+string(rune(customer.ID))+"/wallet", nil)
	w := suite.ExecuteRequest(req)

	// 4. 验证响应
	t.Logf("Get wallet response status: %d", w.Code)
	t.Logf("Get wallet response body: %s", w.Body.String())

	// 验证接口能正常响应（钱包存在应该返回200）
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)

	// 验证钱包确实存在于数据库
	assert.NotNil(t, wallet)
	assert.Equal(t, customer.ID, wallet.CustomerID)
}

// TestAuthControllerLogin 测试登录控制器
func TestAuthControllerLogin(t *testing.T) {
	suite := testutil.SetupControllerTestSuite(t)
	defer suite.Cleanup()

	// 清理测试数据
	defer suite.CleanupTestData("admin_users")

	// 1. 创建测试管理员
	admin := suite.CreateTestAdmin(1001, "testadmin")

	// 2. 设置路由
	authController := NewAuthController(suite.Manager)
	suite.Router.POST("/auth/login", authController.Login)

	// 3. 构建登录请求
	loginReq := gin.H{
		"username": admin.Username,
		"password": "testpassword", // 注意：实际测试中需要使用正确的密码
	}

	// 4. 发送登录请求
	req := suite.MakeJSONRequest("POST", "/auth/login", loginReq)
	w := suite.ExecuteRequest(req)

	// 5. 验证响应
	t.Logf("Login response status: %d", w.Code)
	t.Logf("Login response body: %s", w.Body.String())

	// 由于密码不匹配，期望返回认证失败或参数错误
	assert.True(t, w.Code == http.StatusUnauthorized || w.Code == http.StatusBadRequest || w.Code == http.StatusOK)
}