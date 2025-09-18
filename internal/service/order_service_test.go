package service

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dto"
	"fmt"
    "os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type OrderServiceSuite struct {
	suite.Suite
	db      *gorm.DB
	service *OrderService
	ctx     context.Context
}

func (s *OrderServiceSuite) SetupSuite() {
	var err error
	s.db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.Require().NoError(err)

	// Manually create tables to avoid SQLite compatibility issues
	err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100),
			phone VARCHAR(20),
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	s.Require().NoError(err)

	err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			price DECIMAL(10,2) NOT NULL,
			category VARCHAR(50),
			stock_quantity INTEGER DEFAULT 0,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	s.Require().NoError(err)

	err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_no VARCHAR(50) NOT NULL,
			customer_id INTEGER NOT NULL,
			order_date DATETIME NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
			final_amount DECIMAL(12,2) NOT NULL,
			remark TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	s.Require().NoError(err)

	err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			product_name VARCHAR(100),
			quantity INTEGER NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			final_price DECIMAL(12,2) NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	s.Require().NoError(err)

	rm := resource.NewManager()
	dbResource := resource.NewDBResource(config.DBOptions{})
	dbResource.DB = s.db
	_ = rm.Register(resource.DBServiceKey, dbResource)

	s.service = NewOrderService(rm)
	s.ctx = context.Background()
}

func (s *OrderServiceSuite) TearDownSuite() {
	sqlDB, _ := s.db.DB()
	sqlDB.Close()
}

func (s *OrderServiceSuite) SetupTest() {
	// Clean up database before each test
	s.db.Exec("DELETE FROM order_items")
	s.db.Exec("DELETE FROM orders")
	s.db.Exec("DELETE FROM products")
	s.db.Exec("DELETE FROM customers")
}

func TestOrderService(t *testing.T) {
    if os.Getenv("RUN_DB_TESTS") != "1" {
        t.Skip("Skipping integration-like order tests because RUN_DB_TESTS is not set")
        return
    }
    suite.Run(t, new(OrderServiceSuite))
}

// Helper method to create test customer
func (s *OrderServiceSuite) createTestCustomer() *model.Customer {
	customer := &model.Customer{
		Name:  "Test Customer",
		Email: "test@example.com",
		Phone: "1234567890",
	}
	s.Require().NoError(s.db.Create(customer).Error)
	return customer
}

// Helper method to create test product
func (s *OrderServiceSuite) createTestProduct(name string, price float64) *model.Product {
	product := &model.Product{
		Name:          name,
		Description:   "Test product description",
		Price:         price,
		Category:      "TEST-SKU-" + name,
		StockQuantity: 100,
	}
	s.Require().NoError(s.db.Create(product).Error)
	return product
}

func (s *OrderServiceSuite) TestCreateOrder_Success() {
	// Create test data
	customer := s.createTestCustomer()
	product1 := s.createTestProduct("Product1", 10.0)
	product2 := s.createTestProduct("Product2", 20.0)

	// Create order request
	req := &dto.OrderCreateRequest{
		CustomerID: customer.ID,
		OrderDate:  time.Now(),
		Status:     "pending",
		Remark:     "Test order",
		Items: []*dto.OrderItemRequest{
			{
				ProductID: product1.ID,
				Quantity:  2,
				UnitPrice: 10.0,
			},
			{
				ProductID: product2.ID,
				Quantity:  1,
				UnitPrice: 20.0,
			},
		},
	}

	// Create order
	resp, err := s.service.CreateOrder(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify response
	s.Equal(customer.ID, resp.CustomerID)
	s.Equal("pending", resp.Status)
	s.Equal("Test order", resp.Remark)
	s.Equal(float64(40), resp.TotalAmount) // 2*10 + 1*20 = 40
	s.Equal(float64(40), resp.FinalAmount)
	s.Len(resp.Items, 2)
	s.NotEmpty(resp.OrderNo)

	// Verify order items
	s.Equal(product1.ID, resp.Items[0].ProductID)
	s.Equal(2, resp.Items[0].Quantity)
	s.Equal(float64(10), resp.Items[0].UnitPrice)
	s.Equal(float64(20), resp.Items[0].FinalPrice)

	s.Equal(product2.ID, resp.Items[1].ProductID)
	s.Equal(1, resp.Items[1].Quantity)
	s.Equal(float64(20), resp.Items[1].UnitPrice)
	s.Equal(float64(20), resp.Items[1].FinalPrice)
}

func (s *OrderServiceSuite) TestCreateOrder_CustomerNotFound() {
	product := s.createTestProduct("Product1", 10.0)

	req := &dto.OrderCreateRequest{
		CustomerID: 99999, // Non-existent customer
		OrderDate:  time.Now(),
		Items: []*dto.OrderItemRequest{
			{
				ProductID: product.ID,
				Quantity:  1,
				UnitPrice: 10.0,
			},
		},
	}

	resp, err := s.service.CreateOrder(s.ctx, req)
	s.Require().Error(err)
	s.Nil(resp)
	s.Equal(ErrCustomerNotFound, err)
}

func (s *OrderServiceSuite) TestCreateOrder_ProductNotFound() {
	customer := s.createTestCustomer()

	req := &dto.OrderCreateRequest{
		CustomerID: customer.ID,
		OrderDate:  time.Now(),
		Items: []*dto.OrderItemRequest{
			{
				ProductID: 99999, // Non-existent product
				Quantity:  1,
				UnitPrice: 10.0,
			},
		},
	}

	resp, err := s.service.CreateOrder(s.ctx, req)
	s.Require().Error(err)
	s.Nil(resp)
	s.Equal(ErrProductNotFound, err)
}

func (s *OrderServiceSuite) TestGetOrderByID_Success() {
	// Create test data
	customer := s.createTestCustomer()
	product := s.createTestProduct("Product1", 10.0)

	// Create order first
	req := &dto.OrderCreateRequest{
		CustomerID: customer.ID,
		OrderDate:  time.Now(),
		Items: []*dto.OrderItemRequest{
			{
				ProductID: product.ID,
				Quantity:  1,
				UnitPrice: 10.0,
			},
		},
	}
	createdOrder, err := s.service.CreateOrder(s.ctx, req)
	s.Require().NoError(err)

	// Get order by ID
	resp, err := s.service.GetOrderByID(s.ctx, fmt.Sprintf("%d", createdOrder.ID))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify response
	s.Equal(createdOrder.ID, resp.ID)
	s.Equal(customer.ID, resp.CustomerID)
	s.Len(resp.Items, 1)
}

func (s *OrderServiceSuite) TestGetOrderByID_NotFound() {
	resp, err := s.service.GetOrderByID(s.ctx, "99999")
	s.Require().Error(err)
	s.Nil(resp)
	s.Equal(ErrOrderNotFound, err)
}

func (s *OrderServiceSuite) TestGetOrderByID_InvalidID() {
	resp, err := s.service.GetOrderByID(s.ctx, "invalid")
	s.Require().Error(err)
	s.Nil(resp)
	s.Equal(ErrOrderNotFound, err)
}