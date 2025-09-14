// Package marketing 营销推广域服务接口
// 职责：营销活动管理、营销记录管理、推广效果分析
// 核心原则：营销活动生命周期管理、推广效果跟踪、ROI分析
package marketing

import (
	"context"
	"time"
)

// Campaign 营销活动领域模型
type Campaign struct {
	ID          int64     `json:"id"`           // 活动ID
	Name        string    `json:"name"`         // 活动名称
	Description string    `json:"description"`  // 活动描述
	Type        string    `json:"type"`         // 活动类型：promotion/discount/event
	Channel     string    `json:"channel"`      // 推广渠道：sms/email/wechat/app
	Status      string    `json:"status"`       // 状态：draft/active/paused/completed
	StartTime   time.Time `json:"start_time"`   // 开始时间
	EndTime     time.Time `json:"end_time"`     // 结束时间
	Budget      int64     `json:"budget"`       // 预算（分）
	Spent       int64     `json:"spent"`        // 已花费（分）
	TargetCount int64     `json:"target_count"` // 目标触达数
	ActualCount int64     `json:"actual_count"` // 实际触达数
	CreatedBy   int64     `json:"created_by"`   // 创建人ID
	CreatedAt   int64     `json:"created_at"`   // 创建时间
	UpdatedAt   int64     `json:"updated_at"`   // 更新时间
}

// Record 营销记录领域模型
type Record struct {
	ID          int64  `json:"id"`           // 记录ID
	CampaignID  int64  `json:"campaign_id"`  // 关联活动ID
	CustomerID  int64  `json:"customer_id"`  // 目标客户ID
	Channel     string `json:"channel"`      // 推广渠道
	Content     string `json:"content"`      // 推广内容
	Status      string `json:"status"`       // 状态：pending/sent/delivered/failed
	SentAt      int64  `json:"sent_at"`      // 发送时间
	DeliveredAt int64  `json:"delivered_at"` // 送达时间
	OpenedAt    int64  `json:"opened_at"`    // 打开时间
	ClickedAt   int64  `json:"clicked_at"`   // 点击时间
	Cost        int64  `json:"cost"`         // 推广成本（分）
	CreatedAt   int64  `json:"created_at"`   // 创建时间
}

// CreateCampaignRequest 创建活动请求
type CreateCampaignRequest struct {
	Name        string    `json:"name"`         // 活动名称
	Description string    `json:"description"`  // 活动描述
	Type        string    `json:"type"`         // 活动类型
	Channel     string    `json:"channel"`      // 推广渠道
	StartTime   time.Time `json:"start_time"`   // 开始时间
	EndTime     time.Time `json:"end_time"`     // 结束时间
	Budget      int64     `json:"budget"`       // 预算（分）
	TargetCount int64     `json:"target_count"` // 目标触达数
}

// UpdateCampaignRequest 更新活动请求
type UpdateCampaignRequest struct {
	Name        *string    `json:"name,omitempty"`         // 活动名称
	Description *string    `json:"description,omitempty"`  // 活动描述
	Status      *string    `json:"status,omitempty"`       // 状态
	StartTime   *time.Time `json:"start_time,omitempty"`   // 开始时间
	EndTime     *time.Time `json:"end_time,omitempty"`     // 结束时间
	Budget      *int64     `json:"budget,omitempty"`       // 预算
	TargetCount *int64     `json:"target_count,omitempty"` // 目标触达数
}

// CreateRecordRequest 创建记录请求
type CreateRecordRequest struct {
	CampaignID int64  `json:"campaign_id"` // 关联活动ID
	CustomerID int64  `json:"customer_id"` // 目标客户ID
	Channel    string `json:"channel"`     // 推广渠道
	Content    string `json:"content"`     // 推广内容
}

// UpdateRecordRequest 更新记录请求
type UpdateRecordRequest struct {
	Status      *string `json:"status,omitempty"`       // 状态
	SentAt      *int64  `json:"sent_at,omitempty"`      // 发送时间
	DeliveredAt *int64  `json:"delivered_at,omitempty"` // 送达时间
	OpenedAt    *int64  `json:"opened_at,omitempty"`    // 打开时间
	ClickedAt   *int64  `json:"clicked_at,omitempty"`   // 点击时间
	Cost        *int64  `json:"cost,omitempty"`         // 成本
}

