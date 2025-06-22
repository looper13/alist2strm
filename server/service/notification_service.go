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
	notifyChan      chan struct{} // 通知信号通道，用于事件驱动处理机制
	retryTimer      *time.Timer   // 重试定时器
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

	// 使用原子操作更新设置
	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()

	// 如果通知功能被禁用，则不初始化渠道
	if !settings.Enabled {
		s.logger.Info("通知功能已禁用")
		return nil
	}

	// 创建通道实例（在加锁外）
	channels := make(map[notification.NotificationChannelType]notification_channel.Channel)

	// 初始化 Telegram 渠道
	telegramChannel := notification_channel.NewTelegramChannel(s.logger, settings)
	if telegramChannel.IsEnabled() {
		channels[telegramChannel.GetType()] = telegramChannel
		s.logger.Info("Telegram 通知渠道已启用")
	}

	// 初始化企业微信渠道
	weworkChannel := notification_channel.NewWeworkChannel(s.logger, settings)
	if weworkChannel.IsEnabled() {
		channels[weworkChannel.GetType()] = weworkChannel
		s.logger.Info("企业微信通知渠道已启用")
	}

	// 原子更新通道
	s.mu.Lock()
	s.channels = channels
	s.mu.Unlock()

	s.logger.Info("通知渠道初始化完成", zap.Int("总渠道数", len(channels)))

	// 启动队列处理
	s.startQueueProcessor()

	// 注册配置更新监听器
	GetConfigListenerService().Register("NOTIFICATION_SETTINGS", s)

	return nil
}

// Reload 重新加载通知设置
func (s *NotificationService) Reload() error {
	s.logger.Info("重新加载通知设置")

	// 提前加载设置，避免在持有锁时进行IO操作
	settings, err := repository.Notification.GetSettings()
	if err != nil {
		s.logger.Error("加载通知设置失败", zap.Error(err))
		return err
	}

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

	// 为了减少锁持有时间，将所有可能的准备工作放在获取锁之前
	var newChannels map[notification.NotificationChannelType]notification_channel.Channel
	if settings.Enabled {
		// 预先创建新的通道实例（无需锁）
		newChannels = make(map[notification.NotificationChannelType]notification_channel.Channel)

		// 初始化 Telegram 渠道
		telegramChannel := notification_channel.NewTelegramChannel(s.logger, settings)
		if telegramChannel.IsEnabled() {
			newChannels[telegramChannel.GetType()] = telegramChannel
		}

		// 初始化企业微信渠道
		weworkChannel := notification_channel.NewWeworkChannel(s.logger, settings)
		if weworkChannel.IsEnabled() {
			newChannels[weworkChannel.GetType()] = weworkChannel
		}
	}

	// 使用最短时间持有写锁更新设置和通道
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		// 更新设置
		s.settings = settings

		// 如果通知功能被禁用，清空通道
		if !settings.Enabled {
			s.channels = make(map[notification.NotificationChannelType]notification_channel.Channel)
			s.logger.Info("通知功能已禁用")
			return
		}

		// 更新通道
		s.channels = newChannels
		s.logger.Info("通知渠道初始化完成", zap.Int("总渠道数", len(s.channels)))

		// 在锁内重置处理状态
		if wasProcessing {
			s.queueProcessing = false
		}
	}()

	// 如果之前是启动状态，在锁外重新启动处理器
	if wasProcessing && settings.Enabled {
		s.startQueueProcessor()
	}

	return nil
}

