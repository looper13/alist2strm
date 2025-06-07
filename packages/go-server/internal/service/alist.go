package service

import (
	"alist2strm/internal/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	configCode = "ALIST" // 配置表中的code
)

type AlistService struct {
	db         *gorm.DB
	logger     *zap.Logger
	config     *model.AlistConfig
	httpClient *http.Client
}

// NewAlistService 创建新的AlistService实例
func NewAlistService(db *gorm.DB, logger *zap.Logger) (*AlistService, error) {
	service := &AlistService{
		db:     db,
		logger: logger,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	// 加载配置
	if err := service.loadConfig(); err != nil {
		return nil, err
	}

	return service, nil
}

// loadConfig 从数据库加载Alist配置
func (s *AlistService) loadConfig() error {
	var config model.Config
	if err := s.db.Where("code = ?", configCode).First(&config).Error; err != nil {
		s.logger.Error("加载Alist配置失败", zap.Error(err))
		return err
	}

	var alistConfig model.AlistConfig
	if err := json.Unmarshal([]byte(config.Value), &alistConfig); err != nil {
		s.logger.Error("解析Alist配置失败", zap.Error(err))
		return err
	}

	s.config = &alistConfig
	return nil
}

// List 获取目录列表
func (s *AlistService) List(dirPath string) ([]model.FileInfo, error) {
	url := fmt.Sprintf("%s/api/fs/list", s.config.Host)
	body := strings.NewReader(fmt.Sprintf(`{"path":"%s"}`, dirPath))

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", s.config.Token)
	req.Header.Set("Content-Type", "application/json")

	var response model.ListResponse
	err = s.doRequest(req, &response)
	if err != nil {
		return nil, err
	}

	return response.Data.Content, nil
}

// Get 获取文件信息
func (s *AlistService) Get(filePath string) (*model.FileInfo, error) {
	url := fmt.Sprintf("%s/api/fs/get", s.config.Host)
	body := strings.NewReader(fmt.Sprintf(`{"path":"%s"}`, filePath))

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", s.config.Token)
	req.Header.Set("Content-Type", "application/json")

	var response model.FsGetResponse
	err = s.doRequest(req, &response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// GetDownloadLink 获取文件下载链接
func (s *AlistService) GetDownloadLink(filePath string) (string, error) {
	fileInfo, err := s.Get(filePath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/p%s?sign=%s", s.config.Host, fileInfo.Path, fileInfo.Sign), nil
}

// GetFileContent 获取文件内容
func (s *AlistService) GetFileContent(filePath string) ([]byte, error) {
	downloadLink, err := s.GetDownloadLink(filePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", downloadLink, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// IsVideo 检查文件是否为视频文件
func (s *AlistService) IsVideo(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpeg", ".mpg", ".ts":
		return true
	default:
		return false
	}
}

// IsMetadata 检查文件是否为元数据文件
func (s *AlistService) IsMetadata(filename string, metadataExts string) bool {
	ext := strings.ToLower(path.Ext(filename))
	exts := strings.Split(metadataExts, ",")
	for _, allowedExt := range exts {
		if strings.TrimSpace(allowedExt) == ext {
			return true
		}
	}
	return false
}

// IsSubtitle 检查文件是否为字幕文件
func (s *AlistService) IsSubtitle(filename string, subtitleExts string) bool {
	ext := strings.ToLower(path.Ext(filename))
	exts := strings.Split(subtitleExts, ",")
	for _, allowedExt := range exts {
		if strings.TrimSpace(allowedExt) == ext {
			return true
		}
	}
	return false
}

// doRequest 执行HTTP请求并处理重试逻辑
func (s *AlistService) doRequest(req *http.Request, response interface{}) error {
	var lastErr error
	for i := 0; i <= s.config.ReqRetryCount; i++ {
		if i > 0 {
			s.logger.Warn("请求重试",
				zap.String("url", req.URL.String()),
				zap.Int("attempt", i),
				zap.Error(lastErr))
			time.Sleep(time.Duration(s.config.ReqRetryInterval) * time.Millisecond)
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		if err := json.Unmarshal(body, response); err != nil {
			lastErr = err
			continue
		}

		time.Sleep(time.Duration(s.config.ReqInterval) * time.Millisecond)
		return nil
	}

	return fmt.Errorf("请求失败，已重试%d次: %v", s.config.ReqRetryCount, lastErr)
}
