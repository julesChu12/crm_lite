package validators

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"crm_lite/internal/constants"
)

// RegisterCustomValidators registers all custom validators with gin's validator
func RegisterCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Customer related validators
		_ = v.RegisterValidation("customer_gender", validateCustomerGender)
		_ = v.RegisterValidation("customer_source", validateCustomerSource)
		_ = v.RegisterValidation("customer_level", validateCustomerLevel)

		// Order related validators
		_ = v.RegisterValidation("order_create_status", validateOrderCreateStatus)
		_ = v.RegisterValidation("order_update_status", validateOrderUpdateStatus)

		// Marketing related validators
		_ = v.RegisterValidation("marketing_channel", validateMarketingChannel)
		_ = v.RegisterValidation("marketing_campaign_status", validateMarketingCampaignStatus)
		_ = v.RegisterValidation("marketing_record_status", validateMarketingRecordStatus)
		_ = v.RegisterValidation("marketing_execution_type", validateMarketingExecutionType)

		// Wallet related validators
		_ = v.RegisterValidation("wallet_transaction_type", validateWalletTransactionType)

		// Dashboard related validators
		_ = v.RegisterValidation("date_range", validateDateRange)
		_ = v.RegisterValidation("granularity", validateGranularity)

		// Product related validators
		_ = v.RegisterValidation("product_type", validateProductType)
	}
}

// validateCustomerGender validates customer gender values
func validateCustomerGender(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidCustomerGenders(), value)
}

// validateCustomerSource validates customer source values
func validateCustomerSource(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidCustomerSources(), value)
}

// validateCustomerLevel validates customer level values
func validateCustomerLevel(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidCustomerLevels(), value)
}

// validateOrderCreateStatus validates order creation status values
func validateOrderCreateStatus(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidOrderCreateStatuses(), value)
}

// validateOrderUpdateStatus validates order update status values
func validateOrderUpdateStatus(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidOrderUpdateStatuses(), value)
}

// validateMarketingChannel validates marketing channel type values
func validateMarketingChannel(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidMarketingChannelTypes(), value)
}

// validateMarketingCampaignStatus validates marketing campaign status values
func validateMarketingCampaignStatus(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidMarketingCampaignStatuses(), value)
}

// validateMarketingRecordStatus validates marketing record status values
func validateMarketingRecordStatus(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidMarketingRecordStatuses(), value)
}

// validateMarketingExecutionType validates marketing execution type values
func validateMarketingExecutionType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidMarketingExecutionTypes(), value)
}

// validateWalletTransactionType validates wallet transaction type values
func validateWalletTransactionType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return contains(constants.ValidWalletTransactionTypes(), value)
}

// validateDateRange validates date range values
func validateDateRange(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidDateRanges(), value)
}

// validateGranularity validates granularity values
func validateGranularity(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidGranularities(), value)
}

// validateProductType validates product type values
func validateProductType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values for optional fields
	}
	return contains(constants.ValidProductTypes(), value)
}

// contains checks if a slice contains a specific string value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// GetValidationErrorMessage returns user-friendly error messages for custom validations
func GetValidationErrorMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()

	switch tag {
	case "customer_gender":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidCustomerGenders(), ", "))
	case "customer_source":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidCustomerSources(), ", "))
	case "customer_level":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidCustomerLevels(), ", "))
	case "order_create_status":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidOrderCreateStatuses(), ", "))
	case "order_update_status":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidOrderUpdateStatuses(), ", "))
	case "marketing_channel":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidMarketingChannelTypes(), ", "))
	case "marketing_campaign_status":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidMarketingCampaignStatuses(), ", "))
	case "marketing_record_status":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidMarketingRecordStatuses(), ", "))
	case "marketing_execution_type":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidMarketingExecutionTypes(), ", "))
	case "wallet_transaction_type":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidWalletTransactionTypes(), ", "))
	case "date_range":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidDateRanges(), ", "))
	case "granularity":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidGranularities(), ", "))
	case "product_type":
		return fmt.Sprintf("%s must be one of: %s", field, strings.Join(constants.ValidProductTypes(), ", "))
	default:
		return fmt.Sprintf("%s validation failed", field)
	}
}