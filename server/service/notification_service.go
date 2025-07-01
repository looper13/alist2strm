package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	configRequest "github.com/MccRay-s/alist2strm/model/configs/request"
	"github.com/MccRay-s/alist2strm/model/notification"
	"github.com/MccRay-s/alist2strm/model/task"
	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/service/notification_channel"
	"github.com/MccRay-s/alist2strm/utils"
	"go.uber.org/zap"
)

// NotificationService 通知服务
type NotificationService struct {
	logger   *zap.Logger
	mu       sync.RWMutex
	settings *notification.Settings
	channels map[notification.NotificationChannelType]notification_channel.Channel
	// 内存队列相关
	memoryQueue     chan *notification.Queue
	queueProcessing bool
	stopChan        chan struct{}
	cleanupStopChan chan struct{}
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

	// 1.1 实例化通知服务 (已完成)
	s.mu.Lock()
	// 初始化内存队列，设置合理的缓冲区大小
	s.memoryQueue = make(chan *notification.Queue, 1000)
	s.mu.Unlock()

	// 1.2 加载 NOTIFICATION_SETTINGS 配置
	settings, err := s.loadNotificationSettings()
	if err != nil {
		s.logger.Error("加载通知设置失败", zap.Error(err))
		// 配置加载失败，但不影响服务初始化
		s.logger.Info("通知服务初始化完成（配置加载失败）")
		// 注册配置更新监听器，等待配置修复
		GetConfigListenerService().Register("NOTIFICATION_SETTINGS", s)
		return nil
	}

	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()

	// 如果配置不存在，或通知未开启，亦或没有通知渠道，则不处理配置信息，完成初始化
	if !s.shouldStartNotificationProcessing(settings) {
		s.logger.Info("通知服务初始化完成（通知功能未启用或无可用渠道）")
		// 注册配置更新监听器，等待配置变更
		GetConfigListenerService().Register("NOTIFICATION_SETTINGS", s)
		return nil
	}

	// 1.3 开启通知，且有通知渠道，则加载配置信息
	if err := s.initializeNotificationFeatures(settings); err != nil {
		s.logger.Error("初始化通知功能失败", zap.Error(err))
		// 功能初始化失败，但服务本身初始化成功
		s.logger.Info("通知服务初始化完成（功能初始化失败）")
		// 注册配置更新监听器
		GetConfigListenerService().Register("NOTIFICATION_SETTINGS", s)
		return nil
	}

	// 注册配置更新监听器
	GetConfigListenerService().Register("NOTIFICATION_SETTINGS", s)

	s.logger.Info("通知服务初始化完成（所有功能已启用）")
	return nil
}

// Reload 重新加载通知设置
func (s *NotificationService) Reload() error {
	s.logger.Info("重新加载通知设置")

	// 1.2 加载 NOTIFICATION_SETTINGS 配置
	settings, err := s.loadNotificationSettings()
	if err != nil {
		s.logger.Error("加载通知设置失败", zap.Error(err))
		// 配置加载失败，停止现有功能
		s.stopNotificationFeatures()
		return err
	}

	// 检查当前处理状态
	s.mu.RLock()
	wasProcessing := s.queueProcessing
	s.mu.RUnlock()

	// 判断并执行后续逻辑
	shouldStart := s.shouldStartNotificationProcessing(settings)

	if wasProcessing && !shouldStart {
		// 之前在运行，现在需要停止
		s.logger.Info("配置变更，停止通知功能")
		s.stopNotificationFeatures()
	} else if !wasProcessing && shouldStart {
		// 之前未运行，现在需要启动
		s.logger.Info("配置变更，启动通知功能")

		// 更新设置
		s.mu.Lock()
		s.settings = settings
		s.mu.Unlock()

		// 初始化通知功能
		if err := s.initializeNotificationFeatures(settings); err != nil {
			s.logger.Error("启动通知功能失败", zap.Error(err))
			return err
		}
	} else if wasProcessing && shouldStart {
		// 之前在运行，现在需要重启（配置可能有变更）
		s.logger.Info("配置变更，重启通知功能")

		// 停止现有功能
		s.stopNotificationFeatures()

		// 更新设置
		s.mu.Lock()
		s.settings = settings
		s.mu.Unlock()

		// 重新启动功能
		if err := s.initializeNotificationFeatures(settings); err != nil {
			s.logger.Error("重启通知功能失败", zap.Error(err))
			return err
		}
	} else {
		// 都不运行，只更新设置
		s.mu.Lock()
		s.settings = settings
		s.mu.Unlock()
		s.logger.Info("配置已更新，通知功能保持禁用状态")
	}

	return nil
}

