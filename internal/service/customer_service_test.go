package service

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dto"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockCustomerRepo is a mock type for the ICustomerRepo interface
type MockCustomerRepo struct {
	mock.Mock
}

func (m *MockCustomerRepo) FindByPhoneUnscoped(ctx context.Context, phone string) (*model.Customer, error) {
	args := m.Called(ctx, phone)
	var r0 *model.Customer
	if args.Get(0) != nil {
		r0 = args.Get(0).(*model.Customer)
	}
	return r0, args.Error(1)
}

func (m *MockCustomerRepo) Update(ctx context.Context, customer *model.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepo) Create(ctx context.Context, customer *model.Customer) error {
	args := m.Called(ctx, customer)
	// Set the ID for the created customer, simulating database behavior
	if customer != nil {
		customer.ID = 1
	}
	return args.Error(0)
}

func (m *MockCustomerRepo) List(ctx context.Context, req *dto.CustomerListRequest) ([]*model.Customer, int64, error) {
	args := m.Called(ctx, req)
	// Handle nil case for slice
	var r0 []*model.Customer
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*model.Customer)
	}
	return r0, args.Get(1).(int64), args.Error(2)
}

func (m *MockCustomerRepo) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
	args := m.Called(ctx, id)
	var r0 *model.Customer
	if args.Get(0) != nil {
		r0 = args.Get(0).(*model.Customer)
	}
	return r0, args.Error(1)
}

func (m *MockCustomerRepo) Updates(ctx context.Context, id int64, updates map[string]interface{}) (int64, error) {
	args := m.Called(ctx, id, updates)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCustomerRepo) Delete(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCustomerRepo) FindByPhoneExcludingID(ctx context.Context, phone string, id int64) (int64, error) {
	args := m.Called(ctx, phone, id)
	return args.Get(0).(int64), args.Error(1)
}

// MockWalletService is a mock for IWalletService
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error) {
	args := m.Called(ctx, customerID, walletType)
	var r0 *model.Wallet
	if args.Get(0) != nil {
		r0 = args.Get(0).(*model.Wallet)
	}
	return r0, args.Error(1)
}

func (m *MockWalletService) CreateTransaction(ctx context.Context, customerID int64, operatorID int64, req *dto.WalletTransactionRequest) error {
	args := m.Called(ctx, customerID, operatorID, req)
	return args.Error(0)
}

func (m *MockWalletService) GetWalletByCustomerID(ctx context.Context, customerID int64) (*dto.WalletResponse, error) {
	args := m.Called(ctx, customerID)
	var r0 *dto.WalletResponse
	if args.Get(0) != nil {
		r0 = args.Get(0).(*dto.WalletResponse)
	}
	return r0, args.Error(1)
}

func (m *MockWalletService) GetTransactions(ctx context.Context, customerID int64, page, limit int) ([]*dto.WalletTransactionResponse, int64, error) {
	args := m.Called(ctx, customerID, page, limit)
	var r0 []*dto.WalletTransactionResponse
	if args.Get(0) != nil {
		r0 = args.Get(0).([]*dto.WalletTransactionResponse)
	}
	return r0, args.Get(1).(int64), args.Error(2)
}

func (m *MockWalletService) ProcessRefund(ctx context.Context, customerID int64, operatorID int64, req *dto.WalletRefundRequest) error {
	args := m.Called(ctx, customerID, operatorID, req)
	return args.Error(0)
}

