package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crm_lite/internal/core/resource"
	"crm_lite/internal/dao/model"
	"crm_lite/internal/dao/query"
	"crm_lite/internal/dto"
	"crm_lite/pkg/utils"

	"errors"

	"gorm.io/gorm"
)

type MarketingService struct {
	q        *query.Query
	resource *resource.Manager
}

func NewMarketingService(resManager *resource.Manager) *MarketingService {
	db, err := resource.Get[*resource.DBResource](resManager, resource.DBServiceKey)
	if err != nil {
		panic("Failed to get database resource for MarketingService: " + err.Error())
	}
	return &MarketingService{
		q:        query.Use(db.DB),
		resource: resManager,
	}
}

// ================ 营销活动管理 ================

// CreateCampaign 创建营销活动
func (s *MarketingService) CreateCampaign(ctx context.Context, req *dto.MarketingCampaignCreateRequest, createdBy int64) (*dto.MarketingCampaignResponse, error) {
	// 1. 验证时间范围
	if req.StartTime.After(req.EndTime) {
		return nil, ErrMarketingInvalidTimeRange
	}

	// 2. 检查活动名称唯一性
	count, err := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.Name.Eq(req.Name)).
		Count()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrMarketingCampaignNameExists
	}

	// 3. 序列化目标标签
	targetTagsJSON, err := json.Marshal(req.TargetTags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal target tags: %w", err)
	}

	// 4. 确定初始状态
	status := "draft"
	if req.StartTime.After(time.Now()) {
		status = "scheduled"
	}

	// 5. 创建营销活动
	campaign := &model.MarketingCampaign{
		Name:              req.Name,
		Type:              req.Type,
		Status:            status,
		TargetTags:        string(targetTagsJSON),
		TargetSegmentID:   req.TargetSegmentID,
		ContentTemplateID: req.ContentTemplateID,
		Content:           req.Content,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		CreatedBy:         createdBy,
		UpdatedBy:         createdBy,
	}

	if err := s.q.MarketingCampaign.WithContext(ctx).Create(campaign); err != nil {
		return nil, err
	}

	return s.toCampaignResponse(campaign), nil
}

// GetCampaign 获取单个营销活动
func (s *MarketingService) GetCampaign(ctx context.Context, id string) (*dto.MarketingCampaignResponse, error) {
	idNum, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, ErrMarketingCampaignNotFound
	}

	campaign, err := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(idNum)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMarketingCampaignNotFound
		}
		return nil, err
	}

	return s.toCampaignResponse(campaign), nil
}

