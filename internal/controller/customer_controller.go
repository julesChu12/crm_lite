package controller

import (
	"context"
	"crm_lite/internal/core/resource"
	billingimpl "crm_lite/internal/domains/billing/impl"
	"crm_lite/internal/domains/crm"
	crmimpl "crm_lite/internal/domains/crm/impl"
	"crm_lite/internal/dto"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	customerService interface {
		CreateCustomerLegacy(ctx context.Context, req *crm.CustomerCreateRequest) (*crm.CustomerResponse, error)
		ListCustomersLegacy(ctx context.Context, req *crm.CustomerListRequest) (*crm.CustomerListResponse, error)
		GetCustomerByIDLegacy(ctx context.Context, id string) (*crm.CustomerResponse, error)
		UpdateCustomerLegacy(ctx context.Context, id string, req *crm.CustomerUpdateRequest) error
		DeleteCustomerLegacy(ctx context.Context, id string) error
	}
}

func NewCustomerController(resManager *resource.Manager) *CustomerController {
	// 默认注入 domains 实现；如需回滚，可改回旧 service 构造
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for CustomerController: " + err.Error())
	}
	// 使用新的billing域服务
	billingSvc := billingimpl.NewBillingService(dbRes.DB)
	domainSvc := crmimpl.NewCRMServiceWithBilling(dbRes.DB, billingSvc)
	return &CustomerController{customerService: domainSvc}
}

// CreateCustomer
// @Summary      Create a new customer
// @Description  Add a new customer to the database
// @Tags         Customers
// @Accept       json
// @Produce      json
// @Param        customer  body      dto.CustomerCreateRequest  true  "Customer Create Request"
// @Success      200      {object}  resp.Response{data=dto.CustomerResponse}
// @Failure      400      {object}  resp.Response
// @Failure      500      {object}  resp.Response
// @Router       /customers [post]
func (cc *CustomerController) CreateCustomer(c *gin.Context) {
	var dtoReq dto.CustomerCreateRequest
	if err := c.ShouldBindJSON(&dtoReq); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	// 转换 DTO 为 CRM 域类型
	req := &crm.CustomerCreateRequest{
		Name:       dtoReq.Name,
		Phone:      dtoReq.Phone,
		Email:      dtoReq.Email,
		Gender:     dtoReq.Gender,
		Birthday:   dtoReq.Birthday,
		Level:      dtoReq.Level,
		Tags:       dtoReq.Tags,
		Note:       dtoReq.Note,
		Source:     dtoReq.Source,
		AssignedTo: dtoReq.AssignedTo,
	}
	customer, err := cc.customerService.CreateCustomerLegacy(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, crmimpl.ErrPhoneAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "phone number already exists")
			return
		}
		resp.Error(c, resp.CodeInternalError, "failed to create customer")
		return
	}
	resp.Success(c, customer)
}

// ListCustomers
// @Summary      List customers
// @Description  Get a list of customers with pagination, filtering, and sorting
// @Tags         Customers
// @Produce      json
// @Param        query query     dto.CustomerListRequest false "Query parameters"
// @Success      200  {object}  resp.Response{data=dto.CustomerListResponse}
// @Failure      400  {object}  resp.Response
// @Failure      500  {object}  resp.Response
// @Router       /customers [get]
func (cc *CustomerController) ListCustomers(c *gin.Context) {
	var dtoReq dto.CustomerListRequest
	if err := c.ShouldBindQuery(&dtoReq); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	// 转换 DTO 为 CRM 域类型
	req := &crm.CustomerListRequest{
		Page:     dtoReq.Page,
		PageSize: dtoReq.PageSize,
		IDs:      dtoReq.IDs,
		Name:     dtoReq.Name,
		Phone:    dtoReq.Phone,
		Email:    dtoReq.Email,
		OrderBy:  dtoReq.OrderBy,
	}
	customers, err := cc.customerService.ListCustomersLegacy(c.Request.Context(), req)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to list customers")
		return
	}
	resp.Success(c, customers)
}

