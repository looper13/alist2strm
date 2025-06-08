package service

import (
	"alist2strm/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type EmbyNotificationService struct {
	config *EmbyConfig
	client *http.Client
	logger *zap.Logger
}

// EmbyConfig Emby 配置
type EmbyConfig struct {
	Enabled            bool              `json:"enabled"`
	ServerURL          string            `json:"serverUrl"`
	APIKey             string            `json:"apiKey"`
	UserID             string            `json:"userId"`
	Timeout            int               `json:"timeout"`
	AutoRefreshLibrary bool              `json:"autoRefreshLibrary"`
	NotificationEvents []string          `json:"notificationEvents"`
	Templates          map[string]string `json:"templates"`
}

// EmbyLibraryRefreshRequest 库刷新请求
type EmbyLibraryRefreshRequest struct {
	LibraryID string `json:"libraryId,omitempty"`
}

// EmbyLibraryInfo 库信息
type EmbyLibraryInfo struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
	Type string `json:"CollectionType"`
}

var (
	embyNotificationService *EmbyNotificationService
	embyNotificationOnce    sync.Once
)

// GetEmbyNotificationService 获取 EmbyNotificationService 单例
func GetEmbyNotificationService() *EmbyNotificationService {
	embyNotificationOnce.Do(func() {
		embyNotificationService = &EmbyNotificationService{
			client: &http.Client{
				Timeout: 30 * time.Second,
			},
			logger: utils.Logger,
		}
		embyNotificationService.loadConfig()
	})
	return embyNotificationService
}

// loadConfig 加载 Emby 配置
func (s *EmbyNotificationService) loadConfig() {
	configService := GetConfigService()
	configValue, err := configService.GetByCode("EMBY")
	if err != nil {
		s.logger.Warn("获取 Emby 配置失败", zap.Error(err))
		s.config = &EmbyConfig{Enabled: false}
		return
	}

	var config EmbyConfig
	if err := json.Unmarshal([]byte(configValue.Value), &config); err != nil {
		s.logger.Error("解析 Emby 配置失败", zap.Error(err))
		s.config = &EmbyConfig{Enabled: false}
		return
	}

	s.config = &config
	if s.config.Timeout > 0 {
		s.client.Timeout = time.Duration(s.config.Timeout) * time.Second
	}

	s.logger.Info("加载 Emby 配置成功",
		zap.Bool("enabled", s.config.Enabled),
		zap.String("serverUrl", s.config.ServerURL))
}

// ReloadConfig 重新加载配置
func (s *EmbyNotificationService) ReloadConfig() {
	s.loadConfig()
}

// IsEnabled 检查是否启用
func (s *EmbyNotificationService) IsEnabled() bool {
	return s.config != nil && s.config.Enabled && s.config.ServerURL != "" && s.config.APIKey != ""
}

// SendTaskCompletedNotification 发送任务完成通知
func (s *EmbyNotificationService) SendTaskCompletedNotification(payload map[string]interface{}) (bool, string) {
	if !s.IsEnabled() {
		return false, "Emby 通知未启用或配置不完整"
	}

	// 检查是否需要发送此类型通知
	if !s.shouldSendNotification("task_completed") {
		return true, "跳过此类型通知"
	}

	// 自动刷新媒体库
	if s.config.AutoRefreshLibrary {
		if err := s.RefreshLibrary(""); err != nil {
			s.logger.Warn("刷新 Emby 媒体库失败", zap.Error(err))
		}
	}

	return true, "Emby 任务完成通知处理成功"
}

// SendTaskFailedNotification 发送任务失败通知
func (s *EmbyNotificationService) SendTaskFailedNotification(payload map[string]interface{}) (bool, string) {
	if !s.IsEnabled() {
		return false, "Emby 通知未启用或配置不完整"
	}

	// 检查是否需要发送此类型通知
	if !s.shouldSendNotification("task_failed") {
		return true, "跳过此类型通知"
	}

	// 任务失败时不自动刷新库
	return true, "Emby 任务失败通知处理成功"
}

// SendFileInvalidNotification 发送文件失效通知
func (s *EmbyNotificationService) SendFileInvalidNotification(payload map[string]interface{}) (bool, string) {
	if !s.IsEnabled() {
		return false, "Emby 通知未启用或配置不完整"
	}

	// 检查是否需要发送此类型通知
	if !s.shouldSendNotification("file_invalid") {
		return true, "跳过此类型通知"
	}

	// 文件失效时自动刷新库以移除失效项
	if s.config.AutoRefreshLibrary {
		if err := s.RefreshLibrary(""); err != nil {
			s.logger.Warn("刷新 Emby 媒体库失败", zap.Error(err))
		}
	}

	return true, "Emby 文件失效通知处理成功"
}

// RefreshLibrary 刷新媒体库
func (s *EmbyNotificationService) RefreshLibrary(libraryID string) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Emby 服务未启用")
	}

	// 如果没有指定库ID，刷新所有库
	if libraryID == "" {
		libraries, err := s.GetLibraries()
		if err != nil {
			return fmt.Errorf("获取媒体库列表失败: %v", err)
		}

		for _, library := range libraries {
			if err := s.refreshSingleLibrary(library.ID); err != nil {
				s.logger.Error("刷新媒体库失败",
					zap.String("libraryId", library.ID),
					zap.String("libraryName", library.Name),
					zap.Error(err))
			}
		}
		return nil
	}

	return s.refreshSingleLibrary(libraryID)
}

// refreshSingleLibrary 刷新单个媒体库
func (s *EmbyNotificationService) refreshSingleLibrary(libraryID string) error {
	url := fmt.Sprintf("%s/emby/Items/%s/Refresh?api_key=%s",
		s.config.ServerURL, libraryID, s.config.APIKey)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("刷新媒体库失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	s.logger.Info("刷新 Emby 媒体库成功", zap.String("libraryId", libraryID))
	return nil
}

// GetLibraries 获取媒体库列表
func (s *EmbyNotificationService) GetLibraries() ([]EmbyLibraryInfo, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Emby 服务未启用")
	}

	url := fmt.Sprintf("%s/emby/Library/VirtualFolders?api_key=%s",
		s.config.ServerURL, s.config.APIKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取媒体库列表失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var libraries []EmbyLibraryInfo
	if err := json.NewDecoder(resp.Body).Decode(&libraries); err != nil {
		return nil, err
	}

	return libraries, nil
}

// TestConnection 测试连接
func (s *EmbyNotificationService) TestConnection() error {
	if !s.IsEnabled() {
		return fmt.Errorf("Emby 服务未启用或配置不完整")
	}

	url := fmt.Sprintf("%s/emby/System/Info?api_key=%s",
		s.config.ServerURL, s.config.APIKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("连接测试失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

// shouldSendNotification 检查是否应该发送指定类型的通知
func (s *EmbyNotificationService) shouldSendNotification(eventType string) bool {
	if s.config.NotificationEvents == nil || len(s.config.NotificationEvents) == 0 {
		return true // 默认发送所有通知
	}

	for _, event := range s.config.NotificationEvents {
		if event == eventType {
			return true
		}
	}

	return false
}

// GetConfig 获取当前配置
func (s *EmbyNotificationService) GetConfig() *EmbyConfig {
	return s.config
}

// UpdateConfig 更新配置
func (s *EmbyNotificationService) UpdateConfig(config *EmbyConfig) error {
	configService := GetConfigService()

	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := configService.UpdateByCode("EMBY", string(configData)); err != nil {
		return err
	}

	s.config = config
	if s.config.Timeout > 0 {
		s.client.Timeout = time.Duration(s.config.Timeout) * time.Second
	}

	return nil
}
