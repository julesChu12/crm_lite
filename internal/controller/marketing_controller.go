package controller

import (
	"crm_lite/internal/core/resource"
	"crm_lite/internal/domains/marketing"
	"crm_lite/internal/domains/marketing/impl"
	"crm_lite/internal/dto"
	"crm_lite/pkg/resp"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MarketingController 负责处理营销相关的HTTP请求
// 已完全迁移到 marketing 域服务
type MarketingController struct {
	marketingSvc marketing.Service
}

// NewMarketingController 创建一个新的MarketingController实例
func NewMarketingController(resManager *resource.Manager) *MarketingController {
	// 创建Marketing领域服务
	dbRes, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for MarketingController: " + err.Error())
	}
	marketingSvc := impl.NewMarketingService(dbRes.DB)

	return &MarketingController{
		marketingSvc: marketingSvc,
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

	// 转换为Marketing领域请求
	marketingReq := marketing.CreateCampaignRequest{
		Name:        req.Name,
		Description: req.Content, // 使用Content作为Description
		Type:        req.Type,
		Channel:     req.Type, // 使用Type作为Channel
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Budget:      0, // 默认预算为0
		TargetCount: 0, // 默认目标数为0
	}

	campaign, err := mc.marketingSvc.CreateCampaign(c.Request.Context(), marketingReq, createdByInt)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	campaignResponse := &dto.MarketingCampaignResponse{
		ID:           campaign.ID,
		Name:         campaign.Name,
		Type:         campaign.Type,
		Status:       campaign.Status,
		Content:      campaign.Description,
		StartTime:    campaign.StartTime,
		EndTime:      campaign.EndTime,
		TargetCount:  int32(campaign.TargetCount),
		SentCount:    int32(campaign.ActualCount),
		SuccessCount: int32(campaign.ActualCount),
		ClickCount:   0,
		CreatedBy:    campaign.CreatedBy,
		CreatedAt:    time.Unix(campaign.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		UpdatedAt:    time.Unix(campaign.UpdatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.SuccessWithCode(c, resp.CodeCreated, campaignResponse)
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
	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid campaign ID")
		return
	}

	campaign, err := mc.marketingSvc.GetCampaign(c.Request.Context(), campaignID)
	if err != nil {
		resp.Error(c, resp.CodeNotFound, "marketing campaign not found")
		return
	}

	// 转换为DTO格式
	campaignResponse := &dto.MarketingCampaignResponse{
		ID:           campaign.ID,
		Name:         campaign.Name,
		Type:         campaign.Type,
		Status:       campaign.Status,
		Content:      campaign.Description,
		StartTime:    campaign.StartTime,
		EndTime:      campaign.EndTime,
		TargetCount:  int32(campaign.TargetCount),
		SentCount:    int32(campaign.ActualCount),
		SuccessCount: int32(campaign.ActualCount),
		ClickCount:   0,
		CreatedBy:    campaign.CreatedBy,
		CreatedAt:    time.Unix(campaign.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		UpdatedAt:    time.Unix(campaign.UpdatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.Success(c, campaignResponse)
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

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 调用Marketing领域服务
	campaigns, total, err := mc.marketingSvc.ListCampaigns(c.Request.Context(), req.Status, req.Page, req.PageSize)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// 转换为DTO格式
	campaignResponses := make([]*dto.MarketingCampaignResponse, len(campaigns))
	for i, campaign := range campaigns {
		campaignResponses[i] = &dto.MarketingCampaignResponse{
			ID:           campaign.ID,
			Name:         campaign.Name,
			Type:         campaign.Type,
			Status:       campaign.Status,
			Content:      campaign.Description,
			StartTime:    campaign.StartTime,
			EndTime:      campaign.EndTime,
			TargetCount:  int32(campaign.TargetCount),
			SentCount:    int32(campaign.ActualCount),
			SuccessCount: int32(campaign.ActualCount),
			ClickCount:   0,
			CreatedBy:    campaign.CreatedBy,
			CreatedAt:    time.Unix(campaign.CreatedAt, 0).Format("2006-01-02 15:04:05"),
			UpdatedAt:    time.Unix(campaign.UpdatedAt, 0).Format("2006-01-02 15:04:05"),
		}
	}

	result := &dto.MarketingCampaignListResponse{
		Campaigns: campaignResponses,
		Total:     total,
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

	// Convert to Marketing domain request
	marketingUpdateReq := marketing.UpdateCampaignRequest{
		Name:        &req.Name,
		Description: &req.Content,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid campaign ID")
		return
	}

	err = mc.marketingSvc.UpdateCampaign(c.Request.Context(), campaignID, marketingUpdateReq, updatedByInt)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// Get updated campaign
	campaign, err := mc.marketingSvc.GetCampaign(c.Request.Context(), campaignID)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// Convert to DTO format
	campaignResponse := &dto.MarketingCampaignResponse{
		ID:           campaign.ID,
		Name:         campaign.Name,
		Type:         campaign.Type,
		Status:       campaign.Status,
		Content:      campaign.Description,
		StartTime:    campaign.StartTime,
		EndTime:      campaign.EndTime,
		TargetCount:  int32(campaign.TargetCount),
		SentCount:    int32(campaign.ActualCount),
		SuccessCount: int32(campaign.ActualCount),
		ClickCount:   0,
		CreatedBy:    campaign.CreatedBy,
		CreatedAt:    time.Unix(campaign.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		UpdatedAt:    time.Unix(campaign.UpdatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	resp.Success(c, campaignResponse)
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
	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid campaign ID")
		return
	}

	err = mc.marketingSvc.DeleteCampaign(c.Request.Context(), campaignID)
	if err != nil {
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

	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid campaign ID")
		return
	}

	// For now, start the campaign (simplified implementation)
	err = mc.marketingSvc.StartCampaign(c.Request.Context(), campaignID)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// Return simple response
	result := &dto.MarketingCampaignExecuteResponse{
		Status:      "triggered",
		Message:     "Campaign execution started",
		ExecutionID: fmt.Sprintf("exec-%d", campaignID),
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

	// Set default values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	records, total, err := mc.marketingSvc.ListRecords(c.Request.Context(), req.CampaignID, req.CustomerID, req.Page, req.PageSize)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// Convert to DTO format
	recordResponses := make([]*dto.MarketingRecordResponse, len(records))
	for i, record := range records {
		var sentAt, deliveredAt *time.Time
		if record.SentAt > 0 {
			t := time.Unix(record.SentAt, 0)
			sentAt = &t
		}
		if record.DeliveredAt > 0 {
			t := time.Unix(record.DeliveredAt, 0)
			deliveredAt = &t
		}

		recordResponses[i] = &dto.MarketingRecordResponse{
			ID:          record.ID,
			CampaignID:  record.CampaignID,
			CustomerID:  record.CustomerID,
			Channel:     record.Channel,
			Status:      record.Status,
			SentAt:      sentAt,
			DeliveredAt: deliveredAt,
		}
	}

	result := &dto.MarketingRecordListResponse{
		Records: recordResponses,
		Total:   total,
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
	campaignID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		resp.Error(c, resp.CodeInvalidParam, "invalid campaign ID")
		return
	}

	stats, err := mc.marketingSvc.GetCampaignStats(c.Request.Context(), campaignID)
	if err != nil {
		resp.SystemError(c, err)
		return
	}

	// Convert to DTO format
	statsResponse := &dto.MarketingCampaignStatsResponse{
		CampaignID:        stats.CampaignID,
		TargetCount:       int32(stats.TotalRecords),
		SentCount:         int32(stats.SentCount),
		DeliveredCount:    int32(stats.DeliveredCount),
		FailedCount:       int32(stats.TotalRecords - stats.SentCount),
		OpenedCount:       int32(stats.OpenedCount),
		ClickedCount:      int32(stats.ClickedCount),
		RepliedCount:      0,
		UnsubscribedCount: 0,
		DeliveryRate:      stats.DeliveryRate,
		OpenRate:          stats.OpenRate,
		ClickRate:         stats.ClickRate,
	}
	resp.Success(c, statsResponse)
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

	// For now, return empty customer segment (simplified implementation)
	// This would need to be implemented in the marketing service
	result := &dto.CustomerSegmentResponse{
		Total: 0,
		Customers: []struct {
			ID    int64  `json:"id" example:"100"`
			Name  string `json:"name" example:"张三"`
			Phone string `json:"phone" example:"138****8888"`
			Email string `json:"email,omitempty" example:"zhangsan@example.com"`
		}{},
	}
	resp.Success(c, result)
}
