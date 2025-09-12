package validator

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func TestMobileValidator(t *testing.T) {
	// 注册验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("mobile", MobileValidator)
	}

	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"空字符串", "", true},                          // 空值应该通过（由 required 处理）
		{"中国大陆手机号-13", "13800138001", true},         // 标准格式
		{"中国大陆手机号-15", "15912345678", true},         // 标准格式
		{"中国大陆手机号-18", "18612345678", true},         // 标准格式
		{"中国大陆手机号-19", "19812345678", true},         // 标准格式
		{"带国际区号", "+8613800138001", true},            // 国际格式
		{"无效号码-12开头", "12800138001", false},          // 12开头无效
		{"无效号码-长度不够", "1380013800", false},           // 长度不够
		{"无效号码-长度过长", "138001380011", false},         // 长度过长
		{"无效号码-非数字", "13800a38001", false},           // 包含非数字字符
		{"无效国际区号", "+8612800138001", false},          // 错误的国际区号+12开头
		{"固定电话", "02012345678", false},               // 固定电话格式
	}

	validate := validator.New()
	validate.RegisterValidation("mobile", MobileValidator)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Phone string `validate:"mobile"`
			}
			
			testData := TestStruct{Phone: tt.phone}
			err := validate.Struct(testData)
			
			if tt.expected {
				if err != nil {
					t.Errorf("Expected phone %s to be valid, but got error: %v", tt.phone, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected phone %s to be invalid, but validation passed", tt.phone)
				}
			}
		})
	}
}
