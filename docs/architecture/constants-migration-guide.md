# Constants Refactoring Migration Guide

This document provides guidance for migrating from hardcoded validation strings to the new centralized constants system.

## Overview

We have replaced hardcoded string literals in `binding:"oneof=..."` tags with custom validators that reference centralized constants. This provides:

1. **Type Safety**: Constants are defined as typed strings
2. **Maintainability**: Single source of truth for all business constants
3. **Consistency**: Unified validation approach across DTOs and domain models
4. **Better Error Messages**: Centralized error message handling

## Migration Examples

### Before (Hardcoded)
```go
// Old approach with hardcoded strings
type OrderCreateRequest struct {
    Status string `json:"status" binding:"omitempty,oneof=draft pending confirmed"`
}

type MarketingCampaignCreateRequest struct {
    Type string `json:"type" binding:"required,oneof=sms email push_notification wechat call"`
}
```

### After (Constants + Custom Validators)
```go
// New approach with custom validators
type OrderCreateRequest struct {
    Status string `json:"status" binding:"omitempty,order_create_status"`
}

type MarketingCampaignCreateRequest struct {
    Type string `json:"type" binding:"required,marketing_channel"`
}
```

## Available Custom Validators

### Customer Related
- `customer_gender`: Validates against `male`, `female`, `unknown`
- `customer_source`: Validates against `manual`, `referral`, `marketing`
- `customer_level`: Validates against `普通`, `银牌`, `金牌`, `铂金`

### Order Related
- `order_create_status`: For order creation (draft, pending, confirmed)
- `order_update_status`: For order updates (includes all statuses)

### Marketing Related
- `marketing_channel`: Channel types (sms, email, push_notification, wechat, call)
- `marketing_campaign_status`: Campaign statuses (draft, scheduled, active, etc.)
- `marketing_record_status`: Record statuses (pending, sent, delivered, etc.)
- `marketing_execution_type`: Execution types (actual, simulation)

### Wallet Related
- `wallet_transaction_type`: Transaction types (recharge, consume, refund)

### Analytics Related
- `date_range`: Date ranges (today, week, month, quarter, year)
- `granularity`: Granularity levels (day, week, month)

### Product Related
- `product_type`: Product types (product, service)

## Using Constants in Code

### In Services and Business Logic
```go
import "crm_lite/internal/constants"

// Type-safe constant usage
func CreateOrder(status string) error {
    if status == string(constants.OrderStatusDraft) {
        // Handle draft order
    }
    return nil
}

// Get all valid values
validStatuses := constants.ValidOrderCreateStatuses()
```

### In Domain Models
```go
// Domain models now use the same constants
func (c *CustomerDomain) validateGender() error {
    validGenders := constants.ValidCustomerGenders()
    // validation logic
}
```

## Error Handling

The new system provides better error messages:

```go
// Use the helper function for consistent error messages
import "crm_lite/internal/validators"

func handleValidationError(err error) {
    if fieldError, ok := err.(validator.FieldError); ok {
        message := validators.GetValidationErrorMessage(fieldError)
        // Returns: "Type must be one of: sms, email, push_notification, wechat, call"
    }
}
```

## Testing

The custom validators are thoroughly tested. See `internal/validators/business_validators_test.go` for examples.

### Test Valid Values
```go
func TestValidation(t *testing.T) {
    // Test with valid constant
    body := `{"type": "` + string(constants.MarketingChannelSMS) + `"}`
    // ... test logic
}
```

## Benefits

1. **Refactoring Safety**: Changing a constant value updates all references
2. **IDE Support**: Better autocomplete and navigation
3. **Documentation**: Constants serve as documentation for valid values
4. **Testing**: Easier to write comprehensive tests
5. **API Documentation**: Swagger/OpenAPI can reference the same constants

## Breaking Changes

This is **not** a breaking change for API consumers. The HTTP API still accepts the same string values. Only the internal validation mechanism has changed.

## Rollback Plan

If needed, you can easily rollback by:
1. Reverting DTO binding tags to use `oneof=...`
2. Removing the custom validator registration in bootstrap
3. The constants can remain as they provide value for business logic

## Future Enhancements

Consider these future improvements:
1. Generate OpenAPI documentation from constants
2. Add validation for enum transitions (e.g., order status workflow)
3. Add localization support for error messages
4. Create validation helpers for frontend applications