func TestCreateCustomer_NewCustomer(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()
	req := &dto.CustomerCreateRequest{
		Name:  "Test User",
		Phone: "1234567890",
		Email: "test@example.com",
	}

	// Expectations for a completely new customer
	mockRepo.On("FindByPhoneUnscoped", ctx, req.Phone).Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Customer")).Return(nil)
	mockWalletSvc.On("CreateWallet", ctx, int64(1), "balance").Return(&model.Wallet{ID: 1}, nil)

	// Execute
	resp, err := customerService.CreateCustomer(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test User", resp.Name)
	assert.Equal(t, "1234567890", resp.Phone)

	// Verify that all expectations were met
	mockRepo.AssertExpectations(t)
	mockWalletSvc.AssertExpectations(t)
}

func TestCreateCustomer_ExistingButDeletedCustomer(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()
	req := &dto.CustomerCreateRequest{
		Name:  "Updated User",
		Phone: "1234567890",
	}

	existingCustomer := &model.Customer{
		ID:        1,
		Phone:     "1234567890",
		DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
	}

	// Expectations
	mockRepo.On("FindByPhoneUnscoped", ctx, req.Phone).Return(existingCustomer, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Customer")).Return(nil)
	mockWalletSvc.On("CreateWallet", ctx, int64(1), "balance").Return(&model.Wallet{ID: 1}, nil)

	// Execute
	resp, err := customerService.CreateCustomer(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated User", resp.Name)
	assert.Equal(t, int64(1), resp.ID)

	// Verify
	mockRepo.AssertExpectations(t)
	mockWalletSvc.AssertExpectations(t)
}

func TestCreateCustomer_ExistingActiveCustomer(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()
	req := &dto.CustomerCreateRequest{
		Name:  "Test User",
		Phone: "1234567890",
	}

	existingCustomer := &model.Customer{
		ID:    1,
		Phone: "1234567890",
	}

	// Expectation: FindByPhoneUnscoped returns an active customer
	mockRepo.On("FindByPhoneUnscoped", ctx, req.Phone).Return(existingCustomer, nil)

	// Execute
	_, err := customerService.CreateCustomer(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPhoneAlreadyExists))

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestListCustomers_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()
	req := &dto.CustomerListRequest{Page: 1, PageSize: 10}
	customers := []*model.Customer{
		{ID: 1, Name: "Customer 1"},
		{ID: 2, Name: "Customer 2"},
	}

	// Expectations
	mockRepo.On("List", ctx, req).Return(customers, int64(2), nil)
	mockWalletSvc.On("GetWalletByCustomerID", ctx, int64(1)).Return(&dto.WalletResponse{Balance: 100.0}, nil)
	mockWalletSvc.On("GetWalletByCustomerID", ctx, int64(2)).Return(&dto.WalletResponse{Balance: 200.0}, nil)

	// Execute
	resp, err := customerService.ListCustomers(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.Customers, 2)
	// WalletBalance 由内部逻辑填充，若需要可额外断言

	// Verify
	mockRepo.AssertExpectations(t)
	mockWalletSvc.AssertExpectations(t)
}

func TestGetCustomerByID_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()
	customer := &model.Customer{ID: 1, Name: "Test Customer"}

	// Expectations
	mockRepo.On("FindByID", ctx, int64(1)).Return(customer, nil)
	mockWalletSvc.On("GetWalletByCustomerID", ctx, int64(1)).Return(&dto.WalletResponse{Balance: 50.0}, nil)

	// Execute
	resp, err := customerService.GetCustomerByID(ctx, "1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Customer", resp.Name)
	// WalletBalance 字段已在服务内部填充，基础断言即可

	// Verify
	mockRepo.AssertExpectations(t)
	mockWalletSvc.AssertExpectations(t)
}

func TestGetCustomerByID_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()

	// Expectations
	mockRepo.On("FindByID", ctx, int64(999)).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	_, err := customerService.GetCustomerByID(ctx, "999")

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCustomerNotFound))

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestUpdateCustomer_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()
	req := &dto.CustomerUpdateRequest{Name: "Updated Name"}
	updates := map[string]interface{}{"name": "Updated Name"}

	// Expectations
	mockRepo.On("Updates", ctx, int64(1), updates).Return(int64(1), nil)
	// The service no longer returns the updated customer, so no FindByID call is expected.

	// Execute
	err := customerService.UpdateCustomer(ctx, "1", req)

	// Assert
	assert.NoError(t, err)

	// Verify
	mockRepo.AssertExpectations(t)
}

func TestDeleteCustomer_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockCustomerRepo)
	mockWalletSvc := new(MockWalletService)
	customerService := NewCustomerService(mockRepo, mockWalletSvc)

	ctx := context.Background()

	// Expectations
	mockRepo.On("Delete", ctx, int64(1)).Return(int64(1), nil)

	// Execute
	err := customerService.DeleteCustomer(ctx, "1")

	// Assert
	assert.NoError(t, err)

	// Verify
	mockRepo.AssertExpectations(t)
}
