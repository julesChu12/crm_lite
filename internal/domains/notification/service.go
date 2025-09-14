// Package notification 通知服务域
// 职责：邮件通知、短信通知、推送通知、消息队列
// 核心原则：多渠道通知、模板管理、异步处理、失败重试
package notification

import "context"

// NotificationChannel 通知渠道类型
type NotificationChannel string

const (
	ChannelEmail  NotificationChannel = "email"
	ChannelSMS    NotificationChannel = "sms"
	ChannelPush   NotificationChannel = "push"
	ChannelWechat NotificationChannel = "wechat"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"   // 待发送
	StatusSending   NotificationStatus = "sending"   // 发送中
	StatusSent      NotificationStatus = "sent"      // 已发送
	StatusDelivered NotificationStatus = "delivered" // 已送达
	StatusFailed    NotificationStatus = "failed"    // 发送失败
	StatusRetrying  NotificationStatus = "retrying"  // 重试中
)

// Notification 通知记录领域模型
type Notification struct {
	ID        int64               `json:"id"`         // 通知ID
	Channel   NotificationChannel `json:"channel"`    // 通知渠道
	Recipient string              `json:"recipient"`  // 接收者
	Subject   string              `json:"subject"`    // 主题
	Content   string              `json:"content"`    // 内容
	Template  string              `json:"template"`   // 模板ID
	Variables map[string]string   `json:"variables"`  // 模板变量
	Status    NotificationStatus  `json:"status"`     // 状态
	Error     string              `json:"error"`      // 错误信息
	Attempts  int                 `json:"attempts"`   // 重试次数
	SentAt    int64               `json:"sent_at"`    // 发送时间
	CreatedAt int64               `json:"created_at"` // 创建时间
	UpdatedAt int64               `json:"updated_at"` // 更新时间
}

// Template 通知模板领域模型
type Template struct {
	ID        string              `json:"id"`         // 模板ID
	Name      string              `json:"name"`       // 模板名称
	Channel   NotificationChannel `json:"channel"`    // 适用渠道
	Subject   string              `json:"subject"`    // 主题模板
	Content   string              `json:"content"`    // 内容模板
	IsActive  bool                `json:"is_active"`  // 是否启用
	CreatedAt int64               `json:"created_at"` // 创建时间
}

// SendRequest 发送通知请求
type SendRequest struct {
	Channel   NotificationChannel `json:"channel"`   // 通知渠道
	Recipient string              `json:"recipient"` // 接收者
	Subject   string              `json:"subject"`   // 主题
	Content   string              `json:"content"`   // 内容
	Template  string              `json:"template"`  // 模板ID（可选）
	Variables map[string]string   `json:"variables"` // 模板变量（可选）
	Priority  int                 `json:"priority"`  // 优先级（1-5）
}

// BatchSendRequest 批量发送通知请求
type BatchSendRequest struct {
	Channel    NotificationChannel `json:"channel"`    // 通知渠道
	Recipients []string            `json:"recipients"` // 接收者列表
	Subject    string              `json:"subject"`    // 主题
	Content    string              `json:"content"`    // 内容
	Template   string              `json:"template"`   // 模板ID（可选）
	Variables  map[string]string   `json:"variables"`  // 模板变量（可选）
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Host         string `json:"host"`          // SMTP服务器
	Port         int    `json:"port"`          // SMTP端口
	Username     string `json:"username"`      // 用户名
	Password     string `json:"password"`      // 密码
	FromAddress  string `json:"from_address"`  // 发件人地址
	FromName     string `json:"from_name"`     // 发件人名称
	InsecureSkip bool   `json:"insecure_skip"` // 跳过TLS验证
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider  string `json:"provider"`   // 服务商
	APIKey    string `json:"api_key"`    // API密钥
	APISecret string `json:"api_secret"` // API密钥
	SignName  string `json:"sign_name"`  // 签名
}

// EmailService 邮件服务接口
type EmailService interface {
	// SendEmail 发送邮件
	SendEmail(ctx context.Context, to, subject, body string) error

	// SendEmailWithTemplate 使用模板发送邮件
	SendEmailWithTemplate(ctx context.Context, to, templateID string, variables map[string]string) error

	// BatchSendEmail 批量发送邮件
	BatchSendEmail(ctx context.Context, recipients []string, subject, body string) error

	// ValidateEmailAddress 验证邮箱地址格式
	ValidateEmailAddress(email string) bool
}

// SMSService 短信服务接口
type SMSService interface {
	// SendSMS 发送短信
	SendSMS(ctx context.Context, to, content string) error

	// SendSMSWithTemplate 使用模板发送短信
	SendSMSWithTemplate(ctx context.Context, to, templateID string, variables map[string]string) error

	// BatchSendSMS 批量发送短信
	BatchSendSMS(ctx context.Context, recipients []string, content string) error

	// ValidatePhoneNumber 验证手机号格式
	ValidatePhoneNumber(phone string) bool
}

// TemplateService 模板服务接口
type TemplateService interface {
	// CreateTemplate 创建通知模板
	CreateTemplate(ctx context.Context, template Template) error

	// GetTemplate 获取模板详情
	GetTemplate(ctx context.Context, templateID string) (*Template, error)

	// ListTemplates 查询模板列表
	ListTemplates(ctx context.Context, channel NotificationChannel) ([]Template, error)

	// UpdateTemplate 更新模板
	UpdateTemplate(ctx context.Context, templateID string, template Template) error

	// DeleteTemplate 删除模板
	DeleteTemplate(ctx context.Context, templateID string) error

	// RenderTemplate 渲染模板
	RenderTemplate(ctx context.Context, templateID string, variables map[string]string) (subject, content string, err error)
}

// NotificationService 通知记录服务接口
type NotificationService interface {
	// CreateNotification 创建通知记录
	CreateNotification(ctx context.Context, req SendRequest) (*Notification, error)

	// GetNotification 获取通知详情
	GetNotification(ctx context.Context, notificationID int64) (*Notification, error)

	// ListNotifications 查询通知列表
	ListNotifications(ctx context.Context, channel NotificationChannel, status NotificationStatus, page, pageSize int) ([]Notification, int64, error)

	// UpdateNotificationStatus 更新通知状态
	UpdateNotificationStatus(ctx context.Context, notificationID int64, status NotificationStatus, error string) error

	// RetryNotification 重试发送通知
	RetryNotification(ctx context.Context, notificationID int64) error

	// GetNotificationStats 获取通知统计
	GetNotificationStats(ctx context.Context, channel NotificationChannel, days int) (map[string]int64, error)
}

// Service 通知服务域统一接口
// 整合邮件、短信、模板、记录管理的完整功能
type Service interface {
	EmailService
	SMSService
	TemplateService
	NotificationService

	// Send 统一发送接口
	Send(ctx context.Context, req SendRequest) (*Notification, error)

	// BatchSend 批量发送接口
	BatchSend(ctx context.Context, req BatchSendRequest) ([]Notification, error)

	// ProcessPendingNotifications 处理待发送通知
	ProcessPendingNotifications(ctx context.Context, limit int) error

	// ProcessFailedNotifications 重试失败通知
	ProcessFailedNotifications(ctx context.Context, maxAttempts int, limit int) error
}
