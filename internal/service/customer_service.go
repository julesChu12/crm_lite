package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"strconv"
	"strings"

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

func (s *CustomerService) toCustomerResponse(customer *model.Customer) *dto.CustomerResponse {
	return &dto.CustomerResponse{
		ID:        customer.ID,
		Name:      customer.Name,
		Phone:     customer.Phone,
		Email:     customer.Email,
		CreatedAt: utils.FormatTime(customer.CreatedAt),
		UpdatedAt: utils.FormatTime(customer.UpdatedAt),
	}
}

// CreateCustomer 创建客户
func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CustomerCreateRequest) (*dto.CustomerResponse, error) {
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
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
	}
	err := s.q.Customer.WithContext(ctx).Create(customer)
	if err != nil {
		return nil, err
	}
	return s.toCustomerResponse(customer), nil
}

// ListCustomers 获取客户列表（可扩展分页）
func (s *CustomerService) ListCustomers(ctx context.Context, req *dto.CustomerListRequest) (*dto.CustomerListResponse, error) {
	q := s.q.Customer.WithContext(ctx).Where(s.q.Customer.DeletedAt.IsNull())

	if len(req.IDs) > 0 {
		q = q.Where(s.q.Customer.ID.In(req.IDs...))
	} else {
		// 构建筛选条件
		if req.Name != "" {
			q = q.Where(s.q.Customer.Name.Like("%" + req.Name + "%"))
		}
		if req.Phone != "" {
			q = q.Where(s.q.Customer.Phone.Eq(req.Phone))
		}
		if req.Email != "" {
			q = q.Where(s.q.Customer.Email.Eq(req.Email))
		}
	}

	// 构建排序条件
	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			field := parts[0]
			order := parts[1]
			if col, ok := s.q.Customer.GetFieldByName(field); ok {
				if order == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
		// 默认按创建时间降序
		q = q.Order(s.q.Customer.CreatedAt.Desc())
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// 分页查询
	var customers []*model.Customer
	var errFind error
	if len(req.IDs) == 0 {
		customers, errFind = q.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find()
	} else {
		customers, errFind = q.Find()
	}

	if errFind != nil {
		return nil, errFind
	}

	// 转换成 DTO
	customerResponses := make([]*dto.CustomerResponse, 0, len(customers))
	for _, c := range customers {
		customerResponses = append(customerResponses, s.toCustomerResponse(c))
	}

	return &dto.CustomerListResponse{
		Total:     total,
		Customers: customerResponses,
	}, nil
}

// GetCustomerByID 获取单个客户
func (s *CustomerService) GetCustomerByID(ctx context.Context, id string) (*dto.CustomerResponse, error) {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return nil, ErrCustomerNotFound
	}
	customer, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(idNum), s.q.Customer.DeletedAt.IsNull()).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return s.toCustomerResponse(customer), nil
}

// UpdateCustomer 更新客户
func (s *CustomerService) UpdateCustomer(ctx context.Context, id string, req *dto.CustomerUpdateRequest) error {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return ErrCustomerNotFound
	}

	// 检查 phone 唯一性（排除当前客户和软删除的记录）
	if req.Phone != "" {
		count, err := s.q.Customer.WithContext(ctx).
			Where(s.q.Customer.Phone.Eq(req.Phone), s.q.Customer.ID.Neq(idNum)).
			Where(s.q.Customer.DeletedAt.IsNull()).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrPhoneAlreadyExists
		}
	}

	// 使用 map 构建更新，以支持零值更新
	updates := make(map[string]interface{})
	if req.Name != "" { // 如果希望允许清空，可以去掉这个 if
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		return nil // 没有需要更新的字段
	}

	result, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(idNum)).Updates(updates)
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
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return ErrCustomerNotFound
	}
	result, err := s.q.Customer.WithContext(ctx).Where(s.q.Customer.ID.Eq(idNum)).Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}
