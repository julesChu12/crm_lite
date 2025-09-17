package constants

// Business domain constants for validation and business logic

// ================ Customer Related Constants ================

// CustomerGender defines valid customer gender values
type CustomerGender string

const (
	CustomerGenderMale    CustomerGender = "male"
	CustomerGenderFemale  CustomerGender = "female"
	CustomerGenderUnknown CustomerGender = "unknown"
)

// ValidCustomerGenders returns all valid customer gender values
func ValidCustomerGenders() []string {
	return []string{
		string(CustomerGenderMale),
		string(CustomerGenderFemale),
		string(CustomerGenderUnknown),
	}
}

// CustomerSource defines valid customer source values
type CustomerSource string

const (
	CustomerSourceManual    CustomerSource = "manual"
	CustomerSourceReferral  CustomerSource = "referral"
	CustomerSourceMarketing CustomerSource = "marketing"
)

// ValidCustomerSources returns all valid customer source values
func ValidCustomerSources() []string {
	return []string{
		string(CustomerSourceManual),
		string(CustomerSourceReferral),
		string(CustomerSourceMarketing),
	}
}

// CustomerLevel defines valid customer level values
type CustomerLevel string

const (
	CustomerLevelNormal   CustomerLevel = "普通"
	CustomerLevelSilver   CustomerLevel = "银牌"
	CustomerLevelGold     CustomerLevel = "金牌"
	CustomerLevelPlatinum CustomerLevel = "铂金"
)

// ValidCustomerLevels returns all valid customer level values
func ValidCustomerLevels() []string {
	return []string{
		string(CustomerLevelNormal),
		string(CustomerLevelSilver),
		string(CustomerLevelGold),
		string(CustomerLevelPlatinum),
	}
}

// ================ Order Related Constants ================

// OrderStatus defines valid order status values
type OrderStatus string