// ListCampaigns 获取营销活动列表
func (s *MarketingService) ListCampaigns(ctx context.Context, req *dto.MarketingCampaignListRequest) (*dto.MarketingCampaignListResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	q := s.q.MarketingCampaign.WithContext(ctx)

	// 构建筛选条件
	if req.Name != "" {
		q = q.Where(s.q.MarketingCampaign.Name.Like("%" + req.Name + "%"))
	}
	if req.Type != "" {
		q = q.Where(s.q.MarketingCampaign.Type.Eq(req.Type))
	}
	if req.Status != "" {
		q = q.Where(s.q.MarketingCampaign.Status.Eq(req.Status))
	}

	// 构建排序条件
	if req.OrderBy != "" {
		parts := strings.Split(req.OrderBy, "_")
		if len(parts) == 2 {
			field := parts[0]
			order := parts[1]
			if col, ok := s.q.MarketingCampaign.GetFieldByName(field); ok {
				if order == "desc" {
					q = q.Order(col.Desc())
				} else {
					q = q.Order(col)
				}
			}
		}
	} else {
		// 默认按创建时间降序
		q = q.Order(s.q.MarketingCampaign.CreatedAt.Desc())
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// 分页查询
	campaigns, err := q.Limit(req.PageSize).
		Offset((req.Page - 1) * req.PageSize).
		Find()
	if err != nil {
		return nil, err
	}

	// 转换响应
	responses := make([]*dto.MarketingCampaignResponse, 0, len(campaigns))
	for _, campaign := range campaigns {
		responses = append(responses, s.toCampaignResponse(campaign))
	}

	return &dto.MarketingCampaignListResponse{
		Total:     total,
		Campaigns: responses,
	}, nil
}

// UpdateCampaign 更新营销活动
func (s *MarketingService) UpdateCampaign(ctx context.Context, id string, req *dto.MarketingCampaignUpdateRequest, updatedBy int64) (*dto.MarketingCampaignResponse, error) {
	idNum, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, ErrMarketingCampaignNotFound
	}

	// 1. 检查活动是否存在
	existingCampaign, err := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(idNum)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMarketingCampaignNotFound
		}
		return nil, err
	}

	// 2. 检查是否可以修改
	if existingCampaign.Status == "active" && req.Status != "paused" && req.Status != "completed" {
		return nil, ErrMarketingCampaignCannotModify
	}

	// 3. 验证时间范围
	if req.StartTime != nil && req.EndTime != nil && req.StartTime.After(*req.EndTime) {
		return nil, ErrMarketingInvalidTimeRange
	}

	// 4. 构建更新数据
	updates := make(map[string]interface{})
	if req.Name != "" {
		// 检查名称唯一性（排除当前活动）
		count, err := s.q.MarketingCampaign.WithContext(ctx).
			Where(s.q.MarketingCampaign.Name.Eq(req.Name), s.q.MarketingCampaign.ID.Neq(idNum)).
			Count()
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrMarketingCampaignNameExists
		}
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.TargetTags != nil {
		targetTagsJSON, err := json.Marshal(req.TargetTags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal target tags: %w", err)
		}
		updates["target_tags"] = string(targetTagsJSON)
	}
	if req.TargetSegmentID != 0 {
		updates["target_segment_id"] = req.TargetSegmentID
	}
	if req.ContentTemplateID != 0 {
		updates["content_template_id"] = req.ContentTemplateID
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	updates["updated_by"] = updatedBy

	// 5. 执行更新
	if len(updates) > 0 {
		result, err := s.q.MarketingCampaign.WithContext(ctx).
			Where(s.q.MarketingCampaign.ID.Eq(idNum)).
			Updates(updates)
		if err != nil {
			return nil, err
		}
		if result.RowsAffected == 0 {
			return nil, ErrMarketingCampaignNotFound
		}
	}

	// 6. 重新获取更新后的数据
	return s.GetCampaign(ctx, id)
}

// DeleteCampaign 删除营销活动
func (s *MarketingService) DeleteCampaign(ctx context.Context, id string) error {
	idNum, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return ErrMarketingCampaignNotFound
	}

	// 检查活动状态，运行中的活动不能删除
	campaign, err := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(idNum)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMarketingCampaignNotFound
		}
		return err
	}

	if campaign.Status == "active" {
		return ErrMarketingCampaignCannotModify
	}

	result, err := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(idNum)).
		Delete()
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrMarketingCampaignNotFound
	}

	return nil
}

// ExecuteCampaign 执行营销活动
func (s *MarketingService) ExecuteCampaign(ctx context.Context, id string, req *dto.MarketingCampaignExecuteRequest) (*dto.MarketingCampaignExecuteResponse, error) {
	idNum, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, ErrMarketingCampaignNotFound
	}

	// 1. 获取活动信息
	campaign, err := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(idNum)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMarketingCampaignNotFound
		}
		return nil, err
	}

	// 2. 检查活动状态
	if campaign.Status != "scheduled" && campaign.Status != "draft" {
		return nil, ErrMarketingCampaignInvalidStatus
	}

	// 3. 获取目标客户
	customers, err := s.getTargetCustomers(ctx, campaign)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, ErrMarketingNoTargetCustomers
	}

	// 4. 根据执行类型处理
	executionType := req.ExecutionType
	if executionType == "" {
		executionType = "actual"
	}

	if executionType == "simulation" {
		// 模拟执行，只返回统计信息
		return &dto.MarketingCampaignExecuteResponse{
			Status:      "simulation_started",
			Message:     fmt.Sprintf("模拟执行完成，预计触达 %d 位客户", len(customers)),
			ExecutionID: fmt.Sprintf("sim-%d-%d", campaign.ID, time.Now().Unix()),
		}, nil
	}

	// 5. 实际执行：更新活动状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":            "active",
		"actual_start_time": now,
		"target_count":      int32(len(customers)),
	}

	_, err = s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(idNum)).
		Updates(updates)
	if err != nil {
		return nil, err
	}

	// 6. 创建营销记录（这里简化处理，实际应该异步执行）
	go s.createMarketingRecords(context.Background(), campaign, customers)

	return &dto.MarketingCampaignExecuteResponse{
		Status:      "triggered",
		Message:     fmt.Sprintf("营销活动已成功触发执行，目标客户 %d 位", len(customers)),
		ExecutionID: fmt.Sprintf("exec-%d-%d", campaign.ID, time.Now().Unix()),
	}, nil
}

