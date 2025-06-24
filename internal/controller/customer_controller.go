package controller

import (
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomerController struct {
	customerService *service.CustomerService
}

func NewCustomerController(db *gorm.DB) *CustomerController {
	return &CustomerController{customerService: service.NewCustomerService(db)}
}

func (cc *CustomerController) CreateCustomer(c *gin.Context) {
	var req dto.CustomerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	customer, err := cc.customerService.CreateCustomer(c.Request.Context(), &req)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to create customer")
		return
	}
	resp.Success(c, customer)
}

func (cc *CustomerController) ListCustomers(c *gin.Context) {
	customers, err := cc.customerService.ListCustomers(c.Request.Context())
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to list customers")
		return
	}
	resp.Success(c, customers)
}

func (cc *CustomerController) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	customer, err := cc.customerService.GetCustomerByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Error(c, resp.CodeNotFound, "customer not found")
		} else {
			resp.Error(c, resp.CodeInternalError, "failed to get customer")
		}
		return
	}
	resp.Success(c, customer)
}

func (cc *CustomerController) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var req dto.CustomerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}
	if err := cc.customerService.UpdateCustomer(c.Request.Context(), id, &req); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to update customer")
		return
	}
	resp.Success(c, nil)
}

func (cc *CustomerController) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	if err := cc.customerService.DeleteCustomer(c.Request.Context(), id); err != nil {
		resp.Error(c, resp.CodeInternalError, "failed to delete customer")
		return
	}
	resp.Success(c, nil)
}
