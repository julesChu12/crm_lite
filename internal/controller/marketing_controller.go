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

// MarketingController 负责处理营销相关的HTTP请求
type MarketingController struct {
	marketingService *service.MarketingService
}

// NewMarketingController 创建一个新的MarketingController实例
func NewMarketingController(resManager *resource.Manager) *MarketingController {
	return &MarketingController{
		marketingService: service.NewMarketingService(resManager),
	}
}

// ================ 营销活动管理 ================

// CreateCampaign godoc
// @Summary      创建营销活动
// @Description  创建一个新的营销活动
// @Tags         Marketing
// @Accept       json
// @Produce      json
// @Param        campaign body dto.MarketingCampaignCreateRequest true "营销活动信息"
// @Success      201 {object} resp.Response{data=dto.MarketingCampaignResponse} "创建成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      409 {object} resp.Response "活动名称已存在"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/campaigns [post]
func (mc *MarketingController) CreateCampaign(c *gin.Context) {
	var req dto.MarketingCampaignCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		resp.Error(c, resp.CodeUnauthorized, "user not authenticated")
		return
	}

	createdBy, ok := userID.(string)
	if !ok {
		resp.Error(c, resp.CodeInternalError, "invalid user ID format")
		return
	}

	createdByInt, err := strconv.ParseInt(createdBy, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "invalid user ID")
		return
	}

	campaign, err := mc.marketingService.CreateCampaign(c.Request.Context(), &req, createdByInt)
	if err != nil {
		if errors.Is(err, service.ErrMarketingCampaignNameExists) {
			resp.Error(c, resp.CodeConflict, "campaign name already exists")
			return
		}
		if errors.Is(err, service.ErrMarketingInvalidTimeRange) {
			resp.Error(c, resp.CodeInvalidParam, "invalid time range: start time must be before end time")
			return
		}
		resp.SystemError(c, err)
		return
	}

	resp.SuccessWithCode(c, resp.CodeCreated, campaign)
}

// GetCampaign godoc
// @Summary      获取单个营销活动
// @Description  根据ID获取营销活动详细信息
// @Tags         Marketing
// @Produce      json
// @Param        id path string true "营销活动ID"
// @Success      200 {object} resp.Response{data=dto.MarketingCampaignResponse} "获取成功"
// @Failure      404 {object} resp.Response "营销活动未找到"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/campaigns/{id} [get]
func (mc *MarketingController) GetCampaign(c *gin.Context) {
	id := c.Param("id")
	campaign, err := mc.marketingService.GetCampaign(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrMarketingCampaignNotFound) {
			resp.Error(c, resp.CodeNotFound, "marketing campaign not found")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, campaign)
}

// ListCampaigns godoc
// @Summary      获取营销活动列表
// @Description  获取营销活动列表，支持分页和筛选
// @Tags         Marketing
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页大小" default(10)
// @Param        name query string false "活动名称模糊搜索"
// @Param        type query string false "活动类型" Enums(sms,email,push_notification,wechat,call)
// @Param        status query string false "活动状态" Enums(draft,scheduled,active,paused,completed,archived)
// @Param        order_by query string false "排序字段" example("created_at_desc")
// @Success      200 {object} resp.Response{data=dto.MarketingCampaignListResponse} "获取成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/campaigns [get]
func (mc *MarketingController) ListCampaigns(c *gin.Context) {
	var req dto.MarketingCampaignListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	result, err := mc.marketingService.ListCampaigns(c.Request.Context(), &req)
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, result)
}

// UpdateCampaign godoc
// @Summary      更新营销活动
// @Description  更新营销活动信息
// @Tags         Marketing
// @Accept       json
// @Produce      json
// @Param        id path string true "营销活动ID"
// @Param        campaign body dto.MarketingCampaignUpdateRequest true "要更新的营销活动信息"
// @Success      200 {object} resp.Response{data=dto.MarketingCampaignResponse} "更新成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      404 {object} resp.Response "营销活动未找到"
// @Failure      409 {object} resp.Response "业务冲突（如活动名称已存在或无法修改）"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/campaigns/{id} [put]
func (mc *MarketingController) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	var req dto.MarketingCampaignUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		resp.Error(c, resp.CodeUnauthorized, "user not authenticated")
		return
	}

	updatedBy, ok := userID.(string)
	if !ok {
		resp.Error(c, resp.CodeInternalError, "invalid user ID format")
		return
	}

	updatedByInt, err := strconv.ParseInt(updatedBy, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInternalError, "invalid user ID")
		return
	}

	campaign, err := mc.marketingService.UpdateCampaign(c.Request.Context(), id, &req, updatedByInt)
	if err != nil {
		if errors.Is(err, service.ErrMarketingCampaignNotFound) {
			resp.Error(c, resp.CodeNotFound, "marketing campaign not found")
			return
		}
		if errors.Is(err, service.ErrMarketingCampaignNameExists) {
			resp.Error(c, resp.CodeConflict, "campaign name already exists")
			return
		}
		if errors.Is(err, service.ErrMarketingCampaignCannotModify) {
			resp.Error(c, resp.CodeConflict, "cannot modify campaign in current status")
			return
		}
		if errors.Is(err, service.ErrMarketingInvalidTimeRange) {
			resp.Error(c, resp.CodeInvalidParam, "invalid time range: start time must be before end time")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, campaign)
}