// ================ 营销记录管理 ================

// ListMarketingRecords 获取营销记录列表
func (s *MarketingService) ListMarketingRecords(ctx context.Context, req *dto.MarketingRecordListRequest) (*dto.MarketingRecordListResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	q := s.q.MarketingRecord.WithContext(ctx).
		Where(s.q.MarketingRecord.CampaignID.Eq(req.CampaignID))

	// 构建筛选条件
	if req.CustomerID > 0 {
		q = q.Where(s.q.MarketingRecord.CustomerID.Eq(req.CustomerID))
	}
	if req.Channel != "" {
		q = q.Where(s.q.MarketingRecord.Channel.Eq(req.Channel))
	}
	if req.Status != "" {
		q = q.Where(s.q.MarketingRecord.Status.Eq(req.Status))
	}
	if req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			q = q.Where(s.q.MarketingRecord.CreatedAt.Gte(startTime))
		}
	}
	if req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// 包含结束日期的整天
			endTime = endTime.Add(24 * time.Hour).Add(-time.Second)
			q = q.Where(s.q.MarketingRecord.CreatedAt.Lte(endTime))
		}
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// 分页查询
	records, err := q.Order(s.q.MarketingRecord.CreatedAt.Desc()).
		Limit(req.PageSize).
		Offset((req.Page - 1) * req.PageSize).
		Find()
	if err != nil {
		return nil, err
	}

	// 转换响应
	responses := make([]*dto.MarketingRecordResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, s.toRecordResponse(record))
	}

	return &dto.MarketingRecordListResponse{
		Total:   total,
		Records: responses,
	}, nil
}

// GetCampaignStats 获取营销活动统计
func (s *MarketingService) GetCampaignStats(ctx context.Context, id string) (*dto.MarketingCampaignStatsResponse, error) {
	idNum, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, ErrMarketingCampaignNotFound
	}

	// 1. 获取活动基础信息
	campaign, err := s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(idNum)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMarketingCampaignNotFound
		}
		return nil, err
	}

	// 2. 统计各状态的记录数
	var stats struct {
		DeliveredCount    int64
		FailedCount       int64
		OpenedCount       int64
		ClickedCount      int64
		RepliedCount      int64
		UnsubscribedCount int64
	}

	// 查询各种状态的统计
	deliveredCount, _ := s.q.MarketingRecord.WithContext(ctx).
		Where(s.q.MarketingRecord.CampaignID.Eq(idNum), s.q.MarketingRecord.Status.Eq("delivered")).
		Count()

	failedCount, _ := s.q.MarketingRecord.WithContext(ctx).
		Where(s.q.MarketingRecord.CampaignID.Eq(idNum), s.q.MarketingRecord.Status.Eq("failed")).
		Count()

	openedCount, _ := s.q.MarketingRecord.WithContext(ctx).
		Where(s.q.MarketingRecord.CampaignID.Eq(idNum), s.q.MarketingRecord.Status.Eq("opened")).
		Count()

	clickedCount, _ := s.q.MarketingRecord.WithContext(ctx).
		Where(s.q.MarketingRecord.CampaignID.Eq(idNum), s.q.MarketingRecord.Status.Eq("clicked")).
		Count()

	repliedCount, _ := s.q.MarketingRecord.WithContext(ctx).
		Where(s.q.MarketingRecord.CampaignID.Eq(idNum), s.q.MarketingRecord.Status.Eq("replied")).
		Count()

	unsubscribedCount, _ := s.q.MarketingRecord.WithContext(ctx).
		Where(s.q.MarketingRecord.CampaignID.Eq(idNum), s.q.MarketingRecord.Status.Eq("unsubscribed")).
		Count()

	stats.DeliveredCount = deliveredCount
	stats.FailedCount = failedCount
	stats.OpenedCount = openedCount
	stats.ClickedCount = clickedCount
	stats.RepliedCount = repliedCount
	stats.UnsubscribedCount = unsubscribedCount

	// 3. 计算比率
	var deliveryRate, openRate, clickRate, replyRate, unsubscribeRate float64

	if campaign.SentCount > 0 {
		deliveryRate = float64(stats.DeliveredCount) / float64(campaign.SentCount) * 100
	}
	if stats.DeliveredCount > 0 {
		openRate = float64(stats.OpenedCount) / float64(stats.DeliveredCount) * 100
		clickRate = float64(stats.ClickedCount) / float64(stats.DeliveredCount) * 100
		replyRate = float64(stats.RepliedCount) / float64(stats.DeliveredCount) * 100
		unsubscribeRate = float64(stats.UnsubscribedCount) / float64(stats.DeliveredCount) * 100
	}

	return &dto.MarketingCampaignStatsResponse{
		CampaignID:        campaign.ID,
		CampaignName:      campaign.Name,
		TargetCount:       campaign.TargetCount,
		SentCount:         campaign.SentCount,
		DeliveredCount:    int32(stats.DeliveredCount),
		FailedCount:       int32(stats.FailedCount),
		OpenedCount:       int32(stats.OpenedCount),
		ClickedCount:      int32(stats.ClickedCount),
		RepliedCount:      int32(stats.RepliedCount),
		UnsubscribedCount: int32(stats.UnsubscribedCount),
		DeliveryRate:      deliveryRate,
		OpenRate:          openRate,
		ClickRate:         clickRate,
		ReplyRate:         replyRate,
		UnsubscribeRate:   unsubscribeRate,
	}, nil
}

