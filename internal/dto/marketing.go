package dto

import "time"

// ================ 营销活动相关DTO ================

// MarketingCampaignCreateRequest 创建营销活动请求
type MarketingCampaignCreateRequest struct {
	Name              string    `json:"name" binding:"required" example:"六月会员关怀活动"`
	Type              string    `json:"type" binding:"required,marketing_channel" example:"sms"`
	TargetTags        []string  `json:"target_tags" example:"[\"VIP\", \"新用户\"]"`
	TargetSegmentID   int64     `json:"target_segment_id,omitempty" example:"0"`
	ContentTemplateID int64     `json:"content_template_id,omitempty" example:"0"`
	Content           string    `json:"content" binding:"required" example:"亲爱的{name}，您有一份专属优惠待领取！"`
	StartTime         time.Time `json:"start_time" binding:"required" example:"2024-06-10T09:00:00Z"`
	EndTime           time.Time `json:"end_time" binding:"required" example:"2024-06-15T23:59:59Z"`
}

// MarketingCampaignUpdateRequest 更新营销活动请求
type MarketingCampaignUpdateRequest struct {
	Name              string     `json:"name,omitempty" example:"六月会员关怀活动（更新）"`
	Type              string     `json:"type,omitempty" binding:"omitempty,marketing_channel"`
	Status            string     `json:"status,omitempty" binding:"omitempty,marketing_campaign_status"`
	TargetTags        []string   `json:"target_tags,omitempty"`
	TargetSegmentID   int64      `json:"target_segment_id,omitempty"`
	ContentTemplateID int64      `json:"content_template_id,omitempty"`
	Content           string     `json:"content,omitempty"`
	StartTime         *time.Time `json:"start_time,omitempty"`
	EndTime           *time.Time `json:"end_time,omitempty"`
}

// MarketingCampaignResponse 营销活动响应
type MarketingCampaignResponse struct {
	ID                int64      `json:"id" example:"1"`
	Name              string     `json:"name" example:"六月会员关怀活动"`
	Type              string     `json:"type" example:"sms"`
	Status            string     `json:"status" example:"active"`
	TargetTags        []string   `json:"target_tags" example:"[\"VIP\", \"新用户\"]"`
	TargetSegmentID   int64      `json:"target_segment_id" example:"0"`
	ContentTemplateID int64      `json:"content_template_id" example:"0"`
	Content           string     `json:"content" example:"亲爱的{name}，您有一份专属优惠待领取！"`
	StartTime         time.Time  `json:"start_time" example:"2024-06-10T09:00:00Z"`
	EndTime           time.Time  `json:"end_time" example:"2024-06-15T23:59:59Z"`
	ActualStartTime   *time.Time `json:"actual_start_time,omitempty" example:"2024-06-10T09:05:00Z"`
	ActualEndTime     *time.Time `json:"actual_end_time,omitempty"`
	TargetCount       int32      `json:"target_count" example:"1500"`
	SentCount         int32      `json:"sent_count" example:"1450"`
	SuccessCount      int32      `json:"success_count" example:"1400"`
	ClickCount        int32      `json:"click_count" example:"280"`
	CreatedBy         int64      `json:"created_by" example:"1"`
	UpdatedBy         int64      `json:"updated_by" example:"1"`
	CreatedAt         string     `json:"created_at" example:"2024-06-01T10:00:00Z"`
	UpdatedAt         string     `json:"updated_at" example:"2024-06-10T14:30:00Z"`
}

// MarketingCampaignListRequest 营销活动列表请求
type MarketingCampaignListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	Name     string `form:"name" example:"会员活动"`
	Type     string `form:"type" binding:"omitempty,marketing_channel"`
	Status   string `form:"status" binding:"omitempty,marketing_campaign_status"`
	OrderBy  string `form:"order_by" example:"created_at_desc"`
}

// MarketingCampaignListResponse 营销活动列表响应
type MarketingCampaignListResponse struct {
	Total     int64                        `json:"total" example:"50"`
	Campaigns []*MarketingCampaignResponse `json:"campaigns"`
}

// MarketingCampaignExecuteRequest 执行营销活动请求
type MarketingCampaignExecuteRequest struct {
	ExecutionType string `json:"execution_type" binding:"omitempty,marketing_execution_type" example:"actual"`
}

// MarketingCampaignExecuteResponse 执行营销活动响应
type MarketingCampaignExecuteResponse struct {
	Status      string `json:"status" example:"triggered"`
	Message     string `json:"message" example:"营销活动已成功触发执行"`
	ExecutionID string `json:"execution_id,omitempty" example:"exec-12345"`
}

