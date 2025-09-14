package impl

import (
	"crm_lite/internal/common"
	"crm_lite/internal/domains/notification"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewNotificationService 创建Notification服务实例
func NewNotificationService(
	db *gorm.DB,
	emailConfig notification.EmailConfig,
	smsConfig notification.SMSConfig,
	logger *zap.Logger,
) notification.Service {
	tx := common.NewTx(db)
	return NewNotificationServiceImpl(db, tx, emailConfig, smsConfig, logger)
}

// NewNotificationServiceForController 为控制器创建Notification服务实例
// 支持Legacy兼容接口，便于现有控制器调用
func NewNotificationServiceForController(
	db *gorm.DB,
	emailConfig notification.EmailConfig,
	smsConfig notification.SMSConfig,
	logger *zap.Logger,
) *NotificationServiceImpl {
	tx := common.NewTx(db)
	return NewNotificationServiceImpl(db, tx, emailConfig, smsConfig, logger)
}