// ================ 客户分群 ================

// GetCustomerSegment 获取客户分群
func (s *MarketingService) GetCustomerSegment(ctx context.Context, req *dto.CustomerSegmentRequest) (*dto.CustomerSegmentResponse, error) {
	q := s.q.Customer.WithContext(ctx).Where(s.q.Customer.DeletedAt.IsNull())

	// 构建筛选条件
	if len(req.Tags) > 0 {
		// 这里简化处理，假设tags是JSON数组，实际可能需要更复杂的查询
		for _, tag := range req.Tags {
			q = q.Where(s.q.Customer.Tags.Like("%" + tag + "%"))
		}
	}
	if req.Level != "" {
		q = q.Where(s.q.Customer.Level.Eq(req.Level))
	}
	if req.Gender != "" {
		q = q.Where(s.q.Customer.Gender.Eq(req.Gender))
	}
	if req.Source != "" {
		q = q.Where(s.q.Customer.Source.Eq(req.Source))
	}

	// 获取总数
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// 获取客户列表（限制返回数量）
	customers, err := q.Select(s.q.Customer.ID, s.q.Customer.Name, s.q.Customer.Phone, s.q.Customer.Email).
		Limit(1000). // 限制最多返回1000个客户
		Find()
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &dto.CustomerSegmentResponse{
		Total: total,
		Customers: make([]struct {
			ID    int64  `json:"id" example:"100"`
			Name  string `json:"name" example:"张三"`
			Phone string `json:"phone" example:"138****8888"`
			Email string `json:"email,omitempty" example:"zhangsan@example.com"`
		}, len(customers)),
	}

	for i, customer := range customers {
		response.Customers[i].ID = customer.ID
		response.Customers[i].Name = customer.Name
		response.Customers[i].Phone = customer.Phone
		response.Customers[i].Email = customer.Email
	}

	return response, nil
}

// ================ 私有方法 ================

