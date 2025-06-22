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

	// 获取状态为待处理的通知，或者状态为失败但重试时间已到的通知
	// 只处理状态为pending或者有next_retry_time且时间已到的通知
	err := database.DB.Where("status = ? OR (status = ? AND next_retry_time IS NOT NULL AND next_retry_time <= ?)",
		notification.StatusPending, notification.StatusFailed, now).
		// 优先按重试时间排序，确保到时间的重试任务被优先处理
		Order("CASE WHEN next_retry_time IS NOT NULL THEN next_retry_time ELSE created_at END ASC").
		Limit(limit).
		Find(&notifications).Error

	// 记录查询到的通知数量和状态
	if err == nil {
		if len(notifications) > 0 {
			utils.InfoLogger.Infof("查询到 %d 条待处理通知", len(notifications))
			for _, n := range notifications {
				retryTimeStr := "无"
				if n.NextRetryTime != nil {
					retryTimeStr = n.NextRetryTime.Format("2006-01-02 15:04:05")
				}
				utils.InfoLogger.Infof("通知ID: %d, 状态: %s, 重试次数: %d, 下次重试时间: %s",
					n.ID, n.Status, n.RetryCount, retryTimeStr)
			}
		} else {
			// 记录无待处理通知的原因，便于调试
			var pendingCount, retryCount int64
			database.DB.Model(&notification.Queue{}).Where("status = ?", notification.StatusPending).Count(&pendingCount)
			database.DB.Model(&notification.Queue{}).Where("status = ? AND next_retry_time IS NOT NULL AND next_retry_time <= ?",
				notification.StatusFailed, now).Count(&retryCount)
			utils.InfoLogger.Infof("当前无待处理通知：待处理状态=%d, 可重试状态=%d", pendingCount, retryCount)
		}
	}

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

	// 当状态变为失败时，清除下一次重试时间，避免已经失败的通知被重试
	if status == notification.StatusFailed {
		updates["next_retry_time"] = nil
	}

	err := database.DB.Model(&notification.Queue{}).Where("id = ?", id).Updates(updates).Error
	if err == nil {
		utils.InfoLogger.Infof("已更新通知状态, ID: %d, 状态: %s", id, status)
	}
	return err
}

// RequeueNotification 重新入队通知（用于重试）
func (r *NotificationRepository) RequeueNotification(id uint, retryCount int, nextRetryTime time.Time, errorMsg string) error {
	updates := map[string]interface{}{
		"status":          notification.StatusPending, // 将状态改回待处理，这样会被下一次队列处理拿到
		"retry_count":     retryCount,
		"next_retry_time": nextRetryTime,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	err := database.DB.Model(&notification.Queue{}).Where("id = ?", id).Updates(updates).Error
	if err == nil {
		utils.InfoLogger.Infof("已将通知重新加入队列，ID: %d, 重试次数: %d, 下次重试时间: %s",
			id, retryCount, nextRetryTime.Format("2006-01-02 15:04:05"))
	} else {
		utils.InfoLogger.Errorf("通知重新入队失败，ID: %d, 错误: %v", id, err)
	}
	return err
}

// UpdateNotificationForRetry 更新通知为重试状态（为保持兼容性，保留此方法）
func (r *NotificationRepository) UpdateNotificationForRetry(id uint, retryCount int, nextRetryTime time.Time, errorMsg string) error {
	// 直接调用新的RequeueNotification方法
	return r.RequeueNotification(id, retryCount, nextRetryTime, errorMsg)
}

// CleanSentNotifications 清理已发送的通知
func (r *NotificationRepository) CleanSentNotifications(beforeTime time.Time) error {
	return database.DB.Where("status = ? AND updated_at < ?", notification.StatusSent, beforeTime).Delete(&notification.Queue{}).Error
}

// GetEarliestRetryTime 获取最早需要重试的消息时间
func (r *NotificationRepository) GetEarliestRetryTime() (time.Time, bool) {
	var nextRetry notification.Queue
	result := database.DB.Where("status = ? AND next_retry_time IS NOT NULL", notification.StatusFailed).
		Order("next_retry_time ASC").
		First(&nextRetry)

	if result.Error != nil {
		return time.Time{}, false
	}

	// 确保 NextRetryTime 不为 nil
	if nextRetry.NextRetryTime == nil {
		return time.Time{}, false
	}

	return *nextRetry.NextRetryTime, true
}

// HasPendingNotifications 检查是否有未处理的通知
func (r *NotificationRepository) HasPendingNotifications() (bool, error) {
	var count int64
	now := time.Now()

	err := database.DB.Model(&notification.Queue{}).
		Where("status = ? OR (status = ? AND next_retry_time IS NOT NULL AND next_retry_time <= ?)",
			notification.StatusPending, notification.StatusFailed, now).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
