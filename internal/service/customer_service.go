package service

import (
	"context"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type CustomerService struct {
	q         *query.Query
	resource  *resource.Manager
	walletSvc *WalletService
}

func NewCustomerService(resManager *resource.Manager) *CustomerService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for CustomerService: " + err.Error())
	}
	return &CustomerService{
		q:         query.Use(db.DB),
		resource:  resManager,
		walletSvc: NewWalletService(resManager), // 直接实例化
	}
}

func (s *CustomerService) toCustomerResponse(customer *model.Customer) *dto.CustomerResponse {
	birthday := ""
	if !customer.Birthday.IsZero() {
		birthday = customer.Birthday.Format("2006-01-02")
	}
	return &dto.CustomerResponse{
		ID:         customer.ID,
		Name:       customer.Name,
		Phone:      customer.Phone,
		Email:      customer.Email,
		Gender:     customer.Gender,
		Birthday:   birthday,
		Level:      customer.Level,
		Tags:       customer.Tags,
		Note:       customer.Note,
		Source:     customer.Source,
		AssignedTo: customer.AssignedTo,
		CreatedAt:  utils.FormatTime(customer.CreatedAt),
		UpdatedAt:  utils.FormatTime(customer.UpdatedAt),
	}
}

// CreateCustomer 创建客户
func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CustomerCreateRequest) (*dto.CustomerResponse, error) {
	var customerToReturn *model.Customer
	var isNewCreation bool

	if req.Phone != "" {
		// 使用 Unscoped 查找包括软删除在内的所有记录
		existingCustomer, err := s.q.Customer.WithContext(ctx).Unscoped().Where(s.q.Customer.Phone.Eq(req.Phone)).First()

		// 如果找到了记录
		if err == nil {
			if existingCustomer.DeletedAt.Valid {
				// 如果是软删除的记录，则恢复并更新它
				updates := map[string]interface{}{
					"name":        req.Name,
					"email":       req.Email,
					"gender":      req.Gender,
					"level":       req.Level,
					"tags":        req.Tags,
					"note":        req.Note,
					"source":      req.Source,
					"assigned_to": req.AssignedTo,
					"deleted_at":  nil, // 恢复记录
				}
				var birthday time.Time
				if req.Birthday != "" {
					birthday, err = time.Parse("2006-01-02", req.Birthday)
					if err != nil {
						return nil, fmt.Errorf("invalid birthday format: %w", err)
					}
					updates["birthday"] = birthday
				}

				if _, err := s.q.Customer.WithContext(ctx).Unscoped().Where(s.q.Customer.ID.Eq(existingCustomer.ID)).Updates(updates); err != nil {
					return nil, err
				}

				// 手动更新结构体以用于响应
				existingCustomer.Name = req.Name
				existingCustomer.Email = req.Email
				existingCustomer.Gender = req.Gender
				if req.Birthday != "" {
					existingCustomer.Birthday = birthday
				}
				existingCustomer.Level = req.Level
				existingCustomer.Tags = req.Tags
				existingCustomer.Note = req.Note
				existingCustomer.Source = req.Source
				existingCustomer.AssignedTo = req.AssignedTo
				existingCustomer.DeletedAt.Valid = false

				customerToReturn = existingCustomer
			} else {
				// 如果是活动的记录，则返回号码已存在错误
				return nil, ErrPhoneAlreadyExists
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果是其他非“未找到”的错误，则直接返回
			return nil, err
		}
	}

	if customerToReturn == nil {
		// 如果号码为空或记录未找到，则创建新客户
		customer := &model.Customer{
			Name:       req.Name,
			Phone:      req.Phone,
			Email:      req.Email,
			Gender:     req.Gender,
			Level:      req.Level,
			Tags:       req.Tags,
			Note:       req.Note,
			Source:     req.Source,
			AssignedTo: req.AssignedTo,
		}

		if req.Birthday != "" {
			birthday, err := time.Parse("2006-01-02", req.Birthday)
			if err != nil {
				return nil, fmt.Errorf("invalid birthday format: %w", err)
			}
			customer.Birthday = birthday
		}

		err := s.q.Customer.WithContext(ctx).Create(customer)
		if err != nil {
			return nil, err
		}
		customerToReturn = customer
		isNewCreation = true
	}

	// 无论是新创建还是恢复，都确保钱包存在
	if _, err := s.walletSvc.CreateWallet(ctx, customerToReturn.ID, "balance"); err != nil {
		// 如果是新创建用户时钱包创建失败，我们可能需要考虑回滚用户创建，但这会增加复杂性。
		// 在这里，我们暂时只记录或返回错误。
		if isNewCreation {
			// 对于新用户，钱包创建失败是个严重问题，可能需要返回错误
			return nil, fmt.Errorf("failed to create wallet for new customer: %w", err)
		}
		// 对于已存在的用户，可能只是日志记录一下即可
		// log.Printf("Warning: failed to ensure wallet exists for customer %d: %v", customerToReturn.ID, err)
	}

	return s.toCustomerResponse(customerToReturn), nil
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
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Gender != "" {
		updates["gender"] = req.Gender
	}
	if req.Level != "" {
		updates["level"] = req.Level
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.Note != "" {
		updates["note"] = req.Note
	}
	if req.Source != "" {
		updates["source"] = req.Source
	}
	if req.AssignedTo != 0 {
		updates["assigned_to"] = req.AssignedTo
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return fmt.Errorf("invalid birthday format: %w", err)
		}
		updates["birthday"] = birthday
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