// toCampaignResponse 转换为营销活动响应DTO
func (s *MarketingService) toCampaignResponse(campaign *model.MarketingCampaign) *dto.MarketingCampaignResponse {
	// 解析目标标签
	var targetTags []string
	if campaign.TargetTags != "" {
		json.Unmarshal([]byte(campaign.TargetTags), &targetTags)
	}

	response := &dto.MarketingCampaignResponse{
		ID:                campaign.ID,
		Name:              campaign.Name,
		Type:              campaign.Type,
		Status:            campaign.Status,
		TargetTags:        targetTags,
		TargetSegmentID:   campaign.TargetSegmentID,
		ContentTemplateID: campaign.ContentTemplateID,
		Content:           campaign.Content,
		StartTime:         campaign.StartTime,
		EndTime:           campaign.EndTime,
		TargetCount:       campaign.TargetCount,
		SentCount:         campaign.SentCount,
		SuccessCount:      campaign.SuccessCount,
		ClickCount:        campaign.ClickCount,
		CreatedBy:         campaign.CreatedBy,
		UpdatedBy:         campaign.UpdatedBy,
		CreatedAt:         utils.FormatTime(campaign.CreatedAt),
		UpdatedAt:         utils.FormatTime(campaign.UpdatedAt),
	}

	// 处理可选时间字段
	if !campaign.ActualStartTime.IsZero() {
		response.ActualStartTime = &campaign.ActualStartTime
	}
	if !campaign.ActualEndTime.IsZero() {
		response.ActualEndTime = &campaign.ActualEndTime
	}

	return response
}

// toRecordResponse 转换为营销记录响应DTO
func (s *MarketingService) toRecordResponse(record *model.MarketingRecord) *dto.MarketingRecordResponse {
	response := &dto.MarketingRecordResponse{
		ID:           record.ID,
		CampaignID:   record.CampaignID,
		CustomerID:   record.CustomerID,
		ContactID:    record.ContactID,
		Channel:      record.Channel,
		Status:       record.Status,
		ErrorMessage: record.ErrorMessage,
		Response:     record.Response,
		CreatedAt:    utils.FormatTime(record.CreatedAt),
	}

	// 处理可选时间字段
	if !record.SentAt.IsZero() {
		response.SentAt = &record.SentAt
	}
	if !record.DeliveredAt.IsZero() {
		response.DeliveredAt = &record.DeliveredAt
	}
	if !record.OpenedAt.IsZero() {
		response.OpenedAt = &record.OpenedAt
	}
	if !record.ClickedAt.IsZero() {
		response.ClickedAt = &record.ClickedAt
	}
	if !record.RepliedAt.IsZero() {
		response.RepliedAt = &record.RepliedAt
	}

	return response
}

// getTargetCustomers 获取目标客户列表
func (s *MarketingService) getTargetCustomers(ctx context.Context, campaign *model.MarketingCampaign) ([]*model.Customer, error) {
	q := s.q.Customer.WithContext(ctx).Where(s.q.Customer.DeletedAt.IsNull())

	// 解析目标标签
	var targetTags []string
	if campaign.TargetTags != "" {
		if err := json.Unmarshal([]byte(campaign.TargetTags), &targetTags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal target tags: %w", err)
		}
	}

	// 根据标签筛选客户
	if len(targetTags) > 0 {
		for _, tag := range targetTags {
			q = q.Where(s.q.Customer.Tags.Like("%" + tag + "%"))
		}
	}

	customers, err := q.Find()
	if err != nil {
		return nil, err
	}

	return customers, nil
}

// createMarketingRecords 创建营销记录（异步执行）
func (s *MarketingService) createMarketingRecords(ctx context.Context, campaign *model.MarketingCampaign, customers []*model.Customer) {
	// 这里是简化实现，实际应该根据不同渠道调用不同的发送服务
	// 并根据发送结果更新记录状态

	records := make([]*model.MarketingRecord, 0, len(customers))
	for _, customer := range customers {
		record := &model.MarketingRecord{
			CampaignID: campaign.ID,
			CustomerID: customer.ID,
			Channel:    campaign.Type,
			Status:     "pending",
		}
		records = append(records, record)
	}

	// 批量创建记录
	if err := s.q.MarketingRecord.WithContext(ctx).CreateInBatches(records, 100); err != nil {
		// 记录错误日志
		fmt.Printf("Failed to create marketing records: %v\n", err)
		return
	}

	// 这里应该调用实际的发送服务，如短信、邮件等
	// 发送完成后更新记录状态和活动统计

	// 简化处理：直接更新活动的发送统计
	updates := map[string]interface{}{
		"sent_count":    int32(len(customers)),
		"success_count": int32(len(customers)), // 假设全部发送成功
	}

	s.q.MarketingCampaign.WithContext(ctx).
		Where(s.q.MarketingCampaign.ID.Eq(campaign.ID)).
		Updates(updates)
}
