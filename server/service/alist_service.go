package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MccRay-s/alist2strm/repository"
	"go.uber.org/zap"
)

// AListConfig AList 配置结构
type AListConfig struct {
	Host             string `json:"host"`             // AList 服务器地址
	Username         string `json:"username"`         // 用户名
	Password         string `json:"password"`         // 密码
	Token            string `json:"token"`            // API Token
	Domain           string `json:"domain"`           // 访问域名（用于生成文件URL）
	ReqRetryCount    int    `json:"reqRetryCount"`    // 重试次数
	ReqInterval      int64  `json:"reqInterval"`      // 请求间隔时间(毫秒)
	ReqRetryInterval int64  `json:"reqRetryInterval"` // 重试间隔时间(毫秒)
}

// AListFile Alist 文件信息
type AListFile struct {
	Name     string    `json:"name"`     // 文件名
	Size     int64     `json:"size"`     // 文件大小
	IsDir    bool      `json:"is_dir"`   // 是否是目录
	Modified time.Time `json:"modified"` // 修改时间
	Sign     string    `json:"sign"`     // 签名
	Thumb    string    `json:"thumb"`    // 缩略图
	Type     int       `json:"type"`     // 类型
	HashInfo struct {
		Sha1 string `json:"sha1"` // SHA1 哈希
	} `json:"hash_info"` // 哈希信息
}

// AListListResponse Alist 目录列表响应
type AListListResponse struct {
	Code    int    `json:"code"`    // 状态码
	Message string `json:"message"` // 消息
	Data    struct {
		Content  []AListFile `json:"content"`  // 文件列表
		Total    int         `json:"total"`    // 总数
		Readme   string      `json:"readme"`   // 说明
		Provider string      `json:"provider"` // 提供者
	} `json:"data"`
}

// AListClient Alist API 客户端
type AListClient struct {
	config     *AListConfig
	httpClient *http.Client
	logger     *zap.Logger
	mu         sync.Mutex
}

// AListService AList 服务
type AListService struct {
	client *AListClient
	config *AListConfig
	logger *zap.Logger
	mu     sync.RWMutex
}

// OnConfigUpdate 实现配置更新监听器接口
func (s *AListService) OnConfigUpdate(code string) error {
	if code == "ALIST" {
		s.logger.Info("检测到 AList 配置更新，重新加载配置")
		return s.ReloadConfig()
	}
	return nil
}

// 包级别的全局实例
var (
	alistServiceInstance *AListService
	alistServiceOnce     sync.Once
)

// Initialize 初始化 AList 服务
func InitializeAListService(logger *zap.Logger) *AListService {
	alistServiceOnce.Do(func() {
		alistServiceInstance = &AListService{
			logger: logger,
		}

		// 加载配置，但不阻断启动
		if err := alistServiceInstance.loadConfig(); err != nil {
			logger.Warn("AList 服务初始化时加载配置失败，部分功能可能不可用", zap.Error(err))
		}

		// 注册配置更新监听器
		GetConfigListenerService().Register("ALIST", alistServiceInstance)
	})
	return alistServiceInstance
}

// GetAListService 获取 AList 服务实例
func GetAListService() *AListService {
	return alistServiceInstance
}

// loadConfig 从数据库加载配置
func (s *AListService) loadConfig() error {
	config, err := repository.Config.GetByCode("ALIST")
	if err != nil || config == nil {
		// 配置不存在或为空，记录日志但不返回错误，允许程序继续运行
		s.logger.Warn("未找到 AList 配置或配置为空，AList 功能将不可用", zap.Error(err))

		// 初始化空配置
		s.mu.Lock()
		s.config = nil
		s.client = nil
		s.mu.Unlock()

		return nil
	}

	// 确保配置值不为空
	if config.Value == "" {
		s.logger.Warn("AList 配置为空字符串，AList 功能将不可用")

		// 初始化空配置
		s.mu.Lock()
		s.config = nil
		s.client = nil
		s.mu.Unlock()

		return nil
	}

	var alistConfig AListConfig
	if err := json.Unmarshal([]byte(config.Value), &alistConfig); err != nil {
		s.logger.Error("解析 AList 配置失败", zap.Error(err))

		// 初始化空配置
		s.mu.Lock()
		s.config = nil
		s.client = nil
		s.mu.Unlock()

		return nil // 错误时不返回错误，允许程序继续运行
	}

	s.mu.Lock()
	s.config = &alistConfig
	s.client = &AListClient{
		config: &alistConfig,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		logger: s.logger,
	}
	s.mu.Unlock()

	s.logger.Info("AList 配置加载成功", zap.String("host", alistConfig.Host))
	return nil
}

