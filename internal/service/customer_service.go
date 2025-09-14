package service

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ICustomerRepo 定义了客户数据仓库的接口
type ICustomerRepo interface {
	FindByPhoneUnscoped(ctx context.Context, phone string) (*model.Customer, error)
	Update(ctx context.Context, customer *model.Customer) error
	Create(ctx context.Context, customer *model.Customer) error
	List(ctx context.Context, req *dto.CustomerListRequest) ([]*model.Customer, int64, error)
	FindByID(ctx context.Context, id int64) (*model.Customer, error)
	Updates(ctx context.Context, id int64, updates map[string]interface{}) (int64, error)
	Delete(ctx context.Context, id int64) (int64, error)
	FindByPhoneExcludingID(ctx context.Context, phone string, id int64) (int64, error)
}

// customerRepo 实现了 ICustomerRepo 接口，封装了 gorm 操作
type customerRepo struct {
	q *query.Query
}

// NewCustomerRepo 创建一个新的 customerRepo
func NewCustomerRepo(db *gorm.DB) ICustomerRepo {
	return &customerRepo{
		q: query.Use(db),
	}
}

func (r *customerRepo) FindByPhoneUnscoped(ctx context.Context, phone string) (*model.Customer, error) {
	return r.q.Customer.WithContext(ctx).Unscoped().Where(r.q.Customer.Phone.Eq(phone)).First()
}

func (r *customerRepo) Update(ctx context.Context, customer *model.Customer) error {
	// 使用 map 来构建更新，这样可以更新零值字段
	updates := map[string]interface{}{
		"name":        customer.Name,
		"email":       customer.Email,
		"gender":      customer.Gender,
		"level":       customer.Level,
		"tags":        customer.Tags,
		"note":        customer.Note,
		"source":      customer.Source,
		"assigned_to": customer.AssignedTo,
		"birthday":    customer.Birthday,
		"deleted_at":  nil, // 恢复记录
	}
	_, err := r.q.Customer.WithContext(ctx).Unscoped().Where(r.q.Customer.ID.Eq(customer.ID)).Updates(updates)
	return err
}

func (r *customerRepo) Create(ctx context.Context, customer *model.Customer) error {
	return r.q.Customer.WithContext(ctx).Create(customer)
}

func (r *customerRepo) List(ctx context.Context, req *dto.CustomerListRequest) ([]*model.Customer, int64, error) {
	q := r.q.Customer.WithContext(ctx).Where(r.q.Customer.DeletedAt.IsNull())

	if len(req.IDs) > 0 {
		q = q.Where(r.q.Customer.ID.In(req.IDs...))
	} else {
		// 构建筛选条件
		if req.Name != "" {
			q = q.Where(r.q.Customer.Name.Like("%" + req.Name + "%"))
		}
		if req.Phone != "" {
			q = q.Where(r.q.Customer.Phone.Eq(req.Phone))
		}
		if req.Email != "" {
			q = q.Where(r.q.Customer.Email.Eq(req.Email))
		}
	}

	// 构建排序条件
	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			field := parts[0]
			order := parts[1]
			if col, ok := r.q.Customer.GetFieldByName(field); ok {
				if order == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
		// 默认按创建时间降序
		q = q.Order(r.q.Customer.CreatedAt.Desc())
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	var customers []*model.Customer
	if len(req.IDs) == 0 && req.PageSize > 0 {
		customers, err = q.Limit(req.PageSize).Offset((req.Page - 1) * req.PageSize).Find()
	} else {
		customers, err = q.Find()
	}

	return customers, total, err
}

func (r *customerRepo) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
	return r.q.Customer.WithContext(ctx).Where(r.q.Customer.ID.Eq(id), r.q.Customer.DeletedAt.IsNull()).First()
}