// SendTaskNotification 发送任务通知
func (s *NotificationService) SendTaskNotification(taskInfo *task.Task, taskLogID uint, status string, duration int64, stats map[string]interface{}) error {
	// 准备通知数据（无需锁，只使用本地变量）
	data := &notification.TaskNotificationData{
		TaskID:     taskInfo.ID,
		TaskName:   taskInfo.Name,
		Status:     status,
		Duration:   duration,
		EventTime:  time.Now(),
		SourcePath: taskInfo.SourcePath,
		TargetPath: taskInfo.TargetPath,
	}

	// 提取基础统计数据 (与 task_log.go 中 TaskLog 字段保持一致)
	if totalFile, ok := stats["total_file"].(int); ok {
		data.TotalFile = totalFile
	}
	if generatedFile, ok := stats["generated_file"].(int); ok {
		data.GeneratedFile = generatedFile
	}
	if skipFile, ok := stats["skip_file"].(int); ok {
		data.SkipFile = skipFile
	}
	if overwriteFile, ok := stats["overwrite_file"].(int); ok {
		data.OverwriteFile = overwriteFile
	}
	if metadataCount, ok := stats["metadata_count"].(int); ok {
		data.MetadataCount = metadataCount
	}
	if subtitleCount, ok := stats["subtitle_count"].(int); ok {
		data.SubtitleCount = subtitleCount
	}

	// 提取细分统计数据
	// 下载统计
	if metadataDownloaded, ok := stats["metadata_downloaded"].(int); ok {
		data.MetadataDownloaded = metadataDownloaded
	}
	if subtitleDownloaded, ok := stats["subtitle_downloaded"].(int); ok {
		data.SubtitleDownloaded = subtitleDownloaded
	}

	// 跳过统计（额外的详细信息）
	if metadataSkipped, ok := stats["metadata_skipped"].(int); ok {
		data.MetadataSkipped = metadataSkipped
	}
	if subtitleSkipped, ok := stats["subtitle_skipped"].(int); ok {
		data.SubtitleSkipped = subtitleSkipped
	}
	if otherSkipped, ok := stats["other_skipped"].(int); ok {
		data.OtherSkipped = otherSkipped
	}

	// 失败统计
	if failedCount, ok := stats["failed_count"].(int); ok {
		data.FailedCount = failedCount
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

	// 序列化通知数据（无需锁）
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("序列化通知数据失败", zap.Error(err))
		return err
	}

	// 短暂加锁检查通知功能是否启用并获取必要的配置信息
	var isEnabled bool
	var channelType string
	var availableChannels map[notification.NotificationChannelType]bool

	func() {
		s.mu.RLock()
		defer s.mu.RUnlock()

		if s.settings == nil || !s.settings.Enabled || len(s.channels) == 0 {
			isEnabled = false
			return
		}

		isEnabled = true
		channelType = s.settings.DefaultChannel

		// 构建可用通道列表（避免在锁外访问 map）
		availableChannels = make(map[notification.NotificationChannelType]bool)
		for t := range s.channels {
			availableChannels[t] = true
		}
	}()

	// 如果通知功能未启用，直接返回
	if !isEnabled {
		s.logger.Debug("通知功能已禁用，跳过发送通知")
		return nil
	}

	// 配置默认渠道（不再需要锁）
	if channelType == "" {
		channelType = string(notification.ChannelTypeTelegram)
	}

	// 检查默认渠道是否可用（不再需要锁，使用前面复制的可用通道信息）
	if !availableChannels[notification.NotificationChannelType(channelType)] {
		// 如果默认渠道不可用，尝试使用任何可用的渠道
		channelType = ""
		for t := range availableChannels {
			channelType = string(t)
			break
		}
	}

	// 如果没有可用的渠道，返回错误
	if channelType == "" {
		s.logger.Warn("没有可用的通知渠道")
		return fmt.Errorf("没有可用的通知渠道")
	}

	// 添加到队列（数据库操作，不需要锁）
	err = repository.Notification.AddToQueue(channelType, string(templateType), string(jsonData))
	if err != nil {
		s.logger.Error("将通知添加到队列失败", zap.Error(err))
		return err
	}

	s.logger.Info("通知已加入队列",
		zap.String("channelType", channelType),
		zap.String("templateType", string(templateType)),
		zap.String("taskName", taskInfo.Name))

	// 发送处理信号，触发即时处理（无需锁）
	// 使用非阻塞发送确保不会因通道问题而阻塞整个通知流程
	select {
	case s.notifyChan <- struct{}{}:
		// 信号发送成功
	default:
		// 通道已满，记录日志但不阻塞
		s.logger.Warn("通知信号通道已满，将依靠定期检查机制处理")
	}

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
	s.notifyChan = make(chan struct{}, 100) // 通知信号通道，设置合理的缓冲区大小
	s.queueProcessing = true
	s.mu.Unlock()

	go func() {
		defer func() {
			// 安全地修改队列处理状态
			s.mu.Lock()
			s.queueProcessing = false
			// 关闭并清空通知通道
			if s.retryTimer != nil {
				s.retryTimer.Stop()
				s.retryTimer = nil
			}
			s.mu.Unlock()

			// 关闭完成通道
			close(s.doneChan)
		}()

		// 启动时先处理一次队列，确保处理未完成的历史消息
		s.processQueue()

		// 低频率备份检查，以防有漏网之鱼
		backupTicker := time.NewTicker(5 * time.Minute)
		defer backupTicker.Stop()

		for {
			select {
			case <-s.notifyChan:
				// 收到通知信号，立即处理队列
				s.processQueue()
			case <-backupTicker.C:
				// 定期检查是否有遗漏的消息
				hasPending, _ := repository.Notification.HasPendingNotifications()
				if hasPending {
					s.processQueue()
				}
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
	// 在处理队列前，停止可能正在运行的重试定时器
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.retryTimer != nil {
			s.retryTimer.Stop()
			s.retryTimer = nil
		}
	}()

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
		// 没有待处理通知，但检查是否有需要稍后重试的消息
		s.scheduleNextRetry()
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

	// 处理完一批后，检查是否有下一批待处理的消息
	hasPending, err := repository.Notification.HasPendingNotifications()
	if err == nil && hasPending {
		// 如果还有消息，立即发送信号，继续处理
		select {
		case s.notifyChan <- struct{}{}:
			// 信号发送成功
		default:
			// 通道已满，记录日志但不阻塞
			s.logger.Warn("通知信号通道已满，无法触发下一批处理")
		}
	} else {
		// 没有待处理消息，检查是否有需要稍后重试的消息
		s.scheduleNextRetry()
	}
}

// scheduleNextRetry 安排下一次重试
func (s *NotificationService) scheduleNextRetry() {
	nextRetryTime, hasRetry := repository.Notification.GetEarliestRetryTime()
	if !hasRetry {
		return
	}

	// 计算需要等待的时间
	waitDuration := time.Until(nextRetryTime)
	if waitDuration < 0 {
		// 如果已经到了或过了重试时间，立即触发
		waitDuration = 0
	}

	// 锁保护定时器创建
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经存在定时器，先停止它
	if s.retryTimer != nil {
		s.retryTimer.Stop()
	}

	// 创建新的定时器
	s.retryTimer = time.AfterFunc(waitDuration, func() {
		select {
		case s.notifyChan <- struct{}{}:
			// 信号发送成功
		default:
			// 通道已满，记录日志但不阻塞
			s.logger.Warn("通知信号通道已满，将在下一轮检查中重试")
		}
	})

	if waitDuration > 0 {
		s.logger.Debug("安排下一次消息重试",
			zap.Time("重试时间", nextRetryTime),
			zap.Duration("等待时间", waitDuration))
	} else {
		s.logger.Debug("即将触发消息重试")
	}
}