// BatchGetCustomers godoc
// @Summary      Batch get customers
// @Description  Get a list of customers by their IDs
// @Tags         Customers
// @Accept       json
// @Produce      json
// @Param        query body      dto.CustomerBatchGetRequest true "Customer IDs"
// @Success      200  {object}  resp.Response{data=dto.CustomerListResponse}
// @Failure      400  {object}  resp.Response
// @Failure      500  {object}  resp.Response
// @Router       /customers/batch-get [post]
func (cc *CustomerController) BatchGetCustomers(c *gin.Context) {
	var req dto.CustomerBatchGetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	serviceReq := &crm.CustomerListRequest{
		IDs: req.IDs,
	}

	customers, err := cc.customerService.ListCustomersLegacy(c.Request.Context(), serviceReq)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to list customers")
		return
	}
	resp.Success(c, customers)
}

// GetCustomer
// @Summary      Get a single customer
// @Description  Get a single customer by its UUID
// @Tags         Customers
// @Produce      json
// @Param        id   path      string  true  "Customer ID"
// @Success      200  {object}  resp.Response{data=dto.CustomerResponse}
// @Failure      404  {object}  resp.Response
// @Failure      500  {object}  resp.Response
// @Router       /customers/{id} [get]
func (cc *CustomerController) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	customer, err := cc.customerService.GetCustomerByIDLegacy(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, crmimpl.ErrCustomerNotFound) {
			resp.Error(c, resp.CodeNotFound, "customer not found")
			return
		}
		resp.Error(c, resp.CodeInternalError, "failed to get customer")
		return
	}
	resp.Success(c, customer)
}

// UpdateCustomer
// @Summary      Update a customer
// @Description  Update an existing customer's details
// @Tags         Customers
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true  "Customer ID"
// @Param        customer body      dto.CustomerUpdateRequest  true  "Customer Update Request"
// @Success      200      {object}  resp.Response
// @Failure      400      {object}  resp.Response
// @Failure      500      {object}  resp.Response
// @Router       /customers/{id} [put]
func (cc *CustomerController) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var dtoReq dto.CustomerUpdateRequest
	if err := c.ShouldBindJSON(&dtoReq); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	// 转换 DTO 为 CRM 域类型
	req := &crm.CustomerUpdateRequest{
		Name:       dtoReq.Name,
		Phone:      dtoReq.Phone,
		Email:      dtoReq.Email,
		Gender:     dtoReq.Gender,
		Birthday:   dtoReq.Birthday,
		Level:      dtoReq.Level,
		Tags:       &dtoReq.Tags,
		Note:       dtoReq.Note,
		Source:     dtoReq.Source,
		AssignedTo: dtoReq.AssignedTo,
	}
	if err := cc.customerService.UpdateCustomerLegacy(c.Request.Context(), id, req); err != nil {
		if errors.Is(err, crmimpl.ErrCustomerNotFound) {
			resp.Error(c, resp.CodeNotFound, "customer not found")
			return
		}
		if errors.Is(err, crmimpl.ErrPhoneAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "phone number already exists")
			return
		}
		resp.Error(c, resp.CodeInternalError, "failed to update customer")
		return
	}
	resp.Success(c, nil)
}

// DeleteCustomer
// @Summary      Delete a customer
// @Description  Delete a customer by its UUID
// @Tags         Customers
// @Produce      json
// @Param        id   path      string  true  "Customer ID"
// @Success      200  {object}  resp.Response
// @Failure      500  {object}  resp.Response
// @Router       /customers/{id} [delete]
func (cc *CustomerController) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	if err := cc.customerService.DeleteCustomerLegacy(c.Request.Context(), id); err != nil {
		if errors.Is(err, crmimpl.ErrCustomerNotFound) {
			resp.Error(c, resp.CodeNotFound, "customer not found")
			return
		}
		resp.Error(c, resp.CodeInternalError, "failed to delete customer")
		return
	}
	resp.Success(c, nil)
}