// SendTaskNotification 发送任务通知
func (s *NotificationService) SendTaskNotification(taskInfo *task.Task, taskLogID uint, status string, duration int64, stats map[string]interface{}) error {
	// 检查通知功能是否启用
	s.mu.RLock()
	enabled := s.settings != nil && s.settings.Enabled && len(s.channels) > 0
	defaultChannel := ""
	if s.settings != nil {
		defaultChannel = s.settings.DefaultChannel
	}
	s.mu.RUnlock()

	if !enabled {
		s.logger.Debug("通知功能已禁用，跳过发送通知")
		return nil
	}

	// 准备通知数据
	data := &notification.TaskNotificationData{
		TaskID:     taskInfo.ID,
		TaskName:   taskInfo.Name,
		Status:     status,
		Duration:   duration,
		EventTime:  time.Now().Format("2006-01-02 15:04:05"),
		SourcePath: taskInfo.SourcePath,
		TargetPath: taskInfo.TargetPath,
	}

	// 提取统计数据
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
	if metadataDownloaded, ok := stats["metadata_downloaded"].(int); ok {
		data.MetadataDownloaded = metadataDownloaded
	}
	if subtitleDownloaded, ok := stats["subtitle_downloaded"].(int); ok {
		data.SubtitleDownloaded = subtitleDownloaded
	}
	if metadataSkipped, ok := stats["metadata_skipped"].(int); ok {
		data.MetadataSkipped = metadataSkipped
	}
	if subtitleSkipped, ok := stats["subtitle_skipped"].(int); ok {
		data.SubtitleSkipped = subtitleSkipped
	}
	if otherSkipped, ok := stats["other_skipped"].(int); ok {
		data.OtherSkipped = otherSkipped
	}
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

	// 序列化通知数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("序列化通知数据失败", zap.Error(err))
		return err
	}

	// 配置默认渠道
	channelType := defaultChannel
	if channelType == "" {
		channelType = string(notification.ChannelTypeTelegram)
	}

	// 先保存到数据库，获取ID
	queueID, err := repository.Notification.AddToQueueWithID(channelType, string(templateType), string(jsonData))
	if err != nil {
		s.logger.Error("将通知添加到数据库失败", zap.Error(err))
		return err
	}

	// 创建内存队列项目，包含数据库ID
	queueItem := &notification.Queue{
		ID:           queueID,
		ChannelType:  channelType,
		TemplateType: string(templateType),
		Status:       notification.StatusPending,
		Payload:      string(jsonData),
		RetryCount:   0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 添加到内存队列
	select {
	case s.memoryQueue <- queueItem:
		s.logger.Info("通知已加入内存队列",
			zap.String("channelType", channelType),
			zap.String("templateType", string(templateType)),
			zap.String("taskName", taskInfo.Name))
	default:
		s.logger.Warn("内存队列已满，通知将稍后处理",
			zap.String("taskName", taskInfo.Name))
	}

	return nil
}

