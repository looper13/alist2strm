package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/utils"
	"go.uber.org/zap"
)

// CloudDriveConfig CloudDrive 配置结构
type CloudDriveConfig struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CloudDriveService CloudDrive 服务
type CloudDriveService struct {
	client *utils.Client
	config *CloudDriveConfig
	logger *zap.Logger
	mu     sync.RWMutex
}

var (
	cloudDriveServiceInstance *CloudDriveService
	cloudDriveServiceOnce     sync.Once
)

// InitializeCloudDriveService 初始化 CloudDrive 服务
func InitializeCloudDriveService(logger *zap.Logger) *CloudDriveService {
	cloudDriveServiceOnce.Do(func() {
		cloudDriveServiceInstance = &CloudDriveService{
			logger: logger,
		}
		if err := cloudDriveServiceInstance.loadConfigAndInitClient(); err != nil {
			logger.Warn("CloudDrive 服务初始化时加载配置失败，相关功能可能不可用", zap.Error(err))
		}
	})
	return cloudDriveServiceInstance
}

// GetCloudDriveService 获取 CloudDrive 服务实例
func GetCloudDriveService() *CloudDriveService {
	return cloudDriveServiceInstance
}

// loadConfigAndInitClient 从数据库加载配置并初始化客户端
func (s *CloudDriveService) loadConfigAndInitClient() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, err := repository.Config.GetByCode("CLOUD_DRIVE")
	if err != nil {
		s.logger.Error("获取 CloudDrive 配置失败", zap.Error(err))
		return err
	}
	if config == nil || config.Value == "" {
		s.logger.Warn("未找到或未配置 CloudDrive，相关功能将不可用")
		s.config = nil
		s.client = nil
		return nil
	}

	var cdConfig CloudDriveConfig
	if err := json.Unmarshal([]byte(config.Value), &cdConfig); err != nil {
		s.logger.Error("解析 CloudDrive 配置失败", zap.Error(err))
		return err
	}

	s.config = &cdConfig
	s.client = utils.NewClient(cdConfig.Host, cdConfig.Username, cdConfig.Password, "", "")
	if err := s.client.Login(); err != nil {
		s.logger.Error("登录 CloudDrive 失败", zap.Error(err))
		s.client = nil // 登录失败，客户端置空
		return fmt.Errorf("登录 CloudDrive 失败: %w", err)
	}

	s.logger.Info("CloudDrive 服务初始化并登录成功", zap.String("host", cdConfig.Host))
	return nil
}

// ListFiles 获取指定目录下的文件列表
func (s *CloudDriveService) ListFiles(path string) ([]AListFile, error) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("CloudDrive 客户端未初始化或登录失败")
	}

	// 调用 gRPC 客户端获取文件列表
	// getImmediateSubFiles 已经处理了流式数据的接收
	cloudDriveFiles, err := client.GetImmediateSubFiles(path, false, false)
	if err != nil {
		return nil, fmt.Errorf("从 CloudDrive 获取文件列表失败 [%s]: %w", path, err)
	}
	if cloudDriveFiles == nil {
		return []AListFile{}, nil
	}

	// 将 clouddrive.CloudDriveFile 转换为通用的 AListFile
	var aListFiles []AListFile
	for _, cdFile := range cloudDriveFiles {
		if cdFile == nil {
			continue
		}

		var sha1Hash string
		if cdFile.FileHashes != nil {
			// clouddrive.CloudDriveFile_Sha1 的值为 2
			if hash, ok := cdFile.FileHashes[2]; ok {
				sha1Hash = hash
			}
		}

		aFile := AListFile{
			Name:     cdFile.Name,
			Size:     cdFile.Size,
			IsDir:    cdFile.IsDirectory,
			Modified: cdFile.WriteTime.AsTime(),
			Sign:     cdFile.Id, // 使用 CloudDrive 的文件ID作为签名，用于后续构建URL
			Thumb:    cdFile.ThumbnailUrl,
			Type:     0, // Type 在我们的逻辑中不关键，设为默认值
			HashInfo: struct {
				Sha1 string `json:"sha1"`
			}{
				Sha1: sha1Hash,
			},
		}
		aListFiles = append(aListFiles, aFile)
	}

	return aListFiles, nil
}

// GetFileURL 获取文件的完整访问URL
func (s *CloudDriveService) GetFileURL(sourcePath, filename, sign string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Host == "" {
		s.logger.Warn("获取文件 URL 失败：CloudDrive 配置未初始化")
		return ""
	}

	// 确保 Host 格式正确 (去除末尾的斜杠)
	baseURL := strings.TrimSuffix(s.config.Host, "/")

	// 构建完整的文件路径
	var fullPath string
	if sourcePath != "" {
		fullPath = fmt.Sprintf("%s/%s", strings.Trim(sourcePath, "/"), filename)
	} else {
		fullPath = filename
	}

	// 构建最终的 URL
	return fmt.Sprintf("http://%s/static/http/%s/False/%s", strings.TrimPrefix(baseURL, "http://"), strings.TrimPrefix(baseURL, "http://"), fullPath)
}
