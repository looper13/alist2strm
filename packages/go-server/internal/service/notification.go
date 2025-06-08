package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

var (
	notificationService *NotificationService
	notificationOnce    sync.Once
)

// GetNotificationService 获取 NotificationService 单例
func GetNotificationService() *NotificationService {
	notificationOnce.Do(func() {
		notificationService = &NotificationService{
			db:     model.DB,
			logger: utils.Logger,
		}
	})
	return notificationService
}

// AddToQueue 添加通知到队列
func (s *NotificationService) AddToQueue(req *model.NotificationQueueCreateRequest) (*model.NotificationQueue, error) {
	notification := &model.NotificationQueue{
		Type:       req.Type,
		Event:      req.Event,
		Payload:    req.Payload,
		Priority:   req.Priority,
		MaxRetries: req.MaxRetries,
		Status:     model.NotificationQueueStatusPending,
	}

	// 设置默认值
	if notification.Priority == 0 {
		notification.Priority = 5
	}
	if notification.MaxRetries == 0 {
		notification.MaxRetries = 3
	}

	if err := s.db.Create(notification).Error; err != nil {
		utils.Error("添加通知到队列失败", zap.Error(err))
		return nil, err
	}

	utils.Info("添加通知到队列成功",
		zap.String("type", string(notification.Type)),
		zap.String("event", string(notification.Event)))
	return notification, nil
}

// GetPendingNotifications 获取待处理的通知
func (s *NotificationService) GetPendingNotifications(limit int) ([]*model.NotificationQueue, error) {
	var notifications []*model.NotificationQueue

	err := s.db.Where("status = ? OR (status = ? AND next_retry_at <= ?)",
		model.NotificationQueueStatusPending,
		model.NotificationQueueStatusFailed,
		time.Now()).
		Order("priority ASC, created_at ASC").
		Limit(limit).
		Find(&notifications).Error

	if err != nil {
		utils.Error("获取待处理通知失败", zap.Error(err))
		return nil, err
	}

	return notifications, nil
}

// ProcessNotification 处理通知
func (s *NotificationService) ProcessNotification(id uint) error {
	// 标记为处理中
	if err := s.UpdateStatus(id, model.NotificationQueueStatusProcessing, ""); err != nil {
		return err
	}

	notification, err := s.GetByID(id)
	if err != nil {
		return err
	}

	var success bool
	var errorMsg string

	// 根据通知类型处理
	switch notification.Type {
	case model.NotificationTypeEmby:
		success, errorMsg = s.processEmbyNotification(notification)
	case model.NotificationTypeTelegram:
		success, errorMsg = s.processTelegramNotification(notification)
	default:
		errorMsg = fmt.Sprintf("不支持的通知类型: %s", notification.Type)
	}

	// 更新处理结果
	if success {
		return s.MarkAsCompleted(id)
	} else {
		return s.MarkAsFailed(id, errorMsg)
	}
}

// processEmbyNotification 处理 Emby 通知
func (s *NotificationService) processEmbyNotification(notification *model.NotificationQueue) (bool, string) {
	// 获取 Emby 配置
	embyService := GetEmbyNotificationService()

	// 解析payload
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
		return false, fmt.Sprintf("解析通知内容失败: %v", err)
	}

	// 根据事件类型发送不同的通知
	switch notification.Event {
	case model.NotificationEventTaskCompleted:
		return embyService.SendTaskCompletedNotification(payload)
	case model.NotificationEventTaskFailed:
		return embyService.SendTaskFailedNotification(payload)
	case model.NotificationEventFileInvalid:
		return embyService.SendFileInvalidNotification(payload)
	default:
		return false, fmt.Sprintf("不支持的事件类型: %s", notification.Event)
	}
}

// processTelegramNotification 处理 Telegram 通知
func (s *NotificationService) processTelegramNotification(notification *model.NotificationQueue) (bool, string) {
	// 获取 Telegram 配置
	telegramService := GetTelegramNotificationService()

	// 解析payload
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
		return false, fmt.Sprintf("解析通知内容失败: %v", err)
	}

	// 根据事件类型发送不同的通知
	switch notification.Event {
	case model.NotificationEventTaskCompleted:
		return telegramService.SendTaskCompletedNotification(payload)
	case model.NotificationEventTaskFailed:
		return telegramService.SendTaskFailedNotification(payload)
	case model.NotificationEventFileInvalid:
		return telegramService.SendFileInvalidNotification(payload)
	default:
		return false, fmt.Sprintf("不支持的事件类型: %s", notification.Event)
	}
}