const (
	// Basic statuses for order creation
	OrderStatusDraft     OrderStatus = "draft"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"

	// Extended statuses for order updates
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// ValidOrderCreateStatuses returns valid statuses for order creation
func ValidOrderCreateStatuses() []string {
	return []string{
		string(OrderStatusDraft),
		string(OrderStatusPending),
		string(OrderStatusConfirmed),
	}
}

// ValidOrderUpdateStatuses returns valid statuses for order updates
func ValidOrderUpdateStatuses() []string {
	return []string{
		string(OrderStatusDraft),
		string(OrderStatusPending),
		string(OrderStatusConfirmed),
		string(OrderStatusProcessing),
		string(OrderStatusShipped),
		string(OrderStatusCompleted),
		string(OrderStatusCancelled),
		string(OrderStatusRefunded),
	}
}

// ================ Marketing Related Constants ================

// MarketingChannelType defines valid marketing channel types
type MarketingChannelType string

const (
	MarketingChannelSMS              MarketingChannelType = "sms"
	MarketingChannelEmail            MarketingChannelType = "email"
	MarketingChannelPushNotification MarketingChannelType = "push_notification"
	MarketingChannelWechat           MarketingChannelType = "wechat"
	MarketingChannelCall             MarketingChannelType = "call"
)

// ValidMarketingChannelTypes returns all valid marketing channel types
func ValidMarketingChannelTypes() []string {
	return []string{
		string(MarketingChannelSMS),
		string(MarketingChannelEmail),
		string(MarketingChannelPushNotification),
		string(MarketingChannelWechat),
		string(MarketingChannelCall),
	}
}

// MarketingCampaignStatus defines valid marketing campaign status values
type MarketingCampaignStatus string

const (
	MarketingCampaignStatusDraft     MarketingCampaignStatus = "draft"
	MarketingCampaignStatusScheduled MarketingCampaignStatus = "scheduled"
	MarketingCampaignStatusActive    MarketingCampaignStatus = "active"
	MarketingCampaignStatusPaused    MarketingCampaignStatus = "paused"
	MarketingCampaignStatusCompleted MarketingCampaignStatus = "completed"
	MarketingCampaignStatusArchived  MarketingCampaignStatus = "archived"
)

// ValidMarketingCampaignStatuses returns all valid marketing campaign statuses
func ValidMarketingCampaignStatuses() []string {
	return []string{
		string(MarketingCampaignStatusDraft),
		string(MarketingCampaignStatusScheduled),
		string(MarketingCampaignStatusActive),
		string(MarketingCampaignStatusPaused),
		string(MarketingCampaignStatusCompleted),
		string(MarketingCampaignStatusArchived),
	}
}

// MarketingRecordStatus defines valid marketing record status values
type MarketingRecordStatus string

const (
	MarketingRecordStatusPending      MarketingRecordStatus = "pending"
	MarketingRecordStatusSent         MarketingRecordStatus = "sent"
	MarketingRecordStatusDelivered    MarketingRecordStatus = "delivered"
	MarketingRecordStatusFailed       MarketingRecordStatus = "failed"
	MarketingRecordStatusOpened       MarketingRecordStatus = "opened"
	MarketingRecordStatusClicked      MarketingRecordStatus = "clicked"
	MarketingRecordStatusReplied      MarketingRecordStatus = "replied"
	MarketingRecordStatusUnsubscribed MarketingRecordStatus = "unsubscribed"
)

// ValidMarketingRecordStatuses returns all valid marketing record statuses
func ValidMarketingRecordStatuses() []string {
	return []string{
		string(MarketingRecordStatusPending),
		string(MarketingRecordStatusSent),
		string(MarketingRecordStatusDelivered),
		string(MarketingRecordStatusFailed),
		string(MarketingRecordStatusOpened),
		string(MarketingRecordStatusClicked),
		string(MarketingRecordStatusReplied),
		string(MarketingRecordStatusUnsubscribed),
	}
}

// MarketingExecutionType defines valid marketing execution types
type MarketingExecutionType string

const (
	MarketingExecutionTypeActual     MarketingExecutionType = "actual"
	MarketingExecutionTypeSimulation MarketingExecutionType = "simulation"
)

// ValidMarketingExecutionTypes returns all valid marketing execution types
func ValidMarketingExecutionTypes() []string {
	return []string{
		string(MarketingExecutionTypeActual),
		string(MarketingExecutionTypeSimulation),
	}
}

// ================ Wallet Related Constants ================

// WalletTransactionType defines valid wallet transaction types
type WalletTransactionType string

const (
	WalletTransactionTypeRecharge WalletTransactionType = "recharge"
	WalletTransactionTypeConsume  WalletTransactionType = "consume"
	WalletTransactionTypeRefund   WalletTransactionType = "refund"
)

// ValidWalletTransactionTypes returns all valid wallet transaction types
func ValidWalletTransactionTypes() []string {
	return []string{
		string(WalletTransactionTypeRecharge),
		string(WalletTransactionTypeConsume),
		string(WalletTransactionTypeRefund),
	}
}

// ================ Dashboard Related Constants ================

// DateRange defines valid date range values for analytics
type DateRange string

const (
	DateRangeToday   DateRange = "today"
	DateRangeWeek    DateRange = "week"
	DateRangeMonth   DateRange = "month"
	DateRangeQuarter DateRange = "quarter"
	DateRangeYear    DateRange = "year"
)

// ValidDateRanges returns all valid date range values
func ValidDateRanges() []string {
	return []string{
		string(DateRangeToday),
		string(DateRangeWeek),
		string(DateRangeMonth),
		string(DateRangeQuarter),
		string(DateRangeYear),
	}
}

// Granularity defines valid granularity values for analytics
type Granularity string

const (
	GranularityDay   Granularity = "day"
	GranularityWeek  Granularity = "week"
	GranularityMonth Granularity = "month"
)

// ValidGranularities returns all valid granularity values
func ValidGranularities() []string {
	return []string{
		string(GranularityDay),
		string(GranularityWeek),
		string(GranularityMonth),
	}
}

// ================ Product Related Constants ================

// ProductType defines valid product types
type ProductType string

const (
	ProductTypeProduct ProductType = "product"
	ProductTypeService ProductType = "service"
)

// ValidProductTypes returns all valid product types
func ValidProductTypes() []string {
	return []string{
		string(ProductTypeProduct),
		string(ProductTypeService),
	}
}