package service

import (
    "fmt"
    "sync"
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
    listeners := make([]ConfigUpdateListener, len(s.listeners[configCode]))
    copy(listeners, s.listeners[configCode])
    s.mu.RUnlock()
    
    for _, listener := range listeners {
        go func(l ConfigUpdateListener) {
            if err := l.OnConfigUpdate(configCode); err != nil {
                fmt.Printf("配置更新通知失败 [%s]: %v\n", configCode, err)
            }
        }(listener)
    }
}