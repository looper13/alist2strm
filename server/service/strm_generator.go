package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MccRay-s/alist2strm/model/filehistory"
	"github.com/MccRay-s/alist2strm/model/task"
	"github.com/MccRay-s/alist2strm/model/tasklog"
	"github.com/MccRay-s/alist2strm/repository"
	"go.uber.org/zap"
)

// StrmConfig STRM 配置结构
type StrmConfig struct {
	DefaultSuffix string `json:"defaultSuffix"` // 默认媒体文件后缀
	ReplaceSuffix bool   `json:"replaceSuffix"` // 是否替换后缀
	URLEncode     bool   `json:"urlEncode"`     // 是否URL编码
}

// FileType 文件类型枚举
type FileType int

const (
	FileTypeMedia FileType = iota
	FileTypeMetadata
	FileTypeSubtitle
	FileTypeOther
)

// ProcessedFile 处理后的文件信息
type ProcessedFile struct {
	SourceFile   *AListFile
	TargetPath   string
	FileType     FileType
	Success      bool
	ErrorMessage string
}

// StrmGeneratorService STRM 文件生成服务
type StrmGeneratorService struct {
	alistService *AListService
	logger       *zap.Logger
	mu           sync.RWMutex
}

var (
	strmGeneratorInstance *StrmGeneratorService
	strmGeneratorOnce     sync.Once
)

// GetStrmGeneratorService 获取 STRM 生成服务实例
func GetStrmGeneratorService() *StrmGeneratorService {
	strmGeneratorOnce.Do(func() {
		strmGeneratorInstance = &StrmGeneratorService{}
	})
	return strmGeneratorInstance
}

// Initialize 初始化服务
func (s *StrmGeneratorService) Initialize(logger *zap.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
	s.alistService = GetAListService()
	logger.Info("STRM 生成服务初始化完成")
}

// IsInitialized 检查服务是否已初始化
func (s *StrmGeneratorService) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logger != nil && s.alistService != nil
}

// GenerateStrmFiles 生成 STRM 文件主方法
func (s *StrmGeneratorService) GenerateStrmFiles(taskID uint) error {
	// 检查服务是否已初始化
	if !s.IsInitialized() {
		return fmt.Errorf("STRM 生成服务未正确初始化")
	}

	// 获取任务信息
	taskInfo, err := repository.Task.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %w", err)
	}

	// 创建任务日志
	taskLog := &tasklog.TaskLog{
		TaskID:        taskID,
		Status:        tasklog.TaskLogStatusRunning,
		Message:       "开始生成 STRM 文件",
		StartTime:     time.Now(),
		TotalFile:     0,
		GeneratedFile: 0,
		SkipFile:      0,
		MetadataCount: 0,
		SubtitleCount: 0,
	}

	taskLogID, err := s.createTaskLog(taskLog)
	if err != nil {
		return fmt.Errorf("创建任务日志失败: %w", err)
	}

	// 加载 STRM 配置
	strmConfig, err := s.loadStrmConfig()
	if err != nil {
		s.updateTaskLogWithError(taskLogID, "加载 STRM 配置失败: "+err.Error())
		return err
	}

	// 开始处理文件
	s.logger.Info("开始处理任务",
		zap.Uint("taskId", taskID),
		zap.String("sourcePath", taskInfo.SourcePath),
		zap.String("targetPath", taskInfo.TargetPath))

	totalFiles, generatedFiles, skippedFiles, metadataFiles, subtitleFiles, err := s.processDirectory(
		taskInfo, strmConfig, taskLogID, taskInfo.SourcePath, taskInfo.TargetPath)

	// 更新任务日志
	endTime := time.Now()
	status := tasklog.TaskLogStatusCompleted
	message := "STRM 文件生成完成"

	if err != nil {
		status = tasklog.TaskLogStatusFailed
		message = "STRM 文件生成失败: " + err.Error()
	}

	updateData := map[string]interface{}{
		"status":         status,
		"message":        message,
		"end_time":       &endTime,
		"total_file":     totalFiles,
		"generated_file": generatedFiles,
		"skip_file":      skippedFiles,
		"metadata_count": metadataFiles,
		"subtitle_count": subtitleFiles,
	}

	if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
		s.logger.Error("更新任务日志失败", zap.Error(updateErr))
	}

	return err
}