func (r *customerRepo) Updates(ctx context.Context, id int64, updates map[string]interface{}) (int64, error) {
	result, err := r.q.Customer.WithContext(ctx).Where(r.q.Customer.ID.Eq(id)).Updates(updates)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func (r *customerRepo) Delete(ctx context.Context, id int64) (int64, error) {
	result, err := r.q.Customer.WithContext(ctx).Where(r.q.Customer.ID.Eq(id)).Delete()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func (r *customerRepo) FindByPhoneExcludingID(ctx context.Context, phone string, id int64) (int64, error) {
	return r.q.Customer.WithContext(ctx).
		Where(r.q.Customer.Phone.Eq(phone), r.q.Customer.ID.Neq(id)).
		Where(r.q.Customer.DeletedAt.IsNull()).
		Count()
}

type CustomerService struct {
	repo      ICustomerRepo
	walletSvc ICustomerWalletPort
}

// ICustomerWalletPort 为 CustomerService 依赖的最小钱包端口（仅用于创建与查询余额）
type ICustomerWalletPort interface {
	CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error)
	GetWalletByCustomerID(ctx context.Context, customerID int64) (*dto.WalletResponse, error)
}

func NewCustomerService(repo ICustomerRepo, walletSvc ICustomerWalletPort) *CustomerService {
	return &CustomerService{
		repo:      repo,
		walletSvc: walletSvc,
	}
}

func (s *CustomerService) toCustomerResponse(customer *model.Customer) *dto.CustomerResponse {
	birthday := ""
	if customer != nil && !customer.Birthday.IsZero() {
		birthday = customer.Birthday.Format("2006-01-02")
	}

	var tags []string
	if customer.Tags != "" {
		// 忽略 unmarshal 错误，如果解析失败则返回空数组
		_ = json.Unmarshal([]byte(customer.Tags), &tags)
	}

	// Map source and gender to display values
	sourceDisplay := dto.SourceMap[customer.Source]
	if sourceDisplay == "" {
		sourceDisplay = customer.Source // Fallback to original value if not in map
	}
	genderDisplay := dto.GenderMap[customer.Gender]
	if genderDisplay == "" {
		genderDisplay = customer.Gender // Fallback
	}

	return &dto.CustomerResponse{
		ID:         customer.ID,
		Name:       customer.Name,
		Phone:      customer.Phone,
		Email:      customer.Email,
		Gender:     genderDisplay,
		Birthday:   birthday,
		Level:      customer.Level,
		Tags:       tags,
		Note:       customer.Note,
		Source:     sourceDisplay,
		AssignedTo: customer.AssignedTo,
		CreatedAt:  utils.FormatTime(customer.CreatedAt),
		UpdatedAt:  utils.FormatTime(customer.UpdatedAt),
	}
}

// CreateCustomer 创建客户
func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CustomerCreateRequest) (*dto.CustomerResponse, error) {
	var customerToReturn *model.Customer
	isNewCreation := false

	if req.Phone != "" {
		existingCustomer, err := s.repo.FindByPhoneUnscoped(ctx, req.Phone)
		fmt.Println("existingCustomer", existingCustomer)
		if err == nil {
			if existingCustomer.DeletedAt.Valid {
				// 恢复并更新
				existingCustomer.Name = req.Name
				existingCustomer.Email = req.Email
				existingCustomer.Gender = req.Gender
				existingCustomer.Level = req.Level

				tagsJSON, err := json.Marshal(req.Tags)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tags: %w", err)
				}
				existingCustomer.Tags = string(tagsJSON)

				existingCustomer.Note = req.Note
				existingCustomer.Source = req.Source
				existingCustomer.AssignedTo = req.AssignedTo

				if req.Birthday != "" {
					birthday, err := time.Parse("2006-01-02", req.Birthday)
					if err != nil {
						return nil, fmt.Errorf("invalid birthday format: %w", err)
					}
					existingCustomer.Birthday = birthday
				}

				if err := s.repo.Update(ctx, existingCustomer); err != nil {
					return nil, err
				}
				customerToReturn = existingCustomer
			} else {
				return nil, ErrPhoneAlreadyExists
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if customerToReturn == nil {
		customer := &model.Customer{
			Name:   req.Name,
			Phone:  req.Phone,
			Email:  req.Email,
			Gender: req.Gender,
			Level:  req.Level,
			// Tags:       strings.Join(req.Tags, ","), //
			Note:       req.Note,
			Source:     req.Source,
			AssignedTo: req.AssignedTo,
		}

		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		customer.Tags = string(tagsJSON)

		if req.Birthday != "" {
			birthday, err := time.Parse("2006-01-02", req.Birthday)
			if err != nil {
				return nil, fmt.Errorf("invalid birthday format: %w", err)
			}
			customer.Birthday = birthday
		}
		if err := s.repo.Create(ctx, customer); err != nil {
			return nil, err
		}
		customerToReturn = customer
		isNewCreation = true
	}

	if _, err := s.walletSvc.CreateWallet(ctx, customerToReturn.ID, "balance"); err != nil {
		if isNewCreation {
			return nil, fmt.Errorf("failed to create wallet for new customer: %w", err)
		}
	}

	return s.toCustomerResponse(customerToReturn), nil
}

// ListCustomers 获取客户列表（可扩展分页）
func (s *CustomerService) ListCustomers(ctx context.Context, req *dto.CustomerListRequest) (*dto.CustomerListResponse, error) {
	customers, total, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	customerResponses := make([]*dto.CustomerResponse, 0, len(customers))
	for _, c := range customers {
		resp := s.toCustomerResponse(c)
		if wallet, errW := s.walletSvc.GetWalletByCustomerID(ctx, c.ID); errW == nil {
			resp.WalletBalance = wallet.Balance
		}
		customerResponses = append(customerResponses, resp)
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
	customer, err := s.repo.FindByID(ctx, idNum)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	resp := s.toCustomerResponse(customer)
	if wallet, errW := s.walletSvc.GetWalletByCustomerID(ctx, customer.ID); errW == nil {
		resp.WalletBalance = wallet.Balance
	}
	return resp, nil
}

// UpdateCustomer 更新客户
func (s *CustomerService) UpdateCustomer(ctx context.Context, id string, req *dto.CustomerUpdateRequest) error {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return ErrCustomerNotFound
	}

	if req.Phone != "" {
		count, err := s.repo.FindByPhoneExcludingID(ctx, req.Phone, idNum)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrPhoneAlreadyExists
		}
	}

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
	// 注意：仅当请求中提供了 tags（nil 表示未提供）时才更新；空数组表示清空
	if req.Tags != nil {
		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshal tags: %w", err)
		}
		updates["tags"] = string(tagsJSON)
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
		return nil
	}

	rowsAffected, err := s.repo.Updates(ctx, idNum, updates)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
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
	rowsAffected, err := s.repo.Delete(ctx, idNum)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCustomerNotFound
	}
	return nil
}
