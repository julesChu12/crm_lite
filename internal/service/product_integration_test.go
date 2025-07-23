package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dto"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ProductIntegrationTestSuite 是产品服务集成测试的测试套件
type ProductIntegrationTestSuite struct {
	suite.Suite
	resManager *resource.Manager
	service    *ProductService
	db         *resource.DBResource
}

// SetupSuite 在测试套件开始时运行
func (s *ProductIntegrationTestSuite) SetupSuite() {
	if testResManager == nil {
		s.T().Fatal("Test resource manager was not initialized in TestMain")
	}
	s.resManager = testResManager

	db, err := resource.Get[*resource.DBResource](s.resManager, resource.DBServiceKey)
	if err != nil {
		s.T().Fatalf("Failed to get db resource for test suite: %v", err)
	}
	s.db = db

	productRepo := NewProductRepo(s.db.DB)
	s.service = NewProductService(productRepo)
}

// TearDownSuite 在测试套件结束时运行
func (s *ProductIntegrationTestSuite) TearDownSuite() {
	// TearDown 由 TestMain 统一处理
}

// BeforeTest 在每个测试方法运行前运行，用于清理数据库
func (s *ProductIntegrationTestSuite) BeforeTest(suiteName, testName string) {
	s.db.DB.Exec("DELETE FROM products")
}

// TestRunner 是启动测试套件的入口
func TestProductIntegration(t *testing.T) {
	suite.Run(t, new(ProductIntegrationTestSuite))
}

// TestCreateProductIntegration 是一个端到端的集成测试用例
func (s *ProductIntegrationTestSuite) TestCreateProductIntegration() {
	ctx := context.Background()
	req := &dto.ProductCreateRequest{
		Name:        "Laptop Pro",
		Description: "A powerful new laptop",
		Price:       1999.99,
		SKU:         "LP-PRO-001",
		Stock:       50,
	}

	// 1. 调用服务创建产品
	createdProduct, err := s.service.CreateProduct(ctx, req)
	s.NoError(err)
	s.NotNil(createdProduct)
	s.Equal("LP-PRO-001", createdProduct.SKU)

	// 2. 直接从数据库验证数据
	var dbProduct model.Product
	result := s.db.DB.Where("category = ?", "LP-PRO-001").First(&dbProduct)

	s.NoError(result.Error)
	s.Equal(createdProduct.ID, dbProduct.ID)
	s.Equal("Laptop Pro", dbProduct.Name)
	s.Equal(int32(50), dbProduct.StockQuantity)
}
