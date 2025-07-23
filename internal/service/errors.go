package service

import "errors"

// 公共的业务错误定义
var (
	// 用户模块相关错误
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidToken      = errors.New("invalid or expired token")

	// 角色模块相关错误
	ErrRoleNotFound          = errors.New("role not found")
	ErrRoleNameAlreadyExists = errors.New("role name already exists")
	ErrUserRoleNotFound      = errors.New("user role assignment not found")

	// 权限模块相关错误
	ErrPermissionNotFound = errors.New("permission not found")

	// 邮箱/手机号相关错误（通用）
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPhoneAlreadyExists = errors.New("phone number already exists")

	// 客户模块相关错误
	ErrCustomerNotFound = errors.New("customer not found")

	// 联系人模块相关错误
	ErrContactNotFound             = errors.New("contact not found")
	ErrPrimaryContactAlreadyExists = errors.New("primary contact already exists for this customer")
	ErrContactPhoneAlreadyExists   = errors.New("contact phone number already exists")
	ErrContactEmailAlreadyExists   = errors.New("contact email already exists")

	// 钱包模块相关错误
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrForbidden           = errors.New("operation not permitted")
	ErrInsufficientBalance = errors.New("insufficient balance")

	// 营销模块相关错误
	ErrMarketingCampaignNotFound      = errors.New("marketing campaign not found")
	ErrMarketingCampaignNameExists    = errors.New("marketing campaign name already exists")
	ErrMarketingCampaignInvalidStatus = errors.New("invalid marketing campaign status transition")
	ErrMarketingCampaignAlreadyActive = errors.New("marketing campaign is already active")
	ErrMarketingCampaignCannotModify  = errors.New("cannot modify campaign in current status")
	ErrMarketingRecordNotFound        = errors.New("marketing record not found")
	ErrMarketingInvalidTimeRange      = errors.New("invalid time range: start time must be before end time")
	ErrMarketingInvalidChannel        = errors.New("invalid marketing channel")
	ErrMarketingNoTargetCustomers     = errors.New("no target customers found for this campaign")
	ErrMarketingExecutionFailed       = errors.New("marketing campaign execution failed")
)