// ================ 营销记录相关DTO ================

// MarketingRecordResponse 营销记录响应
type MarketingRecordResponse struct {
	ID           int64      `json:"id" example:"1"`
	CampaignID   int64      `json:"campaign_id" example:"1"`
	CustomerID   int64      `json:"customer_id" example:"100"`
	ContactID    int64      `json:"contact_id" example:"1"`
	Channel      string     `json:"channel" example:"sms"`
	Status       string     `json:"status" example:"delivered"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Response     string     `json:"response,omitempty" example:"{\"action\":\"click\",\"link_id\":\"promo123\"}"`
	SentAt       *time.Time `json:"sent_at,omitempty" example:"2024-06-10T09:05:00Z"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty" example:"2024-06-10T09:06:00Z"`
	OpenedAt     *time.Time `json:"opened_at,omitempty" example:"2024-06-10T10:15:00Z"`
	ClickedAt    *time.Time `json:"clicked_at,omitempty" example:"2024-06-10T10:16:00Z"`
	RepliedAt    *time.Time `json:"replied_at,omitempty"`
	CreatedAt    string     `json:"created_at" example:"2024-06-10T09:00:00Z"`
}

// MarketingRecordListRequest 营销记录列表请求
type MarketingRecordListRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"50"`
	CampaignID int64  `form:"campaign_id" binding:"required" example:"1"`
	CustomerID int64  `form:"customer_id,omitempty" example:"100"`
	Channel    string `form:"channel,omitempty" binding:"omitempty,marketing_channel"`
	Status     string `form:"status,omitempty" binding:"omitempty,marketing_record_status"`
	StartDate  string `form:"start_date,omitempty" example:"2024-06-01"`
	EndDate    string `form:"end_date,omitempty" example:"2024-06-30"`
}

// MarketingRecordListResponse 营销记录列表响应
type MarketingRecordListResponse struct {
	Total   int64                      `json:"total" example:"1500"`
	Records []*MarketingRecordResponse `json:"records"`
}

// ================ 营销统计相关DTO ================

// MarketingCampaignStatsResponse 营销活动统计响应
type MarketingCampaignStatsResponse struct {
	CampaignID        int64  `json:"campaign_id" example:"1"`
	CampaignName      string `json:"campaign_name" example:"六月会员关怀活动"`
	TargetCount       int32  `json:"target_count" example:"1500"`
	SentCount         int32  `json:"sent_count" example:"1450"`
	DeliveredCount    int32  `json:"delivered_count" example:"1400"`
	FailedCount       int32  `json:"failed_count" example:"50"`
	OpenedCount       int32  `json:"opened_count" example:"420"`
	ClickedCount      int32  `json:"clicked_count" example:"280"`
	RepliedCount      int32  `json:"replied_count" example:"35"`
	UnsubscribedCount int32  `json:"unsubscribed_count" example:"12"`

	// 计算比率
	DeliveryRate    float64 `json:"delivery_rate" example:"96.55"`   // 送达率 = delivered / sent * 100
	OpenRate        float64 `json:"open_rate" example:"30.00"`       // 打开率 = opened / delivered * 100
	ClickRate       float64 `json:"click_rate" example:"20.00"`      // 点击率 = clicked / delivered * 100
	ReplyRate       float64 `json:"reply_rate" example:"2.50"`       // 回复率 = replied / delivered * 100
	UnsubscribeRate float64 `json:"unsubscribe_rate" example:"0.86"` // 退订率 = unsubscribed / delivered * 100
}

// ================ 客户分群相关DTO ================

// CustomerSegmentRequest 客户分群请求
type CustomerSegmentRequest struct {
	Tags   []string `json:"tags,omitempty" example:"[\"VIP\", \"新用户\"]"`
	Level  string   `json:"level,omitempty" example:"VIP"`
	Gender string   `json:"gender,omitempty" example:"male"`
	AgeMin int      `json:"age_min,omitempty" example:"18"`
	AgeMax int      `json:"age_max,omitempty" example:"65"`
	Source string   `json:"source,omitempty" example:"微信"`
}

// CustomerSegmentResponse 客户分群响应
type CustomerSegmentResponse struct {
	Total     int64 `json:"total" example:"500"`
	Customers []struct {
		ID    int64  `json:"id" example:"100"`
		Name  string `json:"name" example:"张三"`
		Phone string `json:"phone" example:"138****8888"`
		Email string `json:"email,omitempty" example:"zhangsan@example.com"`
	} `json:"customers"`
}