// loadStrmConfig 加载 STRM 配置
func (s *StrmGeneratorService) loadStrmConfig() (*StrmConfig, error) {
	config, err := repository.Config.GetByCode("STRM")
	if err != nil {
		return nil, fmt.Errorf("获取 STRM 配置失败: %w", err)
	}

	var strmConfig StrmConfig
	if err := json.Unmarshal([]byte(config.Value), &strmConfig); err != nil {
		return nil, fmt.Errorf("解析 STRM 配置失败: %w", err)
	}

	return &strmConfig, nil
}

// processDirectory 处理目录（递归）
func (s *StrmGeneratorService) processDirectory(taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint, sourcePath, targetPath string) (int, int, int, int, int, error) {
	totalFiles := 0
	generatedFiles := 0
	skippedFiles := 0
	metadataFiles := 0
	subtitleFiles := 0

	// 获取当前目录的文件列表
	files, err := s.alistService.ListFiles(sourcePath)
	if err != nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("获取目录文件列表失败 [%s]: %w", sourcePath, err)
	}

	s.logger.Info("处理目录",
		zap.String("sourcePath", sourcePath),
		zap.String("targetPath", targetPath),
		zap.Int("fileCount", len(files)))

	// 创建目标目录
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return 0, 0, 0, 0, 0, fmt.Errorf("创建目标目录失败 [%s]: %w", targetPath, err)
	}

	// 处理当前目录下的文件
	for _, file := range files {
		totalFiles++

		// 构建完整路径
		currentSourcePath := filepath.Join(sourcePath, file.Name)
		currentTargetPath := filepath.Join(targetPath, file.Name)

		if file.IsDir {
			// 递归处理子目录
			subTotal, subGenerated, subSkipped, subMetadata, subSubtitle, err := s.processDirectory(
				taskInfo, strmConfig, taskLogID, currentSourcePath, currentTargetPath)
			if err != nil {
				return totalFiles, generatedFiles, skippedFiles, metadataFiles, subtitleFiles, err
			}

			totalFiles += subTotal
			generatedFiles += subGenerated
			skippedFiles += subSkipped
			metadataFiles += subMetadata
			subtitleFiles += subSubtitle
		} else {
			// 处理文件
			fileType := s.determineFileType(&file, taskInfo, strmConfig)
			processed := s.processFile(&file, fileType, taskInfo, strmConfig, taskLogID, currentSourcePath, currentTargetPath)

			// 记录文件历史
			s.recordFileHistory(taskInfo.ID, taskLogID, &file, currentSourcePath, currentTargetPath, fileType, processed.Success)

			// 统计结果
			switch fileType {
			case FileTypeMedia:
				if processed.Success {
					generatedFiles++
				} else {
					skippedFiles++
				}
			case FileTypeMetadata:
				if processed.Success {
					metadataFiles++
				} else {
					skippedFiles++
				}
			case FileTypeSubtitle:
				if processed.Success {
					subtitleFiles++
				} else {
					skippedFiles++
				}
			default:
				skippedFiles++
			}
		}
	}

	return totalFiles, generatedFiles, skippedFiles, metadataFiles, subtitleFiles, nil
}

// determineFileType 确定文件类型
func (s *StrmGeneratorService) determineFileType(file *AListFile, taskInfo *task.Task, strmConfig *StrmConfig) FileType {
	ext := strings.ToLower(filepath.Ext(file.Name))

	// 检查是否为媒体文件
	mediaExtensions := strings.Split(strings.ToLower(strmConfig.DefaultSuffix), ",")
	for _, mediaExt := range mediaExtensions {
		if ext == "."+strings.TrimSpace(mediaExt) {
			return FileTypeMedia
		}
	}

	// 检查是否为元数据文件
	if taskInfo.DownloadMetadata {
		metadataExtensions := strings.Split(strings.ToLower(taskInfo.MetadataExtensions), ",")
		for _, metaExt := range metadataExtensions {
			if ext == "."+strings.TrimSpace(metaExt) {
				return FileTypeMetadata
			}
		}
	}

	// 检查是否为字幕文件
	if taskInfo.DownloadSubtitle {
		subtitleExtensions := strings.Split(strings.ToLower(taskInfo.SubtitleExtensions), ",")
		for _, subExt := range subtitleExtensions {
			if ext == "."+strings.TrimSpace(subExt) {
				return FileTypeSubtitle
			}
		}
	}

	return FileTypeOther
}

