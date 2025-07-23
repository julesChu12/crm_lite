package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	customerService *service.CustomerService
}

func NewCustomerController(resManager *resource.Manager) *CustomerController {
	// 1. 从资源管理器获取数据库资源
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for CustomerController: " + err.Error())
	}
	// 2. 创建 repo
	customerRepo := service.NewCustomerRepo(db.DB)
	// 3. 创建依赖的服务
	walletSvc := service.NewWalletService(resManager)
	// 4. 注入 repo 和依赖服务来创建 service
	customerSvc := service.NewCustomerService(customerRepo, walletSvc)

	return &CustomerController{customerService: customerSvc}
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
	var req dto.CustomerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	customer, err := cc.customerService.CreateCustomer(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
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
	var req dto.CustomerListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	customers, err := cc.customerService.ListCustomers(c.Request.Context(), &req)
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

	serviceReq := &dto.CustomerListRequest{
		IDs: req.IDs,
	}

	customers, err := cc.customerService.ListCustomers(c.Request.Context(), serviceReq)
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
	customer, err := cc.customerService.GetCustomerByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
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
	var req dto.CustomerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	if err := cc.customerService.UpdateCustomer(c.Request.Context(), id, &req); err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			resp.Error(c, resp.CodeNotFound, "customer not found")
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
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
	if err := cc.customerService.DeleteCustomer(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			resp.Error(c, resp.CodeNotFound, "customer not found")
			return
		}
		resp.Error(c, resp.CodeInternalError, "failed to delete customer")
		return
	}
	resp.Success(c, nil)
}
