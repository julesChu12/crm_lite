package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerService struct {
	q        *query.Query
	resource *resource.Manager
}

func NewCustomerService(resManager *resource.Manager) *CustomerService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for CustomerService: " + err.Error())
	}
	return &CustomerService{
		q:        query.Use(db.DB),
		resource: resManager,
	}
}

// CreateCustomer 创建客户
func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CustomerCreateRequest) (*model.Customer, error) {
	// 检查 phone 唯一性（排除软删除的记录）
	if req.Phone != "" {
		count, err := s.q.Customer.WithContext(ctx).
			Where(s.q.Customer.Phone.Eq(req.Phone)).
			Where(s.q.Customer.DeletedAt.IsNull()).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrPhoneAlreadyExists
		}
	}

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
	customer, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return customer, nil
}

// UpdateCustomer 更新客户
func (s *CustomerService) UpdateCustomer(ctx context.Context, id string, req *dto.CustomerUpdateRequest) error {
	// 检查 phone 唯一性（排除当前客户和软删除的记录）
	if req.Phone != "" {
		count, err := s.q.Customer.WithContext(ctx).
			Where(s.q.Customer.Phone.Eq(req.Phone), s.q.Customer.ID.Neq(id)).
			Where(s.q.Customer.DeletedAt.IsNull()).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrPhoneAlreadyExists
		}
	}

	result, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(id)).Updates(model.Customer{
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
	})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}

// DeleteCustomer 删除客户
func (s *CustomerService) DeleteCustomer(ctx context.Context, id string) error {
	result, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}