// shouldStartNotificationProcessing 判断是否应该启动通知处理
func (s *NotificationService) shouldStartNotificationProcessing(settings *notification.Settings) bool {
	if settings == nil {
		s.logger.Debug("配置为空，不启动通知处理")
		return false
	}

	if !settings.Enabled {
		s.logger.Info("通知功能已禁用")
		return false
	}

	// 检查是否有可用的通知渠道
	hasAvailableChannel := false
	for channelName, channelConfig := range settings.Channels {
		if channelConfig.Enabled {
			hasAvailableChannel = true
			s.logger.Debug("发现可用通知渠道", zap.String("channel", channelName))
			break
		}
	}

	if !hasAvailableChannel {
		s.logger.Info("没有可用的通知渠道")
		return false
	}

	return true
}

// initializeNotificationFeatures 初始化通知功能
func (s *NotificationService) initializeNotificationFeatures(settings *notification.Settings) error {
	s.logger.Info("开始初始化通知功能")

	// 加载配置信息 - 初始化通知渠道
	if err := s.initializeChannels(settings); err != nil {
		return fmt.Errorf("初始化通知渠道失败: %w", err)
	}

	// 加载数据库通知队列，进入内存队列
	if err := s.loadPendingNotificationsToMemory(); err != nil {
		s.logger.Error("加载待处理通知到内存队列失败", zap.Error(err))
		return fmt.Errorf("加载待处理通知到内存队列失败: %w", err)
	}

	// 启动通知处理器
	s.startQueueProcessor()

	// 启动定期清理任务
	s.mu.Lock()
	s.cleanupStopChan = make(chan struct{})
	s.mu.Unlock()
	go s.startCleanupTask()

	s.logger.Info("通知功能初始化完成")
	return nil
}

// initializeChannels 初始化通知渠道
func (s *NotificationService) initializeChannels(settings *notification.Settings) error {
	// 创建通道实例
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

	// 更新通道
	s.mu.Lock()
	s.channels = channels
	s.mu.Unlock()

	s.logger.Info("通知渠道初始化完成", zap.Int("总渠道数", len(channels)))
	return nil
}

// loadPendingNotificationsToMemory 从数据库加载待处理通知到内存队列
func (s *NotificationService) loadPendingNotificationsToMemory() error {
	s.logger.Info("开始加载待处理通知到内存队列")

	// 获取所有待处理的通知（包括需要重试的）
	notifications, err := repository.Notification.GetPendingNotifications(0) // 0表示获取所有
	if err != nil {
		return fmt.Errorf("获取待处理通知失败: %w", err)
	}

	loadedCount := 0
	for _, notif := range notifications {
		select {
		case s.memoryQueue <- notif:
			loadedCount++
		default:
			s.logger.Warn("内存队列已满，无法加载更多通知", zap.Uint("notificationId", notif.ID))
		}
	}

	s.logger.Info("已加载待处理通知到内存队列",
		zap.Int("加载数量", loadedCount),
		zap.Int("队列中数量", len(s.memoryQueue)))

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
	s.queueProcessing = true
	s.mu.Unlock()

	go func() {
		defer func() {
			// 安全地修改队列处理状态
			s.mu.Lock()
			s.queueProcessing = false
			// 如果 stopChan 还存在，说明是正常退出，需要重置
			if s.stopChan != nil {
				s.stopChan = nil
			}
			s.mu.Unlock()
		}()

		s.logger.Info("通知队列处理器已启动，开始消费内存队列")

		// 无限循环消费内存队列
		for {
			select {
			case notification := <-s.memoryQueue:
				// 处理单个通知
				s.processNotification(notification)
			case <-s.stopChan:
				s.logger.Info("收到停止信号，通知队列处理器即将退出")
				return
			}
		}
	}()
}

// stopQueueProcessor 停止队列处理器
func (s *NotificationService) stopQueueProcessor() {
	s.mu.Lock()
	if !s.queueProcessing || s.stopChan == nil {
		s.mu.Unlock()
		return
	}

	stopCh := s.stopChan
	// 重置 stopChan 以防止重复关闭
	s.stopChan = nil
	s.queueProcessing = false
	s.mu.Unlock()

	// 关闭停止通道，通知处理器退出
	close(stopCh)

	// 等待一段时间让处理器正常退出
	time.Sleep(100 * time.Millisecond)

	s.logger.Info("通知队列处理器已停止")
}

