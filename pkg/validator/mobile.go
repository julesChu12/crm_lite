package validator

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 中国大陆手机号正则表达式
// 支持格式：13800138001 或 +8613800138001
var mobileRegex = regexp.MustCompile(`^(\+86)?1[3-9]\d{9}$`)

// MobileValidator 验证中国大陆手机号
func MobileValidator(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // 空值由 required 标签处理
	}
	return mobileRegex.MatchString(phone)
}

// RegisterMobileValidator 注册手机号验证器到 gin 的验证器引擎
func RegisterMobileValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("mobile", MobileValidator)
	}
}
