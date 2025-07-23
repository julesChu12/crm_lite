package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/dto"
	"crm_lite/internal/service"
	"crm_lite/pkg/resp"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContactController struct {
	svc *service.ContactService
}

func NewContactController(resManager *resource.Manager) *ContactController {
	return &ContactController{svc: service.NewContactService(resManager)}
}

// List @Summary 获取客户的联系人列表
// @Description 根据客户ID获取其所有联系人
// @Tags Contacts
// @Produce json
// @Param id path int true "客户ID"
// @Success 200 {object} resp.Response{data=[]dto.ContactResponse} "成功"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/contacts [get]
func (cc *ContactController) List(ctx *gin.Context) {
	customerID, _ := strconv.ParseInt(ctx.Param("id"), 10, 64)
	list, err := cc.svc.List(ctx.Request.Context(), customerID)
	if err != nil {
		resp.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(ctx, list)
}

// Create @Summary 为客户创建新联系人
// @Description 为指定客户创建一个新的联系人
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param contact body dto.ContactCreateRequest true "联系人信息"
// @Success 200 {object} resp.Response{data=dto.ContactResponse} "成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/contacts [post]
func (cc *ContactController) Create(ctx *gin.Context) {
	customerID, _ := strconv.ParseInt(ctx.Param("id"), 10, 64)
	var req dto.ContactCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, 400, err.Error())
		return
	}
	res, err := cc.svc.Create(ctx.Request.Context(), customerID, &req)
	if err != nil {
		resp.Error(ctx, 500, err.Error())
		return
	}
	resp.Success(ctx, res)
}

// Update @Summary 更新联系人信息
// @Description 更新指定ID的联系人信息
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param contact_id path int true "联系人ID"
// @Param contact body dto.ContactUpdateRequest true "要更新的联系人信息"
// @Success 200 {object} resp.Response "操作成功"
// @Failure 400 {object} resp.Response "请求参数错误"
// @Failure 404 {object} resp.Response "联系人未找到"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/contacts/{contact_id} [put]
func (cc *ContactController) Update(ctx *gin.Context) {
	id, _ := strconv.ParseInt(ctx.Param("contact_id"), 10, 64)
	var req dto.ContactUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.Error(ctx, 400, err.Error())
		return
	}
	err := cc.svc.Update(ctx.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Error(ctx, 404, "not found")
		} else {
			resp.Error(ctx, 500, err.Error())
		}
		return
	}
	resp.Success(ctx, nil)
}

// Delete @Summary 删除联系人
// @Description 删除指定ID的联系人
// @Tags Contacts
// @Produce json
// @Param id path int true "客户ID"
// @Param contact_id path int true "联系人ID"
// @Success 200 {object} resp.Response "操作成功"
// @Failure 404 {object} resp.Response "联系人未找到"
// @Failure 500 {object} resp.Response "服务器内部错误"
// @Router /v1/customers/{id}/contacts/{contact_id} [delete]
func (cc *ContactController) Delete(ctx *gin.Context) {
	id, _ := strconv.ParseInt(ctx.Param("contact_id"), 10, 64)
	err := cc.svc.Delete(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Error(ctx, 404, "not found")
		} else {
			resp.Error(ctx, 500, err.Error())
		}
		return
	}
	resp.Success(ctx, nil)
}