// CampaignStats 活动统计
type CampaignStats struct {
	CampaignID     int64   `json:"campaign_id"`     // 活动ID
	TotalRecords   int64   `json:"total_records"`   // 总记录数
	SentCount      int64   `json:"sent_count"`      // 发送数量
	DeliveredCount int64   `json:"delivered_count"` // 送达数量
	OpenedCount    int64   `json:"opened_count"`    // 打开数量
	ClickedCount   int64   `json:"clicked_count"`   // 点击数量
	TotalCost      int64   `json:"total_cost"`      // 总成本
	DeliveryRate   float64 `json:"delivery_rate"`   // 送达率
	OpenRate       float64 `json:"open_rate"`       // 打开率
	ClickRate      float64 `json:"click_rate"`      // 点击率
	CostPerClick   float64 `json:"cost_per_click"`  // 每次点击成本
	ROI            float64 `json:"roi"`             // 投资回报率
}

// CampaignService 营销活动服务接口
type CampaignService interface {
	// CreateCampaign 创建营销活动
	CreateCampaign(ctx context.Context, req CreateCampaignRequest, createdBy int64) (*Campaign, error)

	// GetCampaign 获取活动详情
	GetCampaign(ctx context.Context, campaignID int64) (*Campaign, error)

	// ListCampaigns 分页查询活动列表
	ListCampaigns(ctx context.Context, status string, page, pageSize int) ([]Campaign, int64, error)

	// UpdateCampaign 更新活动信息
	UpdateCampaign(ctx context.Context, campaignID int64, req UpdateCampaignRequest) error

	// DeleteCampaign 删除活动
	DeleteCampaign(ctx context.Context, campaignID int64) error

	// StartCampaign 启动活动
	StartCampaign(ctx context.Context, campaignID int64) error

	// PauseCampaign 暂停活动
	PauseCampaign(ctx context.Context, campaignID int64) error

	// CompleteCampaign 完成活动
	CompleteCampaign(ctx context.Context, campaignID int64) error
}

// RecordService 营销记录服务接口
type RecordService interface {
	// CreateRecord 创建营销记录
	CreateRecord(ctx context.Context, req CreateRecordRequest) (*Record, error)

	// GetRecord 获取记录详情
	GetRecord(ctx context.Context, recordID int64) (*Record, error)

	// ListRecords 分页查询记录列表
	ListRecords(ctx context.Context, campaignID int64, customerID int64, page, pageSize int) ([]Record, int64, error)

	// UpdateRecord 更新记录状态
	UpdateRecord(ctx context.Context, recordID int64, req UpdateRecordRequest) error

	// DeleteRecord 删除记录
	DeleteRecord(ctx context.Context, recordID int64) error

	// BatchCreateRecords 批量创建记录
	BatchCreateRecords(ctx context.Context, campaignID int64, customerIDs []int64, content string) error

	// UpdateRecordStatus 更新记录状态（用于推送回调）
	UpdateRecordStatus(ctx context.Context, recordID int64, status string, timestamp int64) error
}

// AnalyticsService 营销分析服务接口
type AnalyticsService interface {
	// GetCampaignStats 获取活动统计
	GetCampaignStats(ctx context.Context, campaignID int64) (*CampaignStats, error)

	// GetCampaignPerformance 获取活动效果对比
	GetCampaignPerformance(ctx context.Context, campaignIDs []int64) ([]CampaignStats, error)

	// GetChannelStats 获取渠道统计
	GetChannelStats(ctx context.Context, channel string, startTime, endTime time.Time) (*CampaignStats, error)

	// GetCustomerEngagement 获取客户参与度
	GetCustomerEngagement(ctx context.Context, customerID int64) ([]Record, error)
}

// Service 营销域统一服务接口
// 整合活动管理、记录管理、数据分析的完整功能
type Service interface {
	CampaignService
	RecordService
	AnalyticsService
}
