package service

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerService struct {
	q *query.Query
}

func NewCustomerService(db *gorm.DB) *CustomerService {
	return &CustomerService{q: query.Use(db)}
}

// CreateCustomer 创建客户
func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CustomerCreateRequest) (*model.Customer, error) {
	customer := &model.Customer{
		ID:    uuid.New().String(),
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
	}
	err := s.q.Customer.WithContext(ctx).Create(customer)
	return customer, err
}

// ListCustomers 获取客户列表（可扩展分页）
func (s *CustomerService) ListCustomers(ctx context.Context) ([]*model.Customer, error) {
	return s.q.Customer.WithContext(ctx).Find()
}

// GetCustomerByID 获取单个客户
func (s *CustomerService) GetCustomerByID(ctx context.Context, id string) (*model.Customer, error) {
	return s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(id)).First()
}

// UpdateCustomer 更新客户
func (s *CustomerService) UpdateCustomer(ctx context.Context, id string, req *dto.CustomerUpdateRequest) error {
	_, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(id)).Updates(model.Customer{
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
	})
	return err
}

// DeleteCustomer 删除客户
func (s *CustomerService) DeleteCustomer(ctx context.Context, id string) error {
	_, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(id)).Delete()
	return err
}