// processNotification 处理单个通知
func (s *NotificationService) processNotification(notif *notification.Queue) {
	s.mu.RLock()
	if s.settings == nil || !s.settings.Enabled || len(s.channels) == 0 {
		s.mu.RUnlock()
		s.logger.Debug("通知功能已禁用，跳过处理")
		return
	}

	maxRetries := s.settings.QueueSettings.MaxRetries
	retryInterval := s.settings.QueueSettings.RetryInterval
	s.mu.RUnlock()

	s.logger.Debug("开始处理通知",
		zap.Uint("id", notif.ID),
		zap.String("channelType", notif.ChannelType),
		zap.Int("retryCount", notif.RetryCount))

	// 更新状态为处理中（仅在数据库中有ID时更新）
	if notif.ID > 0 {
		err := repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusProcessing, "")
		if err != nil {
			s.logger.Error("更新通知状态失败", zap.Error(err), zap.Uint("id", notif.ID))
			return
		}
	}

	// 获取渠道
	s.mu.RLock()
	channelType := notification.NotificationChannelType(notif.ChannelType)
	channel, ok := s.channels[channelType]
	s.mu.RUnlock()

	if !ok {
		errMsg := fmt.Sprintf("通知渠道不可用: %s", notif.ChannelType)
		s.logger.Warn(errMsg, zap.Uint("id", notif.ID))
		if notif.ID > 0 {
			repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusFailed, errMsg)
		}
		return
	}

	// 解析数据
	var data notification.TaskNotificationData
	err := json.Unmarshal([]byte(notif.Payload), &data)
	if err != nil {
		errMsg := fmt.Sprintf("解析通知数据失败: %v", err)
		s.logger.Error(errMsg, zap.Uint("id", notif.ID))
		if notif.ID > 0 {
			repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusFailed, errMsg)
		}
		return
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
			// 计算新的重试次数和下次重试时间
			newRetryCount := notif.RetryCount + 1
			nextRetry := time.Now().Add(time.Duration(retryInterval) * time.Second)

			// 先更新数据库状态为pending，设置下次重试时间和重试次数
			if notif.ID > 0 {
				err = repository.Notification.RequeueNotification(notif.ID, newRetryCount, nextRetry, errMsg)
				if err != nil {
					s.logger.Error("重新入队通知失败", zap.Error(err), zap.Uint("id", notif.ID))
					return
				}
			}

			// 更新内存中的重试次数（与数据库保持一致）
			notif.RetryCount = newRetryCount
			notif.Status = notification.StatusPending
			notif.UpdatedAt = time.Now()

			// 延迟后重新入内存队列
			go func(retryNotif *notification.Queue, delay time.Duration) {
				time.Sleep(delay)
				select {
				case s.memoryQueue <- retryNotif:
					s.logger.Info("通知已重新入队",
						zap.Uint("id", retryNotif.ID),
						zap.Int("retryCount", retryNotif.RetryCount))
				default:
					s.logger.Warn("内存队列已满，无法重新入队通知", zap.Uint("id", retryNotif.ID))
					// 如果内存队列满了，通知仍在数据库中保持pending状态，下次服务重启时会重新加载
				}
			}(notif, time.Duration(retryInterval)*time.Second)

			s.logger.Info("通知将在稍后重试",
				zap.Uint("id", notif.ID),
				zap.Int("retryCount", newRetryCount),
				zap.Time("nextRetryTime", nextRetry))
		} else {
			// 达到最大重试次数，标记为最终失败
			if notif.ID > 0 {
				err = repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusFailed, errMsg)
				if err != nil {
					s.logger.Error("更新通知状态失败", zap.Error(err), zap.Uint("id", notif.ID))
				}
			}
			s.logger.Warn("通知达到最大重试次数，已标记为失败", zap.Uint("id", notif.ID))
		}
		return
	}

	// 发送成功，更新为已发送状态
	if notif.ID > 0 {
		err := repository.Notification.UpdateNotificationStatus(notif.ID, notification.StatusSent, "")
		if err != nil {
			s.logger.Error("更新通知发送成功状态失败", zap.Error(err), zap.Uint("id", notif.ID))
			// 虽然发送成功了，但数据库更新失败，记录警告但不影响主流程
		}
	}
	s.logger.Info("通知已成功发送",
		zap.Uint("id", notif.ID),
		zap.String("channelType", notif.ChannelType),
		zap.String("taskName", data.TaskName))
}

