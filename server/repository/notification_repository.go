package repository

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/model/configs"
	"github.com/MccRay-s/alist2strm/model/notification"
	"github.com/MccRay-s/alist2strm/utils"
	"gorm.io/gorm"
)

// NotificationRepository 通知仓库
type NotificationRepository struct{}

// Notification 通知仓库实例
var Notification = &NotificationRepository{}

// GetSettings 获取通知设置
func (r *NotificationRepository) GetSettings() (*notification.Settings, error) {
	config, err := Config.GetByCode("NOTIFICATION_SETTINGS")
	if err != nil {
		utils.Error("获取通知设置失败", "error", err.Error())
		return nil, err
	}

	// 处理配置不存在的情况
	if config == nil {
		// 配置不存在，创建默认配置
		defaultSettings := notification.DefaultSettings()
		jsonBytes, err := json.Marshal(defaultSettings)
		if err != nil {
			utils.Error("序列化默认通知设置失败", "error", err.Error())
			return nil, err
		}

		newConfig := &configs.Config{
			Name:  "通知系统配置",
			Code:  "NOTIFICATION_SETTINGS",
			Value: string(jsonBytes),
		}

		if err := Config.Create(newConfig); err != nil {
			utils.Error("创建默认通知设置失败", "error", err.Error())
			return nil, err
		}

		return defaultSettings, nil
	}

	// 解析配置内容
	var settings notification.Settings
	if err := json.Unmarshal([]byte(config.Value), &settings); err != nil {
		utils.Error("解析通知设置失败", "error", err.Error())
		return nil, err
	}

	return &settings, nil
}

// UpdateSettings 更新通知设置
func (r *NotificationRepository) UpdateSettings(settings *notification.Settings) error {
	jsonBytes, err := json.Marshal(settings)
	if err != nil {
		utils.Error("序列化通知设置失败", "error", err.Error())
		return err
	}

	config, err := Config.GetByCode("NOTIFICATION_SETTINGS")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 配置不存在，创建新配置
			newConfig := &configs.Config{
				Name:  "通知系统配置",
				Code:  "NOTIFICATION_SETTINGS",
				Value: string(jsonBytes),
			}
			return Config.Create(newConfig)
		}
		return err
	}

	// 更新现有配置
	config.Value = string(jsonBytes)
	return Config.Update(config)
}

// AddToQueue 添加通知到队列
func (r *NotificationRepository) AddToQueue(channelType string, templateType string, payload string) error {
	queue := &notification.Queue{
		ChannelType:  channelType,
		TemplateType: templateType,
		Status:       notification.StatusPending,
		Payload:      payload,
		RetryCount:   0,
	}

	return database.DB.Create(queue).Error
}

// GetPendingNotifications 获取待处理通知
func (r *NotificationRepository) GetPendingNotifications(limit int) ([]*notification.Queue, error) {
	var notifications []*notification.Queue
	now := time.Now()

	err := database.DB.Where("status = ? OR (status = ? AND next_retry_time <= ?)",
		notification.StatusPending, notification.StatusFailed, now).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications).Error

	return notifications, err
}

// UpdateNotificationStatus 更新通知状态
func (r *NotificationRepository) UpdateNotificationStatus(id uint, status notification.Status, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return database.DB.Model(&notification.Queue{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateNotificationForRetry 更新通知为重试状态
func (r *NotificationRepository) UpdateNotificationForRetry(id uint, retryCount int, nextRetryTime time.Time, errorMsg string) error {
	updates := map[string]interface{}{
		"status":          notification.StatusFailed,
		"retry_count":     retryCount,
		"next_retry_time": nextRetryTime,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return database.DB.Model(&notification.Queue{}).Where("id = ?", id).Updates(updates).Error
}

// CleanSentNotifications 清理已发送的通知
func (r *NotificationRepository) CleanSentNotifications(beforeTime time.Time) error {
	return database.DB.Where("status = ? AND updated_at < ?", notification.StatusSent, beforeTime).Delete(&notification.Queue{}).Error
}
