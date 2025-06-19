package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/MccRay-s/alist2strm/model/notification"
	"github.com/MccRay-s/alist2strm/model/task"
	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/service/notification_channel"
	"github.com/MccRay-s/alist2strm/utils"
	"go.uber.org/zap"
)

// NotificationService 通知服务
type NotificationService struct {
	logger          *zap.Logger
	mu              sync.RWMutex
	settings        *notification.Settings
	channels        map[notification.NotificationChannelType]notification_channel.Channel
	queueProcessing bool
	stopChan        chan struct{}
	doneChan        chan struct{}
}

// OnConfigUpdate 实现配置更新监听器接口
func (s *NotificationService) OnConfigUpdate(code string) error {
	if code == "NOTIFICATION_SETTINGS" {
		s.logger.Info("检测到通知配置更新，重新加载配置")
		return s.Reload()
	}
	return nil
}

var (
	notificationInstance *NotificationService
	notificationOnce     sync.Once
)

// GetNotificationService 获取通知服务实例
func GetNotificationService() *NotificationService {
	notificationOnce.Do(func() {
		logger := utils.InfoLogger.Desugar()
		notificationInstance = &NotificationService{
			logger:   logger,
			channels: make(map[notification.NotificationChannelType]notification_channel.Channel),
		}
		notificationInstance.Initialize()
	})
	return notificationInstance
}

// Initialize 初始化通知服务
func (s *NotificationService) Initialize() error {
	s.logger.Info("初始化通知服务")

	// 加载设置
	settings, err := repository.Notification.GetSettings()
	if err != nil {
		s.logger.Error("加载通知设置失败", zap.Error(err))
		return err
	}
	s.settings = settings

	// 如果通知功能被禁用，则不初始化渠道
	if !settings.Enabled {
		s.logger.Info("通知功能已禁用")
		return nil
	}

	// 初始化通知渠道
	s.initChannels()

	// 启动队列处理
	s.startQueueProcessor()

	// 注册配置更新监听器
	GetConfigListenerService().Register("NOTIFICATION_SETTINGS", s)

	return nil
}

// Reload 重新加载通知设置
func (s *NotificationService) Reload() error {
	s.logger.Info("重新加载通知设置")

	// 标记队列处理状态
	var wasProcessing bool

	// 在停止队列处理器前获取状态但不持有完整锁
	func() {
		s.mu.RLock()
		defer s.mu.RUnlock()
		wasProcessing = s.queueProcessing
	}()

	// 停止现有队列处理器前不持有锁，以避免死锁
	if wasProcessing {
		s.stopQueueProcessor()
	}

	// 重新加载设置时获取锁
	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空现有渠道
	s.channels = make(map[notification.NotificationChannelType]notification_channel.Channel)

	// 加载设置
	settings, err := repository.Notification.GetSettings()
	if err != nil {
		s.logger.Error("加载通知设置失败", zap.Error(err))
		return err
	}
	s.settings = settings

	// 如果通知功能被禁用，则不初始化渠道
	if !settings.Enabled {
		s.logger.Info("通知功能已禁用")
		return nil
	}

	// 初始化通知渠道
	s.initChannels()

	// 重新启动队列处理（如果之前是启动状态）
	if wasProcessing {
		// 在锁内部设置状态
		s.queueProcessing = false
		// 在释放锁后启动处理器
		defer s.startQueueProcessor()
	}

	return nil
}

// initChannels 初始化通知渠道
func (s *NotificationService) initChannels() {
	// 初始化 Telegram 渠道
	telegramChannel := notification_channel.NewTelegramChannel(s.logger, s.settings)
	if telegramChannel.IsEnabled() {
		s.channels[telegramChannel.GetType()] = telegramChannel
		s.logger.Info("Telegram 通知渠道已启用")
	}

	// 初始化企业微信渠道
	weworkChannel := notification_channel.NewWeworkChannel(s.logger, s.settings)
	if weworkChannel.IsEnabled() {
		s.channels[weworkChannel.GetType()] = weworkChannel
		s.logger.Info("企业微信通知渠道已启用")
	}

	s.logger.Info("通知渠道初始化完成", zap.Int("总渠道数", len(s.channels)))
}

