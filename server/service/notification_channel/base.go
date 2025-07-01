package notification_channel

import (
	"github.com/MccRay-s/alist2strm/model/notification"
	"go.uber.org/zap"
)

// Channel 通知渠道接口
type Channel interface {
	// Send 发送通知
	Send(templateType notification.TemplateType, data interface{}) error
	// IsEnabled 检查是否启用
	IsEnabled() bool
	// GetType 获取渠道类型
	GetType() notification.NotificationChannelType
}

// BaseChannel 基础通知渠道
type BaseChannel struct {
	logger   *zap.Logger
	enabled  bool
	settings *notification.Settings
}

// IsEnabled 检查是否启用
func (c *BaseChannel) IsEnabled() bool {
	return c.enabled
}

// NewBaseChannel 创建基础通知渠道
func NewBaseChannel(logger *zap.Logger, settings *notification.Settings) *BaseChannel {
	return &BaseChannel{
		logger:   logger,
		settings: settings,
	}
}