// processFile 处理单个文件
func (s *StrmGeneratorService) processFile(file *AListFile, fileType FileType, taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint, sourcePath, targetPath string) *ProcessedFile {
	result := &ProcessedFile{
		SourceFile: file,
		TargetPath: targetPath,
		FileType:   fileType,
		Success:    false,
	}

	// 记录开始处理文件
	s.logger.Debug("开始处理文件",
		zap.String("文件名", file.Name),
		zap.String("文件类型", getFileTypeString(fileType)),
		zap.String("源路径", sourcePath),
		zap.String("目标路径", targetPath),
		zap.Int64("文件大小", file.Size))

	switch fileType {
	case FileTypeMedia:
		// 生成 STRM 文件 - 仅使用 AListFile 中已有信息
		result.Success, result.ErrorMessage = s.generateStrmFile(file, strmConfig, taskInfo, sourcePath, targetPath)
	case FileTypeMetadata, FileTypeSubtitle:
		// 下载元数据或字幕文件 - 仅使用 AListFile 中已有信息
		result.Success, result.ErrorMessage = s.downloadFile(file, sourcePath, targetPath)
	default:
		result.ErrorMessage = "不支持的文件类型，已跳过"
	}

	// 记录处理结果
	if !result.Success {
		s.logger.Warn("处理文件失败",
			zap.String("文件名", file.Name),
			zap.String("文件类型", getFileTypeString(fileType)),
			zap.String("错误", result.ErrorMessage))
	} else {
		s.logger.Debug("处理文件成功",
			zap.String("文件名", file.Name),
			zap.String("文件类型", getFileTypeString(fileType)))
	}

	return result
}

// getFileTypeString 获取文件类型的字符串表示
func getFileTypeString(fileType FileType) string {
	switch fileType {
	case FileTypeMedia:
		return "媒体文件"
	case FileTypeMetadata:
		return "元数据文件"
	case FileTypeSubtitle:
		return "字幕文件"
	default:
		return "其他文件"
	}
}

// generateStrmFile 生成 STRM 文件
func (s *StrmGeneratorService) generateStrmFile(file *AListFile, strmConfig *StrmConfig, taskConfig *task.Task, sourcePath, targetPath string) (bool, string) {
	// TODO 处理URLEncode

	// 构建 STRM 文件内容 - 直接使用 AListFile 中的信息，避免多余的 API 调用
	// 注意：GetFileURL 方法不会发起额外的 API 请求，仅使用配置和参数构建 URL
	fileURL := s.alistService.GetFileURL(filepath.Dir(sourcePath), file.Name, file.Sign)
	if fileURL == "" {
		return false, "无法生成文件URL，请检查 AList 配置是否完整"
	}

	// 生成 STRM 文件名
	var strmFileName string
	if strmConfig.ReplaceSuffix {
		// 替换后缀为 .strm
		nameWithoutExt := strings.TrimSuffix(file.Name, filepath.Ext(file.Name))
		strmFileName = nameWithoutExt + ".strm"
	} else {
		// 在原文件名后添加 .strm
		strmFileName = file.Name + ".strm"
	}

	// 构建完整的 STRM 文件路径
	strmFilePath := filepath.Join(filepath.Dir(targetPath), strmFileName)

	// 检查是否需要覆盖现有文件
	if !s.shouldOverwrite(strmFilePath, taskConfig) {
		return false, "文件已存在且不允许覆盖"
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(strmFilePath), 0755); err != nil {
		return false, fmt.Sprintf("创建目标目录失败: %v", err)
	}

	// 写入 STRM 文件
	if err := os.WriteFile(strmFilePath, []byte(fileURL), 0644); err != nil {
		return false, fmt.Sprintf("写入 STRM 文件失败: %v", err)
	}

	s.logger.Info("生成 STRM 文件成功",
		zap.String("sourceFile", file.Name),
		zap.String("strmFile", strmFilePath),
		zap.String("url", fileURL))

	return true, ""
}

