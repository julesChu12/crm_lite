package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContactController struct {
	svc *service.ContactService
}

func NewContactController(resManager *resource.Manager) *ContactController {
	return &ContactController{svc: service.NewContactService(resManager)}
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

	list, err := cc.svc.List(c.Request.Context(), customerID)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
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

	contact, err := cc.svc.GetContactByID(c.Request.Context(), contactID)
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			resp.Error(c, resp.CodeNotFound, "contact not found")
			return
		}
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

	contact, err := cc.svc.Create(c.Request.Context(), customerID, &req)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			resp.Error(c, resp.CodeNotFound, "customer not found")
			return
		}
		if errors.Is(err, service.ErrPrimaryContactAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "primary contact already exists for this customer")
			return
		}
		if errors.Is(err, service.ErrContactPhoneAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "contact phone number already exists")
			return
		}
		if errors.Is(err, service.ErrContactEmailAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "contact email already exists")
			return
		}
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

	err = cc.svc.Update(c.Request.Context(), contactID, &req)
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			resp.Error(c, resp.CodeNotFound, "contact not found")
			return
		}
		if errors.Is(err, service.ErrPrimaryContactAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "primary contact already exists for this customer")
			return
		}
		if errors.Is(err, service.ErrContactPhoneAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "contact phone number already exists")
			return
		}
		if errors.Is(err, service.ErrContactEmailAlreadyExists) {
			resp.Error(c, resp.CodeConflict, "contact email already exists")
			return
		}
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

	err = cc.svc.Delete(c.Request.Context(), contactID)
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			resp.Error(c, resp.CodeNotFound, "contact not found")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeNoContent, nil)
}
