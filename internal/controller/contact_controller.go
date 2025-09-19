package controller

import (
	"crm_lite/internal/common"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/billing/impl"
	"crm_lite/internal/domains/crm"
	crmimpl "crm_lite/internal/domains/crm/impl"
	"crm_lite/internal/dto"
	"crm_lite/pkg/resp"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContactController struct {
	crmSvc     crm.Service
	resManager *resource.Manager
}

func NewContactController(resManager *resource.Manager) *ContactController {
	// 创建CRM域服务
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic(err)
	}

	// 创建billing服务作为依赖
	billingSvc := impl.NewBillingService(dbRes.DB)
	crmSvc := crmimpl.NewCRMServiceWithBilling(dbRes.DB, billingSvc)

	return &ContactController{crmSvc: crmSvc, resManager: resManager}
}

// ListContacts godoc
// @Summary      获取客户的联系人列表
// @Description  根据客户ID获取其所有联系人
// @Tags         Contacts
// @Produce      json
// @Param        id path int true "客户ID"
// @Success      200 {object} resp.Response{data=[]dto.ContactResponse} "成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      404 {object} resp.Response "客户未找到"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Router       /customers/{id}/contacts [get]
func (cc *ContactController) ListContacts(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid customer ID")
		return
	}

	list, err := cc.crmSvc.ListContactsLegacy(c.Request.Context(), customerID)
	if err != nil {
		var businessErr2 *common.BusinessError
		if errors.As(err, &businessErr2) && businessErr2.Code == common.ErrCodeResourceNotFound {
			resp.Error(c, resp.CodeNotFound, "customer not found")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, list)
}

// GetContact godoc
// @Summary      获取单个联系人详情
// @Description  根据联系人ID获取其详细信息
// @Tags         Contacts
// @Produce      json
// @Param        id path int true "客户ID"
// @Param        contact_id path int true "联系人ID"
// @Success      200 {object} resp.Response{data=dto.ContactResponse} "成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      404 {object} resp.Response "联系人未找到"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Router       /customers/{id}/contacts/{contact_id} [get]
func (cc *ContactController) GetContact(c *gin.Context) {
	contactIDStr := c.Param("contact_id")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid contact ID")
		return
	}

	contact, err := cc.crmSvc.GetContactByIDLegacy(c.Request.Context(), contactID)
	if err != nil {
		// Generic error handling since specific errors were removed
		resp.SystemError(c, err)
		return
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, contact)
}

// CreateContact godoc
// @Summary      为客户创建新联系人
// @Description  为指定客户创建一个新的联系人
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Param        id path int true "客户ID"
// @Param        contact body dto.ContactCreateRequest true "联系人信息"
// @Success      201 {object} resp.Response{data=dto.ContactResponse} "成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      404 {object} resp.Response "客户未找到"
// @Failure      409 {object} resp.Response "业务冲突（如主要联系人已存在）"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Router       /customers/{id}/contacts [post]
func (cc *ContactController) CreateContact(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid customer ID")
		return
	}

	var req dto.ContactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// Convert DTO to CRM domain type
	crmReq := &crm.ContactCreateRequest{
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		Position:  req.Position,
		IsPrimary: req.IsPrimary,
		Note:      req.Note,
	}

	contact, err := cc.crmSvc.CreateContactLegacy(c.Request.Context(), customerID, crmReq)
	if err != nil {
		var businessErr *common.BusinessError
		if errors.As(err, &businessErr) {
			switch businessErr.Code {
			case common.ErrCodeResourceNotFound:
				resp.Error(c, resp.CodeNotFound, "customer not found")
				return
			case common.ErrCodeDuplicateResource:
				resp.Error(c, resp.CodeConflict, "contact already exists")
				return
			}
		}
		// Generic error handling since specific errors were removed
		resp.SystemError(c, err)
		return
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeCreated, contact)
}

// UpdateContact godoc
// @Summary      更新联系人信息
// @Description  更新指定ID的联系人信息
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Param        id path int true "客户ID"
// @Param        contact_id path int true "联系人ID"
// @Param        contact body dto.ContactUpdateRequest true "要更新的联系人信息"
// @Success      200 {object} resp.Response "操作成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      404 {object} resp.Response "联系人未找到"
// @Failure      409 {object} resp.Response "业务冲突（如主要联系人已存在）"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Router       /customers/{id}/contacts/{contact_id} [put]
func (cc *ContactController) UpdateContact(c *gin.Context) {
	contactIDStr := c.Param("contact_id")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid contact ID")
		return
	}

	var req dto.ContactUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// Convert DTO to CRM domain type
	crmReq := &crm.ContactUpdateRequest{
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		Position:  req.Position,
		IsPrimary: req.IsPrimary,
		Note:      req.Note,
	}

	err = cc.crmSvc.UpdateContactLegacy(c.Request.Context(), contactID, crmReq)
	if err != nil {
		// Generic error handling since specific errors were removed
		resp.SystemError(c, err)
		return
		var businessErr *common.BusinessError
		if errors.As(err, &businessErr) {
			switch businessErr.Code {
			case common.ErrCodeResourceNotFound:
				resp.Error(c, resp.CodeNotFound, "customer not found")
				return
			case common.ErrCodeDuplicateResource:
				resp.Error(c, resp.CodeConflict, "contact already exists")
				return
			}
		}
		// Generic error handling since specific errors were removed
		resp.SystemError(c, err)
		return
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, nil)
}

// DeleteContact godoc
// @Summary      删除联系人
// @Description  删除指定ID的联系人
// @Tags         Contacts
// @Produce      json
// @Param        id path int true "客户ID"
// @Param        contact_id path int true "联系人ID"
// @Success      204 {object} resp.Response "操作成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      404 {object} resp.Response "联系人未找到"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Router       /customers/{id}/contacts/{contact_id} [delete]
func (cc *ContactController) DeleteContact(c *gin.Context) {
	contactIDStr := c.Param("contact_id")
	contactID, err := strconv.ParseInt(contactIDStr, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid contact ID")
		return
	}

	err = cc.crmSvc.DeleteContactLegacy(c.Request.Context(), contactID)
	if err != nil {
		// Generic error handling since specific errors were removed
		resp.SystemError(c, err)
		return
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeNoContent, nil)
}