// downloadFile 下载文件（元数据和字幕）
func (s *StrmGeneratorService) downloadFile(file *AListFile, sourcePath, targetPath string) (bool, string) {
	// 检查目标文件是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		return false, "文件已存在"
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return false, fmt.Sprintf("创建目标目录失败: %v", err)
	}

	// 直接使用 AListFile 中的信息构建文件 URL，不需要额外的 API 调用
	// 注意：GetFileURL 方法不会发起额外的 API 请求，仅使用配置和参数构建 URL
	fileURL := s.alistService.GetFileURL(filepath.Dir(sourcePath), file.Name, file.Sign)
	if fileURL == "" {
		return false, "无法生成文件下载URL，请检查 AList 配置是否完整"
	}

	// 实现 HTTP 下载逻辑
	if err := s.downloadFileFromURL(fileURL, targetPath); err != nil {
		return false, fmt.Sprintf("下载文件失败: %v", err)
	}

	s.logger.Info("下载文件成功",
		zap.String("sourceFile", file.Name),
		zap.String("targetPath", targetPath),
		zap.String("size", humanizeSize(file.Size)))

	return true, ""
}

// humanizeSize 将字节大小转换为友好的字符串表示
func humanizeSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// downloadFileFromURL 从 URL 下载文件
func (s *StrmGeneratorService) downloadFileFromURL(fileURL, targetPath string) error {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 发送 GET 请求
	resp, err := client.Get(fileURL)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载文件失败，状态码: %d", resp.StatusCode)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 创建目标文件
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer file.Close()

	// 复制内容
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// shouldOverwrite 检查是否应该覆盖文件
func (s *StrmGeneratorService) shouldOverwrite(filePath string, taskConfig *task.Task) bool {
	// 如果文件不存在，可以创建
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}

	// 根据任务配置的 Overwrite 字段决定是否覆盖
	return taskConfig.Overwrite
}

// recordFileHistory 记录文件历史
func (s *StrmGeneratorService) recordFileHistory(taskID, taskLogID uint, file *AListFile, sourcePath, targetPath string, fileType FileType, success bool) {
	if !success {
		return // 只记录成功处理的文件
	}

	fileTypeStr := s.getFileTypeString(fileType)

	fileHistory := &filehistory.FileHistory{
		TaskID:         taskID,
		TaskLogID:      taskLogID,
		FileName:       file.Name,
		SourcePath:     sourcePath,
		TargetFilePath: targetPath,
		FileSize:       file.Size,
		FileType:       fileTypeStr,
		FileSuffix:     filepath.Ext(file.Name),
		ModifiedAt:     &file.Modified,
		Hash:           file.HashInfo.Sha1,
	}

	if err := repository.FileHistory.Create(fileHistory); err != nil {
		s.logger.Error("记录文件历史失败",
			zap.String("fileName", file.Name),
			zap.Error(err))
	}
}

// getFileTypeString 获取文件类型字符串
func (s *StrmGeneratorService) getFileTypeString(fileType FileType) string {
	switch fileType {
	case FileTypeMedia:
		return "media"
	case FileTypeMetadata:
		return "metadata"
	case FileTypeSubtitle:
		return "subtitle"
	default:
		return "other"
	}
}

// createTaskLog 创建任务日志
func (s *StrmGeneratorService) createTaskLog(taskLog *tasklog.TaskLog) (uint, error) {
	if err := repository.TaskLog.Create(taskLog); err != nil {
		return 0, err
	}
	return taskLog.ID, nil
}

// updateTaskLogWithError 更新任务日志为错误状态
func (s *StrmGeneratorService) updateTaskLogWithError(taskLogID uint, errorMessage string) {
	endTime := time.Now()
	updateData := map[string]interface{}{
		"status":   tasklog.TaskLogStatusFailed,
		"message":  errorMessage,
		"end_time": &endTime,
	}

	if err := repository.TaskLog.UpdatePartial(taskLogID, updateData); err != nil {
		s.logger.Error("更新任务日志失败", zap.Error(err))
	}
}