// SendTaskNotification 发送任务通知
func (s *NotificationService) SendTaskNotification(taskInfo *task.Task, taskLogID uint, status string, duration int64, stats map[string]interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 检查通知功能是否启用
	if s.settings == nil || !s.settings.Enabled {
		s.logger.Debug("通知功能已禁用，跳过发送通知")
		return nil
	}

	// 准备通知数据
	data := &notification.TaskNotificationData{
		TaskID:   taskInfo.ID,
		TaskName: taskInfo.Name,
		Status:   status,
		Duration: duration,
	}

	// 提取统计数据
	if totalFiles, ok := stats["total_file"].(int); ok {
		data.TotalFiles = totalFiles
	}
	if generatedFiles, ok := stats["generated_file"].(int); ok {
		data.GeneratedFiles = generatedFiles
	}
	if skippedFiles, ok := stats["skip_file"].(int); ok {
		data.SkippedFiles = skippedFiles
	}
	if metadataFiles, ok := stats["metadata_count"].(int); ok {
		data.MetadataFiles = metadataFiles
	}
	if subtitleFiles, ok := stats["subtitle_count"].(int); ok {
		data.SubtitleFiles = subtitleFiles
	}

	// 设置错误信息（如果有）
	if status == "failed" && stats["message"] != nil {
		if errMsg, ok := stats["message"].(string); ok {
			data.ErrorMessage = errMsg
		}
	}

	// 选择模板类型
	var templateType notification.TemplateType
	if status == "completed" {
		templateType = notification.TemplateTypeTaskComplete
	} else {
		templateType = notification.TemplateTypeTaskFailed
	}

	// 将通知加入队列
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("序列化通知数据失败", zap.Error(err))
		return err
	}

	// 获取默认通知渠道
	channelType := s.settings.DefaultChannel
	if channelType == "" {
		channelType = string(notification.ChannelTypeTelegram)
	}

	// 检查默认渠道是否可用
	if _, ok := s.channels[notification.NotificationChannelType(channelType)]; !ok {
		// 如果默认渠道不可用，尝试使用任何可用的渠道
		for t := range s.channels {
			channelType = string(t)
			break
		}
	}

	// 如果没有可用的渠道，返回错误
	if channelType == "" || len(s.channels) == 0 {
		s.logger.Warn("没有可用的通知渠道")
		return fmt.Errorf("没有可用的通知渠道")
	}

	// 添加到队列
	err = repository.Notification.AddToQueue(channelType, string(templateType), string(jsonData))
	if err != nil {
		s.logger.Error("将通知添加到队列失败", zap.Error(err))
		return err
	}

	s.logger.Info("通知已加入队列",
		zap.String("channelType", channelType),
		zap.String("templateType", string(templateType)),
		zap.String("taskName", taskInfo.Name))

	return nil
}

