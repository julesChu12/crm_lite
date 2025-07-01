package resource

import (
	"context"
	"crm_lite/internal/core/config"
)

// EmailServiceKey 定义了邮件服务在资源管理器中的键
const EmailServiceKey = "email"

// EmailResource 仅封装邮件服务的配置
type EmailResource struct {
	Opts config.EmailOptions
}

// NewEmailResource 创建一个新的邮件资源实例
func NewEmailResource(opts config.EmailOptions) *EmailResource {
	return &EmailResource{Opts: opts}
}

// Initialize 实现了Resource接口
func (r *EmailResource) Initialize(ctx context.Context) error {
	// 仅验证配置，不在此处创建服务实例
	return nil
}

// Close 实现了Resource接口
func (r *EmailResource) Close(ctx context.Context) error {
	return nil
}
