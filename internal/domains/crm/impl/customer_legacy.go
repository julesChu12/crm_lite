package impl

import (
	"context"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ICustomerRepo 与 legacy 等价
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

type customerRepo struct {
	q *query.Query
}

func NewCustomerRepo(db *gorm.DB) ICustomerRepo {
	return &customerRepo{q: query.Use(db)}
}

func (r *customerRepo) FindByPhoneUnscoped(ctx context.Context, phone string) (*model.Customer, error) {
	return r.q.Customer.WithContext(ctx).Unscoped().Where(r.q.Customer.Phone.Eq(phone)).First()
}

func (r *customerRepo) Update(ctx context.Context, customer *model.Customer) error {
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
		"deleted_at":  nil,
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
	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			if col, ok := r.q.Customer.GetFieldByName(parts[0]); ok {
				if parts[1] == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
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

// 使用最小钱包端口（沿用 legacy 端口定义）
type ICustomerWalletPort interface {
	CreateWallet(ctx context.Context, customerID int64, walletType string) (*model.Wallet, error)
	GetWalletByCustomerID(ctx context.Context, customerID int64) (*dto.WalletResponse, error)
}

type CustomerService struct {
	repo      ICustomerRepo
	walletSvc ICustomerWalletPort
}

func NewCustomerService(repo ICustomerRepo, walletSvc ICustomerWalletPort) *CustomerService {
	return &CustomerService{repo: repo, walletSvc: walletSvc}
}

func (s *CustomerService) toCustomerResponse(customer *model.Customer) *dto.CustomerResponse {
	birthday := ""
	if customer != nil && !customer.Birthday.IsZero() {
		birthday = customer.Birthday.Format("2006-01-02")
	}
	var tags []string
	if customer.Tags != "" {
		_ = json.Unmarshal([]byte(customer.Tags), &tags)
	}
	sourceDisplay := dto.SourceMap[customer.Source]
	if sourceDisplay == "" {
		sourceDisplay = customer.Source
	}
	genderDisplay := dto.GenderMap[customer.Gender]
	if genderDisplay == "" {
		genderDisplay = customer.Gender
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

func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CustomerCreateRequest) (*dto.CustomerResponse, error) {
	var customerToReturn *model.Customer
	isNewCreation := false
	if req.Phone != "" {
		existingCustomer, err := s.repo.FindByPhoneUnscoped(ctx, req.Phone)
		fmt.Println("existingCustomer", existingCustomer)
		if err == nil {
			if existingCustomer.DeletedAt.Valid {
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
				return nil, service.ErrPhoneAlreadyExists
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if customerToReturn == nil {
		customer := &model.Customer{
			Name:       req.Name,
			Phone:      req.Phone,
			Email:      req.Email,
			Gender:     req.Gender,
			Level:      req.Level,
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
	return &dto.CustomerListResponse{Total: total, Customers: customerResponses}, nil
}

func (s *CustomerService) GetCustomerByID(ctx context.Context, id string) (*dto.CustomerResponse, error) {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return nil, service.ErrCustomerNotFound
	}
	customer, err := s.repo.FindByID(ctx, idNum)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrCustomerNotFound
		}
		return nil, err
	}
	resp := s.toCustomerResponse(customer)
	if wallet, errW := s.walletSvc.GetWalletByCustomerID(ctx, customer.ID); errW == nil {
		resp.WalletBalance = wallet.Balance
	}
	return resp, nil
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, id string, req *dto.CustomerUpdateRequest) error {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return service.ErrCustomerNotFound
	}
	if req.Phone != "" {
		count, err := s.repo.FindByPhoneExcludingID(ctx, req.Phone, idNum)
		if err != nil {
			return err
		}
		if count > 0 {
			return service.ErrPhoneAlreadyExists
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
	// 仅当 Tags 提供时才更新（nil 表示不更新；空数组表示清空）
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
		return service.ErrCustomerNotFound
	}
	return nil
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, id string) error {
	idNum, errConv := strconv.ParseInt(id, 10, 64)
	if errConv != nil {
		return service.ErrCustomerNotFound
	}
	rowsAffected, err := s.repo.Delete(ctx, idNum)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return service.ErrCustomerNotFound
	}
	return nil
}