// GetByID 根据ID获取通知
func (s *NotificationService) GetByID(id uint) (*model.NotificationQueue, error) {
	var notification model.NotificationQueue
	if err := s.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

// UpdateStatus 更新通知状态
func (s *NotificationService) UpdateStatus(id uint, status model.NotificationQueueStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	if status == model.NotificationQueueStatusCompleted {
		now := time.Now()
		updates["processed_at"] = &now
	}

	return s.db.Model(&model.NotificationQueue{}).Where("id = ?", id).Updates(updates).Error
}

// MarkAsCompleted 标记为已完成
func (s *NotificationService) MarkAsCompleted(id uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       model.NotificationQueueStatusCompleted,
		"processed_at": &now,
	}

	if err := s.db.Model(&model.NotificationQueue{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.Error("标记通知为已完成失败", zap.Error(err))
		return err
	}

	utils.Info("通知处理完成", zap.Uint("id", id))
	return nil
}

// MarkAsFailed 标记为失败
func (s *NotificationService) MarkAsFailed(id uint, errorMsg string) error {
	notification, err := s.GetByID(id)
	if err != nil {
		return err
	}

	newRetryCount := notification.RetryCount + 1
	updates := map[string]interface{}{
		"retry_count":   newRetryCount,
		"error_message": errorMsg,
	}

	if newRetryCount >= notification.MaxRetries {
		// 超过最大重试次数，标记为最终失败
		updates["status"] = model.NotificationQueueStatusFailed
		utils.Warn("通知重试次数已达上限",
			zap.Uint("id", id),
			zap.Int("retryCount", newRetryCount),
			zap.String("error", errorMsg))
	} else {
		// 设置下次重试时间
		nextRetry := time.Now().Add(time.Minute * 5) // 5分钟后重试
		updates["next_retry_at"] = &nextRetry
		updates["status"] = model.NotificationQueueStatusPending

		utils.Info("通知处理失败，等待重试",
			zap.Uint("id", id),
			zap.Int("retryCount", newRetryCount),
			zap.Time("nextRetry", nextRetry))
	}

	return s.db.Model(&model.NotificationQueue{}).Where("id = ?", id).Updates(updates).Error
}

// List 获取通知列表
func (s *NotificationService) List(req *model.NotificationQueueQueryRequest) ([]*model.NotificationQueue, int64, error) {
	query := s.db.Model(&model.NotificationQueue{})

	// 构建查询条件
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Event != nil {
		query = query.Where("event = ?", *req.Event)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序和分页
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	var notifications []*model.NotificationQueue
	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// GetStatistics 获取通知统计
func (s *NotificationService) GetStatistics() (*model.NotificationStatistics, error) {
	stats := &model.NotificationStatistics{}

	// 总通知数
	s.db.Model(&model.NotificationQueue{}).Count(&stats.TotalNotifications)

	// 按状态统计
	s.db.Model(&model.NotificationQueue{}).Where("status = ?", model.NotificationQueueStatusPending).Count(&stats.PendingNotifications)
	s.db.Model(&model.NotificationQueue{}).Where("status = ?", model.NotificationQueueStatusCompleted).Count(&stats.CompletedNotifications)
	s.db.Model(&model.NotificationQueue{}).Where("status = ?", model.NotificationQueueStatusFailed).Count(&stats.FailedNotifications)

	// 按类型统计
	s.db.Model(&model.NotificationQueue{}).Where("type = ?", model.NotificationTypeEmby).Count(&stats.EmbyNotifications)
	s.db.Model(&model.NotificationQueue{}).Where("type = ?", model.NotificationTypeTelegram).Count(&stats.TelegramNotifications)

	// 平均重试次数
	var avgRetry float64
	s.db.Model(&model.NotificationQueue{}).Select("AVG(retry_count)").Scan(&avgRetry)
	stats.AverageRetryCount = avgRetry

	return stats, nil
}

// CleanupOldNotifications 清理旧通知
func (s *NotificationService) CleanupOldNotifications(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	result := s.db.Where("created_at < ? AND status = ?", cutoffDate, model.NotificationQueueStatusCompleted).Delete(&model.NotificationQueue{})
	if result.Error != nil {
		utils.Error("清理旧通知记录失败", zap.Error(result.Error))
		return result.Error
	}

	utils.Info("清理旧通知记录成功", zap.Int64("count", result.RowsAffected))
	return nil
}

// ProcessPendingNotifications 批量处理待处理通知
func (s *NotificationService) ProcessPendingNotifications() error {
	notifications, err := s.GetPendingNotifications(50) // 每次处理50个
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := s.ProcessNotification(notification.ID); err != nil {
			utils.Error("处理通知失败",
				zap.Uint("id", notification.ID),
				zap.Error(err))
		}
	}

	if len(notifications) > 0 {
		utils.Info("批量处理通知完成", zap.Int("count", len(notifications)))
	}

	return nil
}