// TestConnection 测试连接
func (s *AListService) TestConnection() error {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("未配置 AList，请先完成配置")
	}

	// 尝试获取根目录列表来测试连接
	_, err := client.ListFiles("/")
	if err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	return nil
}

// ListFiles 获取指定目录下的文件列表
func (s *AListService) ListFiles(dirPath string) ([]AListFile, error) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("AList 客户端未初始化")
	}

	return client.ListFiles(dirPath)
}

// GetFileURL 获取文件的完整访问 URL
func (s *AListService) GetFileURL(sourcePath, filename, sign string) string {
	s.mu.RLock()
	if s.config == nil {
		s.mu.RUnlock()
		s.logger.Warn("获取文件 URL 失败：AList 配置未初始化")
		return ""
	}

	// 优先使用 Domain，如果为空则尝试使用 Host
	var baseURL string
	if s.config.Domain != "" {
		baseURL = s.config.Domain
	} else if s.config.Host != "" {
		// 确保 Host 有正确的 http:// 或 https:// 前缀
		baseURL = s.config.Host
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			baseURL = "http://" + baseURL
		}
		s.logger.Info("Domain 为空，使用 Host 作为基础 URL", zap.String("host", baseURL))
	}
	s.mu.RUnlock()

	if baseURL == "" {
		s.logger.Warn("获取文件 URL 失败：AList Domain 和 Host 均为空")
		return ""
	}
	// 确保域名格式正确 (去除末尾的斜杠)
	baseURL = strings.TrimSuffix(baseURL, "/")

	// 标准化路径处理
	cleanPath := strings.TrimPrefix(sourcePath, "/")
	cleanPath = strings.Trim(cleanPath, "/")

	// 如果路径为空，直接使用文件名
	if cleanPath == "" {
		fileURL := fmt.Sprintf("%s/d/%s", baseURL, filename)
		// 添加签名参数（如果有）
		if sign != "" {
			fileURL += "?sign=" + sign
		}
		return fileURL
	}

	// 构建完整文件 URL
	fileURL := fmt.Sprintf("%s/d/%s/%s", baseURL, cleanPath, filename)

	// 添加签名参数（如果有）
	if sign != "" {
		fileURL += "?sign=" + sign
	}

	return fileURL
}

// ReloadConfig 重新加载配置
func (s *AListService) ReloadConfig() error {
	return s.loadConfig()
}

// IsConfigured 检查是否已配置
func (s *AListService) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.Host != ""
}

// =============================================================================
// AListClient 客户端实现
// =============================================================================

// doRequest 执行 HTTP 请求，包含重试和请求间隔逻辑
func (c *AListClient) doRequest(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 添加认证头
	if c.config.Token != "" {
		req.Header.Set("Authorization", c.config.Token)
	}

	var lastErr error
	maxRetries := c.config.ReqRetryCount
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			// 重试间隔
			retryInterval := c.config.ReqRetryInterval
			if retryInterval <= 0 {
				retryInterval = 1000
			}
			time.Sleep(time.Duration(retryInterval) * time.Millisecond)
		}

		// 请求间隔
		if c.config.ReqInterval > 0 {
			time.Sleep(time.Duration(c.config.ReqInterval) * time.Millisecond)
		}

		resp, err := c.httpClient.Do(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

// ListFiles 获取指定目录下的所有文件（非递归）
func (c *AListClient) ListFiles(dirPath string) ([]AListFile, error) {
	if c.config == nil {
		return nil, fmt.Errorf("客户端未配置")
	}

	// 构建请求
	reqBody := map[string]interface{}{
		"path":     dirPath,
		"password": "",
		"page":     1,
		"per_page": 0,
		"refresh":  false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.Host+"/api/fs/list", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var listResp AListListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, err
	}

	if listResp.Code != 200 {
		return nil, fmt.Errorf("API错误: %s", listResp.Message)
	}

	return listResp.Data.Content, nil
}
