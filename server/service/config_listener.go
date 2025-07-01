package service

import (
	"sync"

	"github.com/MccRay-s/alist2strm/utils"
)

// ConfigUpdateListener 配置更新监听器接口
type ConfigUpdateListener interface {
	OnConfigUpdate(code string) error
}

// ConfigListenerService 配置监听器服务
type ConfigListenerService struct {
	listeners map[string][]ConfigUpdateListener
	mu        sync.RWMutex
}

var (
	configListenerInstance *ConfigListenerService
	configListenerOnce     sync.Once
)

// GetConfigListenerService 获取配置监听器服务实例
func GetConfigListenerService() *ConfigListenerService {
	configListenerOnce.Do(func() {
		configListenerInstance = &ConfigListenerService{
			listeners: make(map[string][]ConfigUpdateListener),
		}
	})
	return configListenerInstance
}

// Register 注册配置更新监听器
func (s *ConfigListenerService) Register(configCode string, listener ConfigUpdateListener) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listeners[configCode] == nil {
		s.listeners[configCode] = make([]ConfigUpdateListener, 0)
	}
	s.listeners[configCode] = append(s.listeners[configCode], listener)
}

// Notify 通知配置更新
func (s *ConfigListenerService) Notify(configCode string) {
	s.mu.RLock()
	// 检查是否存在该配置码的监听器
	if _, exists := s.listeners[configCode]; !exists {
		s.mu.RUnlock()
		utils.Warn("未找到配置码的监听器", "config_code", configCode)
		return
	}

	listeners := make([]ConfigUpdateListener, len(s.listeners[configCode]))
	copy(listeners, s.listeners[configCode])
	s.mu.RUnlock()

	utils.Info("开始通知配置更新", "config_code", configCode, "listener_count", len(listeners))

	for i, listener := range listeners {
		go func(index int, l ConfigUpdateListener) {
			if err := l.OnConfigUpdate(configCode); err != nil {
				utils.Error("配置更新通知失败", "config_code", configCode, "listener_index", index, "error", err.Error())
			} else {
				// 成功通知
				utils.Info("配置更新通知成功", "config_code", configCode, "listener_index", index)
			}
		}(i, listener)
	}
}
