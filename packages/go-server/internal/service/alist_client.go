package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AlistFile Alist 文件信息
type AlistFile struct {
	Name     string    `json:"name"`     // 文件名
	Size     int64     `json:"size"`     // 文件大小
	IsDir    bool      `json:"is_dir"`   // 是否是目录
	Modified time.Time `json:"modified"` // 修改时间
	Sign     string    `json:"sign"`     // 签名
	Thumb    string    `json:"thumb"`    // 缩略图
	Type     int       `json:"type"`     // 类型
}

// AlistListResponse Alist 目录列表响应
type AlistListResponse struct {
	Code    int    `json:"code"`    // 状态码
	Message string `json:"message"` // 消息
	Data    struct {
		Content  []AlistFile `json:"content"`  // 文件列表
		Total    int         `json:"total"`    // 总数
		Readme   string      `json:"readme"`   // 说明
		Provider string      `json:"provider"` // 提供者
	} `json:"data"`
}

// AlistConfig Alist 配置
type AlistConfig struct {
	Token            string `json:"token"`            // API Token
	Host             string `json:"host"`             // Alist 服务器地址
	Domain           string `json:"domain"`           // 访问域名
	ReqInterval      int64  `json:"reqInterval"`      // 请求间隔时间(毫秒)
	ReqRetryCount    int    `json:"reqRetryCount"`    // 失败重试次数
	ReqRetryInterval int64  `json:"reqRetryInterval"` // 重试间隔时间(毫秒)
}

// AlistClient Alist API 客户端
type AlistClient struct {
	config      *AlistConfig
	httpClient  *http.Client
	logger      *zap.Logger
	lastReqTime time.Time  // 上一次请求时间
	mu          sync.Mutex // 用于保护 lastReqTime
}

var (
	alistClient *AlistClient
	alistOnce   sync.Once
)

// GetAlistClient 获取 AlistClient 单例
func GetAlistClient(logger *zap.Logger) *AlistClient {
	alistOnce.Do(func() {
		alistClient = &AlistClient{
			httpClient: &http.Client{
				Timeout: time.Second * 30,
			},
			logger: logger,
		}
	})
	return alistClient
}

// UpdateConfig 更新配置
func (c *AlistClient) UpdateConfig(config *AlistConfig) {
	c.config = config
}

// doRequest 执行 HTTP 请求，包含重试和请求间隔逻辑
func (c *AlistClient) doRequest(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	// 检查是否需要等待请求间隔
	if !c.lastReqTime.IsZero() {
		elapsed := time.Since(c.lastReqTime)
		if elapsed < time.Duration(c.config.ReqInterval)*time.Millisecond {
			time.Sleep(time.Duration(c.config.ReqInterval)*time.Millisecond - elapsed)
		}
	}
	c.lastReqTime = time.Now()
	c.mu.Unlock()

	var resp *http.Response
	var err error
	var retryCount int

	for retryCount = 0; retryCount <= c.config.ReqRetryCount; retryCount++ {
		if retryCount > 0 {
			c.logger.Info("重试请求",
				zap.String("url", req.URL.String()),
				zap.Int("attempt", retryCount),
				zap.Int("maxAttempts", c.config.ReqRetryCount))
			time.Sleep(time.Duration(c.config.ReqRetryInterval) * time.Millisecond)
		}

		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		if retryCount < c.config.ReqRetryCount {
			c.logger.Warn("请求失败，准备重试",
				zap.String("url", req.URL.String()),
				zap.Error(err),
				zap.Int("statusCode", resp.StatusCode))
		}
	}

	return nil, fmt.Errorf("请求失败，已重试 %d 次: %w", retryCount-1, err)
}

// ListFiles 获取指定目录下的所有文件（非递归）
func (c *AlistClient) ListFiles(dirPath string) ([]AlistFile, error) {
	return c.listSingleDirectory(dirPath)
}

// listSingleDirectory 获取单个目录下的所有文件（支持分页）
func (c *AlistClient) listSingleDirectory(dirPath string) ([]AlistFile, error) {
	if c.config == nil {
		return nil, fmt.Errorf("alist 配置未设置")
	}

	var allFiles []AlistFile
	page := 1
	const defaultPerPage = 100

	for {
		data := map[string]interface{}{
			"path":     dirPath,
			"password": "",
			"page":     page,
			"per_page": defaultPerPage,
			"refresh":  false,
			// "refresh":  page == 1, // 第一页时强制刷新目录缓存
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/fs/list", c.config.Host), bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", c.config.Token)

		resp, err := c.doRequest(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var response AlistListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}

		if response.Code != 200 {
			return nil, errors.New(response.Message)
		}

		// 添加当前页的文件到结果列表
		allFiles = append(allFiles, response.Data.Content...)

		c.logger.Debug("获取目录文件列表",
			zap.String("path", dirPath),
			zap.Int("page", page),
			zap.Int("currentCount", len(response.Data.Content)),
			zap.Int("totalCount", len(allFiles)))

		// 检查是否还有下一页
		if len(response.Data.Content) < defaultPerPage {
			break
		}
		page++
	}

	return allFiles, nil
}

// GetFileURL 获取文件的完整URL（包含签名如果有的话）
func (c *AlistClient) GetFileURL(sourcePath, filename, sign string) string {
	baseURL := fmt.Sprintf("%s/d%s/%s", c.config.Host, sourcePath, filename)
	if sign != "" {
		return fmt.Sprintf("%s?sign=%s", baseURL, sign)
	}
	return baseURL
}