// DeleteCampaign godoc
// @Summary      删除营销活动
// @Description  删除指定的营销活动
// @Tags         Marketing
// @Produce      json
// @Param        id path string true "营销活动ID"
// @Success      204 {object} resp.Response "删除成功"
// @Failure      404 {object} resp.Response "营销活动未找到"
// @Failure      409 {object} resp.Response "无法删除运行中的活动"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/campaigns/{id} [delete]
func (mc *MarketingController) DeleteCampaign(c *gin.Context) {
	id := c.Param("id")
	err := mc.marketingService.DeleteCampaign(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrMarketingCampaignNotFound) {
			resp.Error(c, resp.CodeNotFound, "marketing campaign not found")
			return
		}
		if errors.Is(err, service.ErrMarketingCampaignCannotModify) {
			resp.Error(c, resp.CodeConflict, "cannot delete active campaign")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.SuccessWithCode(c, resp.CodeNoContent, nil)
}

// ExecuteCampaign godoc
// @Summary      执行营销活动
// @Description  手动触发营销活动执行或进行模拟执行
// @Tags         Marketing
// @Accept       json
// @Produce      json
// @Param        id path string true "营销活动ID"
// @Param        request body dto.MarketingCampaignExecuteRequest false "执行参数"
// @Success      200 {object} resp.Response{data=dto.MarketingCampaignExecuteResponse} "执行成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      404 {object} resp.Response "营销活动未找到"
// @Failure      409 {object} resp.Response "活动状态不允许执行"
// @Failure      422 {object} resp.Response "没有找到目标客户"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/campaigns/{id}/execute [post]
func (mc *MarketingController) ExecuteCampaign(c *gin.Context) {
	id := c.Param("id")
	var req dto.MarketingCampaignExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	result, err := mc.marketingService.ExecuteCampaign(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrMarketingCampaignNotFound) {
			resp.Error(c, resp.CodeNotFound, "marketing campaign not found")
			return
		}
		if errors.Is(err, service.ErrMarketingCampaignInvalidStatus) {
			resp.Error(c, resp.CodeConflict, "campaign status does not allow execution")
			return
		}
		if errors.Is(err, service.ErrMarketingNoTargetCustomers) {
			resp.Error(c, resp.CodeInvalidParam, "no target customers found for this campaign")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, result)
}

// ================ 营销记录管理 ================

// ListMarketingRecords godoc
// @Summary      获取营销记录列表
// @Description  获取指定营销活动的客户触达记录
// @Tags         Marketing
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页大小" default(50)
// @Param        campaign_id query int true "营销活动ID"
// @Param        customer_id query int false "客户ID"
// @Param        channel query string false "触达渠道" Enums(sms,email,push_notification,wechat,call)
// @Param        status query string false "记录状态" Enums(pending,sent,delivered,failed,opened,clicked,replied,unsubscribed)
// @Param        start_date query string false "开始日期" format(date)
// @Param        end_date query string false "结束日期" format(date)
// @Success      200 {object} resp.Response{data=dto.MarketingRecordListResponse} "获取成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/records [get]
func (mc *MarketingController) ListMarketingRecords(c *gin.Context) {
	var req dto.MarketingRecordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	result, err := mc.marketingService.ListMarketingRecords(c.Request.Context(), &req)
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, result)
}

// GetCampaignStats godoc
// @Summary      获取营销活动统计
// @Description  获取指定营销活动的详细统计数据
// @Tags         Marketing
// @Produce      json
// @Param        id path string true "营销活动ID"
// @Success      200 {object} resp.Response{data=dto.MarketingCampaignStatsResponse} "获取成功"
// @Failure      404 {object} resp.Response "营销活动未找到"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/campaigns/{id}/stats [get]
func (mc *MarketingController) GetCampaignStats(c *gin.Context) {
	id := c.Param("id")
	stats, err := mc.marketingService.GetCampaignStats(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrMarketingCampaignNotFound) {
			resp.Error(c, resp.CodeNotFound, "marketing campaign not found")
			return
		}
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, stats)
}

// ================ 客户分群管理 ================

// GetCustomerSegment godoc
// @Summary      获取客户分群
// @Description  根据指定条件获取客户分群信息
// @Tags         Marketing
// @Accept       json
// @Produce      json
// @Param        segment body dto.CustomerSegmentRequest true "分群条件"
// @Success      200 {object} resp.Response{data=dto.CustomerSegmentResponse} "获取成功"
// @Failure      400 {object} resp.Response "请求参数错误"
// @Failure      500 {object} resp.Response "服务器内部错误"
// @Security     ApiKeyAuth
// @Router       /marketing/customer-segments [post]
func (mc *MarketingController) GetCustomerSegment(c *gin.Context) {
	var req dto.CustomerSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, resp.CodeInvalidParam, err.Error())
		return
	}

	result, err := mc.marketingService.GetCustomerSegment(c.Request.Context(), &req)
	if err != nil {
		resp.SystemError(c, err)
		return
	}
	resp.Success(c, result)
}