// startQueueProcessor 启动队列处理器
func (s *NotificationService) startQueueProcessor() {
	// 使用互斥锁保护队列处理状态和通道创建
	s.mu.Lock()
	if s.queueProcessing {
		s.mu.Unlock()
		return
	}

	s.stopChan = make(chan struct{})
	s.doneChan = make(chan struct{})
	s.queueProcessing = true
	s.mu.Unlock()

	go func() {
		defer func() {
			// 安全地修改队列处理状态
			s.mu.Lock()
			s.queueProcessing = false
			s.mu.Unlock()

			// 关闭完成通道
			close(s.doneChan)
		}()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processQueue()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// stopQueueProcessor 停止队列处理器
func (s *NotificationService) stopQueueProcessor() {
	// 检查处理状态，并安全地获取停止通道
	var shouldStop bool
	var stopCh, doneCh chan struct{}

	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if !s.queueProcessing {
			return
		}

		shouldStop = true
		stopCh = s.stopChan
		doneCh = s.doneChan
	}()

	// 如果没有正在运行的处理器，直接返回
	if !shouldStop {
		return
	}

	// 关闭停止通道，通知处理器退出
	close(stopCh)

	// 添加超时机制，避免可能的死锁
	select {
	case <-doneCh:
		// 正常接收到关闭信号
	case <-time.After(5 * time.Second):
		s.logger.Warn("等待队列处理器停止超时")

		// 在超时时手动标记状态
		s.mu.Lock()
		s.queueProcessing = false
		s.mu.Unlock()
	}
}

// processQueue 处理通知队列
func (s *NotificationService) processQueue() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.settings == nil || !s.settings.Enabled || len(s.channels) == 0 {
		return
	}

	// 获取队列设置
	concurrency := s.settings.QueueSettings.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	maxRetries := s.settings.QueueSettings.MaxRetries
	retryInterval := s.settings.QueueSettings.RetryInterval

	// 获取待处理通知
	notifications, err := repository.Notification.GetPendingNotifications(concurrency)
	if err != nil {
		s.logger.Error("获取待处理通知失败", zap.Error(err))
		return
	}

	if len(notifications) == 0 {
		return
	}

	s.logger.Debug("开始处理通知队列", zap.Int("数量", len(notifications)))

	for _, notif := range notifications {
		// 更新状态为处理中
		err := repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusProcessing, "")
		if err != nil {
			s.logger.Error("更新通知状态失败", zap.Error(err), zap.Uint("id", notif.ID))
			continue
		}

		// 获取渠道
		channelType := notification.NotificationChannelType(notif.ChannelType)
		channel, ok := s.channels[channelType]
		if !ok {
			errMsg := fmt.Sprintf("通知渠道不可用: %s", notif.ChannelType)
			s.logger.Warn(errMsg, zap.Uint("id", notif.ID))
			repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusFailed, errMsg)
			continue
		}

		// 解析数据
		var data notification.TaskNotificationData
		err = json.Unmarshal([]byte(notif.Payload), &data)
		if err != nil {
			errMsg := fmt.Sprintf("解析通知数据失败: %v", err)
			s.logger.Error(errMsg, zap.Uint("id", notif.ID))
			repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusFailed, errMsg)
			continue
		}

		// 发送通知
		err = channel.Send(notification.TemplateType(notif.TemplateType), &data)
		if err != nil {
			errMsg := fmt.Sprintf("发送通知失败: %v", err)
			s.logger.Error(errMsg,
				zap.Uint("id", notif.ID),
				zap.String("channelType", notif.ChannelType),
				zap.String("taskName", data.TaskName))

			// 检查是否应该重试
			if notif.RetryCount < maxRetries {
				nextRetry := time.Now().Add(time.Duration(retryInterval) * time.Second)
				repository.Notification.UpdateNotificationForRetry(notif.ID, notif.RetryCount+1, nextRetry, errMsg)
				s.logger.Info("通知将在稍后重试",
					zap.Uint("id", notif.ID),
					zap.Int("retryCount", notif.RetryCount+1),
					zap.Time("nextRetryTime", nextRetry))
			} else {
				repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusFailed, errMsg)
				s.logger.Warn("通知达到最大重试次数，已标记为失败", zap.Uint("id", notif.ID))
			}
			continue
		}

		// 更新为已发送状态
		repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusSent, "")
		s.logger.Info("通知已成功发送",
			zap.Uint("id", notif.ID),
			zap.String("channelType", notif.ChannelType),
			zap.String("taskName", data.TaskName))
	}

	// 清理较早的已发送通知（30天前）
	cleanCutoff := time.Now().AddDate(0, 0, -30)
	repository.Notification.CleanSentNotifications(cleanCutoff)
}