// startCleanupTask 启动定期清理任务
func (s *NotificationService) startCleanupTask() {
	// 每24小时清理一次历史数据
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	s.logger.Info("通知清理任务已启动")

	for {
		select {
		case <-ticker.C:
			// 清理30天前的已发送通知
			cleanCutoff := time.Now().AddDate(0, 0, -30)
			err := repository.Notification.CleanSentNotifications(cleanCutoff)
			if err != nil {
				s.logger.Error("清理已发送通知失败", zap.Error(err))
			} else {
				s.logger.Info("已清理历史通知数据")
			}
		case <-s.cleanupStopChan:
			s.logger.Info("收到停止信号，清理任务即将退出")
			return
		}
	}
}

// stopNotificationFeatures 停止通知功能
func (s *NotificationService) stopNotificationFeatures() {
	s.logger.Info("停止通知功能")

	// 停止队列处理器
	s.stopQueueProcessor()

	// 停止清理任务
	s.mu.Lock()
	s.channels = make(map[notification.NotificationChannelType]notification_channel.Channel)
	s.mu.Unlock()

	s.logger.Info("通知功能已停止")
}

// loadNotificationSettings 从配置服务加载通知设置
func (s *NotificationService) loadNotificationSettings() (*notification.Settings, error) {
	// 通过配置服务获取通知设置
	req := &configRequest.ConfigByCodeReq{
		Code: "NOTIFICATION_SETTINGS",
	}

	configInfo, err := Config.GetConfigByCode(req)
	if err != nil {
		s.logger.Debug("获取通知配置失败，将创建默认配置", zap.Error(err))
		// 配置不存在，创建默认配置
		return s.createDefaultNotificationSettings()
	}

	// 解析 JSON 配置
	var settings notification.Settings
	err = json.Unmarshal([]byte(configInfo.Value), &settings)
	if err != nil {
		s.logger.Error("解析通知配置JSON失败", zap.Error(err), zap.String("configValue", configInfo.Value))
		return nil, fmt.Errorf("解析通知配置失败: %w", err)
	}

	s.logger.Debug("成功加载通知配置")
	return &settings, nil
}

// createDefaultNotificationSettings 创建默认通知设置
func (s *NotificationService) createDefaultNotificationSettings() (*notification.Settings, error) {
	// 创建默认配置
	defaultSettings := notification.DefaultSettings()

	// 序列化为 JSON
	jsonData, err := json.Marshal(defaultSettings)
	if err != nil {
		s.logger.Error("序列化默认通知配置失败", zap.Error(err))
		return nil, fmt.Errorf("序列化默认通知配置失败: %w", err)
	}

	// 创建配置记录
	createReq := &configRequest.ConfigCreateReq{
		Name:  "通知系统配置",
		Code:  "NOTIFICATION_SETTINGS",
		Value: string(jsonData),
	}

	err = Config.Create(createReq)
	if err != nil {
		s.logger.Error("创建默认通知配置失败", zap.Error(err))
		return nil, fmt.Errorf("创建默认通知配置失败: %w", err)
	}

	s.logger.Info("已创建默认通知配置")
	return defaultSettings, nil
}
