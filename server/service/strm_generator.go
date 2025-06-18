package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// FileProcessResult 文件处理结果
type FileProcessResult struct {
	Entry      FileEntry
	Processed  *ProcessedFile
	FileType   FileType
	Success    bool
	IsSubtitle bool
	IsMetadata bool
}

// FileProcessQueue 文件处理队列类型
type FileProcessQueue struct {
	StrmFiles     []FileEntry  // 用于生成 STRM 的媒体文件队列
	DownloadFiles []FileEntry  // 需要下载的文件队列 (字幕、元数据)
	FilesMutex    sync.RWMutex // 用于安全访问队列的互斥锁
}

// ProcessingStats 文件处理统计信息
type ProcessingStats struct {
	TotalFiles             int          // 扫描到的总文件数
	GeneratedFiles         int          // 成功生成的 STRM 文件
	SkippedFiles           int          // 跳过的文件
	MetadataProcessed      int          // 处理的元数据文件
	SubtitleProcessed      int          // 处理的字幕文件
	ScanFinished           bool         // 目录扫描是否已完成
	StrmProcessingDone     bool         // STRM 文件处理是否已完成
	DownloadProcessingDone bool         // 下载文件处理是否已完成
	Mutex                  sync.RWMutex // 用于安全访问统计的互斥锁
}

// StrmGeneratorService STRM 文件生成服务
type StrmGeneratorService struct {
	alistService *AListService
	logger       *zap.Logger
	mu           sync.RWMutex
	queue        *FileProcessQueue // 文件处理队列
	stats        *ProcessingStats  // 处理统计
}

var (
	strmGeneratorInstance *StrmGeneratorService
	strmGeneratorOnce     sync.Once
)

// GetStrmGeneratorService 获取 STRM 生成服务实例
func GetStrmGeneratorService() *StrmGeneratorService {
	strmGeneratorOnce.Do(func() {
		strmGeneratorInstance = &StrmGeneratorService{
			queue: &FileProcessQueue{
				StrmFiles:     make([]FileEntry, 0),
				DownloadFiles: make([]FileEntry, 0),
			},
			stats: &ProcessingStats{},
		}
	})
	return strmGeneratorInstance
}

// Initialize 初始化服务
func (s *StrmGeneratorService) Initialize(logger *zap.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
	s.alistService = GetAListService()

	// 初始化队列和统计信息
	s.queue = &FileProcessQueue{
		StrmFiles:     make([]FileEntry, 0),
		DownloadFiles: make([]FileEntry, 0),
	}
	s.stats = &ProcessingStats{}

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

	// 重置处理队列和统计信息
	s.queue = &FileProcessQueue{
		StrmFiles:     make([]FileEntry, 0),
		DownloadFiles: make([]FileEntry, 0),
	}
	s.stats = &ProcessingStats{}

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

	// 先启动STRM文件处理协程（并发），让它等待队列中的项目
	var strmProcessingErr error
	var wg sync.WaitGroup
	wg.Add(1)

	// 创建信号通道，用于向STRM处理协程通知扫描完成
	strmScanDoneChan := make(chan bool)

	// 1. 先启动STRM文件生成协程（并发），它会立即开始处理媒体文件
	go func() {
		defer wg.Done()
		strmProcessingErr = s.processStrmFileQueueAsync(taskInfo, strmConfig, taskLogID, strmScanDoneChan)
	}()

	// 现在开始递归扫描，边扫描边将媒体文件加入队列（立即处理）
	startTime := time.Now()
	err = s.scanDirectoryRecursive(taskInfo, strmConfig, taskLogID, taskInfo.SourcePath, taskInfo.TargetPath)
	if err != nil {
		// 通知STRM协程扫描已结束（失败）
		close(strmScanDoneChan)

		// 等待STRM协程结束
		wg.Wait()

		s.updateTaskLogWithError(taskLogID, "扫描目录失败: "+err.Error())
		return err
	}
	scanDuration := time.Since(startTime)

	// 目录扫描完成后，标记扫描结束，并更新任务日志中的总文件数
	s.stats.Mutex.Lock()
	s.stats.ScanFinished = true
	totalFiles := s.stats.TotalFiles
	s.stats.Mutex.Unlock()

	s.logger.Info("目录扫描完成",
		zap.Int("总文件数", totalFiles),
		zap.Duration("扫描用时", scanDuration),
		zap.Int("STRM队列长度", len(s.queue.StrmFiles)),
		zap.Int("下载队列长度", len(s.queue.DownloadFiles)))

	// 更新任务日志中的总文件数
	updateTotalData := map[string]interface{}{
		"total_file": totalFiles,
	}
	if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateTotalData); updateErr != nil {
		s.logger.Error("更新任务日志总文件数失败", zap.Error(updateErr))
	}

	// 通知STRM协程扫描已结束
	close(strmScanDoneChan)

	// 启动下载文件处理协程（串行）
	var downloadProcessingErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		downloadProcessingErr = s.processDownloadFileQueue(taskInfo, strmConfig, taskLogID)
	}()

	// 等待所有处理都完成
	wg.Wait()

	// 更新任务日志
	s.stats.Mutex.RLock()
	generatedFiles := s.stats.GeneratedFiles
	skippedFiles := s.stats.SkippedFiles
	metadataFiles := s.stats.MetadataProcessed
	subtitleFiles := s.stats.SubtitleProcessed
	s.stats.Mutex.RUnlock()

	endTime := time.Now()
	status := tasklog.TaskLogStatusCompleted
	message := "STRM 文件生成完成"

	// 如果任一处理出错，标记任务失败
	if strmProcessingErr != nil {
		status = tasklog.TaskLogStatusFailed
		message = "STRM 文件生成失败: " + strmProcessingErr.Error()
		err = strmProcessingErr
	} else if downloadProcessingErr != nil {
		status = tasklog.TaskLogStatusFailed
		message = "下载文件处理失败: " + downloadProcessingErr.Error()
		err = downloadProcessingErr
	}

	// 获取当前任务日志记录以获取开始时间
	taskLogRecord, logErr := repository.TaskLog.GetByID(taskLogID)

	// 计算持续时间（秒）
	var durationSeconds int64 = 0
	if logErr == nil {
		durationSeconds = int64(endTime.Sub(taskLogRecord.StartTime).Seconds())
		s.logger.Info("计算任务持续时间",
			zap.Uint("taskLogID", taskLogID),
			zap.Int64("duration", durationSeconds),
			zap.String("taskName", taskInfo.Name))
	} else {
		s.logger.Warn("无法获取任务日志记录，无法计算持续时间",
			zap.Error(logErr),
			zap.Uint("taskLogID", taskLogID))
	}

	updateData := map[string]interface{}{
		"status":         status,
		"message":        message,
		"end_time":       &endTime,
		"duration":       durationSeconds,
		"total_file":     totalFiles,
		"generated_file": generatedFiles,
		"skip_file":      skippedFiles,
		"metadata_count": metadataFiles,
		"subtitle_count": subtitleFiles,
	}

	if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
		s.logger.Error("更新任务日志失败", zap.Error(updateErr))
	}

	// 发送Telegram通知
	notifyErr := s.sendTelegramNotification(taskInfo, taskLogID, status, durationSeconds, updateData)
	if notifyErr != nil {
		s.logger.Error("发送Telegram通知失败", zap.Error(notifyErr))
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

// FileEntry 文件条目，包含完整信息
type FileEntry struct {
	File           *AListFile
	FileType       FileType
	SourcePath     string
	TargetPath     string
	NameWithoutExt string // 不含扩展名的文件名
}

// processDirectory 方法已被重构，使用了新的任务队列设计

// scanDirectoryRecursive 递归扫描目录，只收集文件信息，不进行处理
func (s *StrmGeneratorService) scanDirectoryRecursive(taskInfo *task.Task, strmConfig *StrmConfig,
	taskLogID uint, sourcePath, targetPath string) error {

	// 获取当前目录的文件列表
	files, err := s.alistService.ListFiles(sourcePath)
	if err != nil {
		return fmt.Errorf("获取目录文件列表失败 [%s]: %w", sourcePath, err)
	}

	s.logger.Info("扫描目录",
		zap.String("sourcePath", sourcePath),
		zap.String("targetPath", targetPath),
		zap.Int("fileCount", len(files)))

	// 创建目标目录
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败 [%s]: %w", targetPath, err)
	}

	// 收集各种文件信息
	var mediaFileEntries []FileEntry
	var subtitleFileEntries []FileEntry
	var metadataFileEntries []FileEntry
	var directoryFiles []*AListFile

	// 增加总文件计数
	s.stats.Mutex.Lock()
	s.stats.TotalFiles += len(files)
	totalFiles := s.stats.TotalFiles
	s.stats.Mutex.Unlock()

	// 定期更新任务日志中的总文件数
	if totalFiles%100 == 0 {
		updateData := map[string]interface{}{
			"total_file": totalFiles,
		}
		if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
			s.logger.Error("更新任务日志总文件数失败", zap.Error(updateErr))
		}
	}

	// 先对文件进行分类
	for _, file := range files {
		if file.IsDir {
			directoryFiles = append(directoryFiles, &file)
			continue
		}

		// 构建完整路径
		currentSourcePath := filepath.Join(sourcePath, file.Name)
		currentTargetPath := filepath.Join(targetPath, file.Name)

		// 确定文件类型
		fileType := s.determineFileType(&file, taskInfo, strmConfig)

		// 获取不含扩展名的文件名
		nameWithoutExt := strings.TrimSuffix(file.Name, filepath.Ext(file.Name))

		entry := FileEntry{
			File:           &file,
			FileType:       fileType,
			SourcePath:     currentSourcePath,
			TargetPath:     currentTargetPath,
			NameWithoutExt: nameWithoutExt,
		}

		// 按类型分组
		switch fileType {
		case FileTypeMedia:
			mediaFileEntries = append(mediaFileEntries, entry)
		case FileTypeSubtitle:
			subtitleFileEntries = append(subtitleFileEntries, entry)
		case FileTypeMetadata:
			metadataFileEntries = append(metadataFileEntries, entry)
		default:
			// 其他文件类型不处理，但计入跳过文件
			s.stats.Mutex.Lock()
			s.stats.SkippedFiles++
			s.stats.Mutex.Unlock()
		}
	}

	// 筛选需要处理的字幕文件（需要与媒体文件匹配）
	var matchedSubtitleEntries []FileEntry
	for _, subEntry := range subtitleFileEntries {
		matched := false

		// 检查是否与任何媒体文件匹配
		for _, mediaEntry := range mediaFileEntries {
			// 字幕文件名需要以媒体文件名为前缀（如movie.mp4与movie.srt）
			if strings.HasPrefix(subEntry.NameWithoutExt, mediaEntry.NameWithoutExt) {
				matched = true
				break
			}
		}

		if matched {
			matchedSubtitleEntries = append(matchedSubtitleEntries, subEntry)
		} else {
			s.logger.Info("跳过未匹配的字幕文件",
				zap.String("fileName", subEntry.File.Name),
				zap.String("path", subEntry.SourcePath))

			s.stats.Mutex.Lock()
			s.stats.SkippedFiles++
			s.stats.Mutex.Unlock()
		}
	}

	// 将收集到的文件添加到相应的处理队列
	// 媒体文件立即添加到STRM生成队列，这样边扫描边处理
	if len(mediaFileEntries) > 0 {
		s.queue.FilesMutex.Lock()
		// 添加媒体文件到 STRM 生成队列
		s.queue.StrmFiles = append(s.queue.StrmFiles, mediaFileEntries...)
		s.queue.FilesMutex.Unlock()
	}

	// 下载文件先收集，等扫描结束后再处理
	if len(matchedSubtitleEntries) > 0 || len(metadataFileEntries) > 0 {
		s.queue.FilesMutex.Lock()
		// 添加匹配的字幕和元数据文件到下载队列
		s.queue.DownloadFiles = append(s.queue.DownloadFiles, matchedSubtitleEntries...)
		s.queue.DownloadFiles = append(s.queue.DownloadFiles, metadataFileEntries...)
		s.queue.FilesMutex.Unlock()
	}

	// 递归处理子目录
	for _, dirFile := range directoryFiles {
		currentSourcePath := filepath.Join(sourcePath, dirFile.Name)
		currentTargetPath := filepath.Join(targetPath, dirFile.Name)

		// 递归处理子目录
		if err := s.scanDirectoryRecursive(taskInfo, strmConfig, taskLogID, currentSourcePath, currentTargetPath); err != nil {
			return err
		}
	}

	return nil
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
		var strmFilePath string
		result.Success, result.ErrorMessage, strmFilePath = s.generateStrmFile(file, strmConfig, taskInfo, sourcePath, targetPath)
		if result.Success {
			// 如果成功生成STRM文件，更新目标路径为实际的STRM文件路径
			result.TargetPath = strmFilePath
		}
	case FileTypeMetadata, FileTypeSubtitle:
		// 下载元数据或字幕文件 - 仅使用 AListFile 中已有信息
		result.Success, result.ErrorMessage = s.downloadFile(file, sourcePath, targetPath, taskInfo)
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

// generateStrmFile 生成 STRM 文件，返回成功状态、错误消息和STRM文件路径
func (s *StrmGeneratorService) generateStrmFile(file *AListFile, strmConfig *StrmConfig, taskConfig *task.Task, sourcePath, targetPath string) (bool, string, string) {
	// 处理路径和文件名的 URL 编码
	dirPath := filepath.Dir(sourcePath)
	fileName := file.Name

	// 根据 URLEncode 配置决定是否需要对路径进行编码
	if strmConfig.URLEncode {
		// 将路径分割为各个部分，对每部分进行单独编码，然后重新连接
		// 这与 Node.js 版本的处理方式相同: path.split('/').map(encodeURIComponent).join('/')
		pathParts := strings.Split(dirPath, "/")
		for i, part := range pathParts {
			pathParts[i] = url.PathEscape(part)
		}
		dirPath = strings.Join(pathParts, "/")

		// 同样处理文件名
		fileName = url.PathEscape(fileName)

		s.logger.Debug("进行了URL编码",
			zap.String("原路径", filepath.Dir(sourcePath)),
			zap.String("编码后路径", dirPath),
			zap.String("原文件名", file.Name),
			zap.String("编码后文件名", fileName))
	}

	// 构建 STRM 文件内容 - 直接使用 AListFile 中的信息，避免多余的 API 调用
	// 注意：GetFileURL 方法不会发起额外的 API 请求，仅使用配置和参数构建 URL
	fileURL := s.alistService.GetFileURL(dirPath, fileName, file.Sign)
	if fileURL == "" {
		return false, "无法生成文件URL，请检查 AList 配置是否完整", ""
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
		return false, "文件已存在且不允许覆盖", strmFilePath
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(strmFilePath), 0755); err != nil {
		return false, fmt.Sprintf("创建目标目录失败: %v", err), strmFilePath
	}

	// 写入 STRM 文件
	if err := os.WriteFile(strmFilePath, []byte(fileURL), 0644); err != nil {
		return false, fmt.Sprintf("写入 STRM 文件失败: %v", err), strmFilePath
	}

	s.logger.Info("生成 STRM 文件成功",
		zap.String("sourceFile", file.Name),
		zap.String("strmFile", strmFilePath),
		zap.String("url", fileURL))

	return true, "", strmFilePath
}

// downloadFile 下载文件（元数据和字幕）
func (s *StrmGeneratorService) downloadFile(file *AListFile, sourcePath, targetPath string, taskConfig *task.Task) (bool, string) {

	if _, err := os.Stat(targetPath); err == nil {
		return false, "文件已存在"
	}

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return false, fmt.Sprintf("创建目标目录失败: %v", err)
	}

	// 获取 STRM 配置以检查是否需要 URL 编码
	strmConfig, err := s.loadStrmConfig()
	if err != nil {
		return false, fmt.Sprintf("加载 STRM 配置失败: %v", err)
	}

	// 处理路径和文件名
	dirPath := filepath.Dir(sourcePath)
	fileName := file.Name

	// 根据 URLEncode 配置决定是否需要对路径进行编码
	if strmConfig.URLEncode {
		// 对路径各部分单独编码
		pathParts := strings.Split(dirPath, "/")
		for i, part := range pathParts {
			pathParts[i] = url.PathEscape(part)
		}
		dirPath = strings.Join(pathParts, "/")

		// 对文件名编码
		fileName = url.PathEscape(fileName)
	}

	// 直接使用 AListFile 中的信息构建文件 URL，不需要额外的 API 调用
	// 注意：GetFileURL 方法不会发起额外的 API 请求，仅使用配置和参数构建 URL
	fileURL := s.alistService.GetFileURL(dirPath, fileName, file.Sign)
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

	// 获取文件Hash
	hash := ""
	if file.HashInfo.Sha1 != "" {
		hash = file.HashInfo.Sha1
	}

	// 记录文件类型和是否有Hash值
	s.logger.Debug("处理文件历史记录",
		zap.String("fileName", file.Name),
		zap.String("fileType", fileTypeStr),
		zap.Bool("hasHash", hash != ""),
		zap.String("hash", hash))

	// 根据Hash查找现有记录（只有当Hash有值时才查找）
	if hash != "" {
		existingRecord, err := repository.FileHistory.GetByHash(hash)
		if err == nil && existingRecord != nil {
			// 找到现有记录，更新而不是创建
			now := time.Now()
			updateData := map[string]interface{}{
				"task_id":     taskID,
				"task_log_id": taskLogID,
				"updated_at":  now,
				"file_size":   file.Size,
				"modified_at": &file.Modified,
			}

			if err := repository.FileHistory.UpdateByID(existingRecord.ID, updateData); err != nil {
				s.logger.Error("更新文件历史记录失败",
					zap.String("fileName", file.Name),
					zap.String("hash", hash),
					zap.Error(err))
			} else {
				s.logger.Info("更新现有文件历史记录",
					zap.String("fileName", file.Name),
					zap.String("hash", hash),
					zap.String("fileType", fileTypeStr),
					zap.Uint("oldTaskID", existingRecord.TaskID),
					zap.Uint("newTaskID", taskID))
			}
			return
		}
	}

	// 没有找到现有记录或者没有Hash，创建新记录
	fileHistory := &filehistory.FileHistory{
		TaskID:         taskID,
		TaskLogID:      taskLogID,
		FileName:       file.Name,
		SourcePath:     sourcePath,
		TargetFilePath: targetPath,
		FileSize:       file.Size,
		FileType:       fileTypeStr,
		FileSuffix:     filepath.Ext(file.Name),
		IsStrm:         fileType == FileTypeMedia, // 如果是媒体文件类型，则标记为STRM文件
		ModifiedAt:     &file.Modified,
		Hash:           hash,
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
	// 获取当前任务日志记录以获取开始时间
	taskLog, err := repository.TaskLog.GetByID(taskLogID)
	if err != nil {
		s.logger.Error("获取任务日志失败", zap.Error(err))
		return
	}

	// 计算持续时间（秒）
	endTime := time.Now()
	durationSeconds := int64(endTime.Sub(taskLog.StartTime).Seconds())

	updateData := map[string]interface{}{
		"status":   tasklog.TaskLogStatusFailed,
		"message":  errorMessage,
		"end_time": &endTime,
		"duration": durationSeconds,
	}

	if err := repository.TaskLog.UpdatePartial(taskLogID, updateData); err != nil {
		s.logger.Error("更新任务日志失败", zap.Error(err))
	} else {
		s.logger.Debug("已更新任务日志持续时间",
			zap.Uint("taskLogID", taskLogID),
			zap.Int64("duration", durationSeconds))
	}
}

// sendTelegramNotification 发送Telegram通知
func (s *StrmGeneratorService) sendTelegramNotification(taskInfo *task.Task, taskLogID uint, status string, duration int64, stats map[string]interface{}) error {
	// Telegram Bot配置（实际应用中应从配置文件或环境变量中读取）
	const (
		enableNotification = true                                                 // 是否启用Telegram通知
		botToken           = "467857328346:AAGwEGenWJYec1irqG26wJMoWxQHs6HArC0eE" // 替换为你的Telegram Bot Token
		chatID             = "5486452678413"                                      // 替换为你的Chat ID (可以是数字或者字符串，如"@channelname")
	)

	// 如果未启用通知或未配置Token/ChatID，则跳过发送
	if !enableNotification || botToken == "YOUR_BOT_TOKEN_HERE" || chatID == "YOUR_CHAT_ID_HERE" {
		s.logger.Info("Telegram通知未启用或未配置完成，跳过发送")
		return nil
	}
	// 构建消息文本
	statusEmoji := "✅"
	errorInfo := ""

	if status != tasklog.TaskLogStatusCompleted {
		statusEmoji = "❌"
		// 如果是失败状态，尝试获取错误消息
		if strings.Contains(status, "失败") {
			errorInfo = strings.TrimPrefix(status, "STRM 文件生成失败: ")
			errorInfo = strings.TrimPrefix(errorInfo, "下载文件处理失败: ")
		}
	}

	// 格式化时间
	durationStr := ""
	if duration > 0 {
		hours := duration / 3600
		minutes := (duration % 3600) / 60
		seconds := duration % 60

		if hours > 0 {
			durationStr = fmt.Sprintf("%d小时%d分钟%d秒", hours, minutes, seconds)
		} else if minutes > 0 {
			durationStr = fmt.Sprintf("%d分钟%d秒", minutes, seconds)
		} else {
			durationStr = fmt.Sprintf("%d秒", seconds)
		}
	}

	// 提取统计数据，确保安全转换
	var (
		totalFiles     int
		generatedFiles int
		skippedFiles   int
		metadataFiles  int
		subtitleFiles  int
	)

	// 类型安全的转换
	if v, ok := stats["total_file"].(int); ok {
		totalFiles = v
	}
	if v, ok := stats["generated_file"].(int); ok {
		generatedFiles = v
	}
	if v, ok := stats["skip_file"].(int); ok {
		skippedFiles = v
	}
	if v, ok := stats["metadata_count"].(int); ok {
		metadataFiles = v
	}
	if v, ok := stats["subtitle_count"].(int); ok {
		subtitleFiles = v
	}

	// 构建消息
	// 为状态文本设置更友好的显示
	statusDisplay := "成功"
	if status != tasklog.TaskLogStatusCompleted {
		statusDisplay = "失败"
	}

	// 添加错误信息部分(如果有)
	errorPart := ""
	if errorInfo != "" {
		errorPart = fmt.Sprintf("\n\n⚠️ *错误信息*\n`%s`", errorInfo)
	}

	message := fmt.Sprintf("🎬 *AList2Strm 任务通知* %s\n\n"+
		"📋 *任务详情*\n"+
		"├ 名称: `%s`\n"+
		"├ 状态: %s *%s*\n"+
		"└ 用时: `%s`\n\n"+
		"📊 *文件统计*\n"+
		"├ 总文件: `%d` 个\n"+
		"├ 已生成: `%d` 个\n"+
		"├ 已跳过: `%d` 个\n"+
		"├ 元数据: `%d` 个\n"+
		"└ 字幕: `%d` 个\n\n"+
		"📁 *路径信息*\n"+
		"├ 源路径: `%s`\n"+
		"└ 目标路径: `%s`%s",
		statusEmoji,
		taskInfo.Name,
		statusEmoji,
		statusDisplay,
		durationStr,
		totalFiles,
		generatedFiles,
		skippedFiles,
		metadataFiles,
		subtitleFiles,
		taskInfo.SourcePath,
		taskInfo.TargetPath,
		errorPart)

	// URL编码消息
	encodedMessage := url.QueryEscape(message)

	// 构建API URL (使用Markdown解析模式)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s&parse_mode=Markdown",
		botToken, chatID, encodedMessage)

	// 发送HTTP请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		s.logger.Error("发送Telegram通知失败", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("Telegram API返回错误",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)))
		return fmt.Errorf("telegram API返回错误: %d", resp.StatusCode)
	}

	s.logger.Info("Telegram通知发送成功",
		zap.String("taskName", taskInfo.Name),
		zap.String("status", status))

	return nil
}

// processStrmFileQueue 处理 STRM 文件生成队列（并发处理）
// 已弃用，请使用 processStrmFileQueueAsync
// nolint:unused
// 保留此方法是为了兼容性
func (s *StrmGeneratorService) processStrmFileQueue(taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint) error {
	s.queue.FilesMutex.RLock()
	totalStrmFiles := len(s.queue.StrmFiles)
	s.queue.FilesMutex.RUnlock()

	if totalStrmFiles == 0 {
		s.logger.Info("没有媒体文件需要生成 STRM")
		s.stats.Mutex.Lock()
		s.stats.StrmProcessingDone = true
		s.stats.Mutex.Unlock()
		return nil
	}

	s.logger.Info("开始并发生成STRM文件",
		zap.Int("媒体文件数", totalStrmFiles))

	// 设置并发数
	concurrency := 100 // 可以根据需要调整，或从配置中读取
	if concurrency <= 0 {
		concurrency = 100
	}
	if concurrency > totalStrmFiles {
		concurrency = totalStrmFiles
	}

	// 创建任务和结果通道
	jobChan := make(chan FileEntry, totalStrmFiles)
	resultChan := make(chan FileProcessResult, totalStrmFiles)

	// 启动工作协程池
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for entry := range jobChan {
				// 只处理媒体文件，生成STRM文件
				processed := s.processFile(entry.File, entry.FileType, taskInfo, strmConfig, taskLogID, entry.SourcePath, entry.TargetPath)

				// 发送结果
				resultChan <- FileProcessResult{
					Entry:     entry,
					Processed: processed,
					FileType:  entry.FileType,
					Success:   processed.Success,
				}
			}
		}()
	}

	// 复制 STRM 文件队列以避免锁冲突
	s.queue.FilesMutex.RLock()
	strmFiles := make([]FileEntry, len(s.queue.StrmFiles))
	copy(strmFiles, s.queue.StrmFiles)
	s.queue.FilesMutex.RUnlock()

	// 提交任务
	go func() {
		// 提交媒体文件
		for _, entry := range strmFiles {
			jobChan <- entry
		}

		// 关闭任务通道，表示没有更多任务
		close(jobChan)

		// 等待所有工作协程完成
		wg.Wait()

		// 关闭结果通道
		close(resultChan)
	}()

	// 收集处理结果
	for result := range resultChan {
		// 确定使用的目标路径
		targetPath := result.Processed.TargetPath

		// 记录文件历史
		s.recordFileHistory(
			taskInfo.ID,
			taskLogID,
			result.Entry.File,
			result.Entry.SourcePath,
			targetPath,
			result.FileType,
			result.Success,
		)

		// 统计结果
		s.stats.Mutex.Lock()
		if result.Success {
			s.stats.GeneratedFiles++
		} else {
			s.stats.SkippedFiles++
		}
		s.stats.Mutex.Unlock()

		// 定期更新任务日志
		if s.stats.GeneratedFiles%100 == 0 || s.stats.SkippedFiles%100 == 0 {
			s.stats.Mutex.RLock()
			updateData := map[string]interface{}{
				"generated_file": s.stats.GeneratedFiles,
				"skip_file":      s.stats.SkippedFiles,
			}
			s.stats.Mutex.RUnlock()

			if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
				s.logger.Error("更新任务日志进度失败", zap.Error(updateErr))
			}
		}
	}

	// 标记 STRM 处理完成
	s.stats.Mutex.Lock()
	s.stats.StrmProcessingDone = true
	s.stats.Mutex.Unlock()

	s.logger.Info("STRM 文件生成队列处理完成",
		zap.Int("生成文件数", s.stats.GeneratedFiles))

	return nil
}

// processDownloadFileQueue 处理下载文件队列（串行处理，带延迟）
func (s *StrmGeneratorService) processDownloadFileQueue(taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint) error {
	s.queue.FilesMutex.RLock()
	totalDownloadFiles := len(s.queue.DownloadFiles)
	s.queue.FilesMutex.RUnlock()

	if totalDownloadFiles == 0 {
		s.logger.Info("没有文件需要下载")
		s.stats.Mutex.Lock()
		s.stats.DownloadProcessingDone = true
		s.stats.Mutex.Unlock()
		return nil
	}

	s.logger.Info("开始串行处理下载任务",
		zap.Int("下载文件总数", totalDownloadFiles))

	// 复制下载队列以避免锁冲突
	s.queue.FilesMutex.RLock()
	downloadFiles := make([]FileEntry, len(s.queue.DownloadFiles))
	copy(downloadFiles, s.queue.DownloadFiles)
	s.queue.FilesMutex.RUnlock()

	// 串行处理每个下载项，带间隔延迟
	for i, entry := range downloadFiles {
		// 添加随机延迟(1-3秒)，防止网盘风控
		if i > 0 {
			// 设置随机延迟，更好地模拟人工操作
			randomDelay := time.Duration(1000+(time.Now().UnixNano()%2000)) * time.Millisecond
			s.logger.Info("等待随机延迟", zap.Duration("delay", randomDelay))
			time.Sleep(randomDelay)
		}

		// 处理文件
		processed := s.processFile(entry.File, entry.FileType, taskInfo, strmConfig, taskLogID, entry.SourcePath, entry.TargetPath)

		// 记录文件历史
		s.recordFileHistory(taskInfo.ID, taskLogID, entry.File, entry.SourcePath, processed.TargetPath, entry.FileType, processed.Success)

		// 更新统计信息
		s.stats.Mutex.Lock()
		if processed.Success {
			if entry.FileType == FileTypeSubtitle {
				s.stats.SubtitleProcessed++
			} else if entry.FileType == FileTypeMetadata {
				s.stats.MetadataProcessed++
			}
		} else {
			s.stats.SkippedFiles++
		}
		s.stats.Mutex.Unlock()

		// 每处理 10 个文件更新一次数据库
		if (i+1)%10 == 0 {
			s.stats.Mutex.RLock()
			updateData := map[string]interface{}{
				"subtitle_count": s.stats.SubtitleProcessed,
				"metadata_count": s.stats.MetadataProcessed,
				"skip_file":      s.stats.SkippedFiles,
			}
			s.stats.Mutex.RUnlock()

			if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
				s.logger.Error("更新任务日志进度失败", zap.Error(updateErr))
			}

			s.logger.Info("下载队列处理进度",
				zap.Int("已处理", i+1),
				zap.Int("总数", totalDownloadFiles),
				zap.Int("字幕文件", s.stats.SubtitleProcessed),
				zap.Int("元数据文件", s.stats.MetadataProcessed))
		}
	}

	// 标记下载处理完成
	s.stats.Mutex.Lock()
	s.stats.DownloadProcessingDone = true
	s.stats.Mutex.Unlock()

	s.logger.Info("下载文件队列处理完成",
		zap.Int("字幕文件", s.stats.SubtitleProcessed),
		zap.Int("元数据文件", s.stats.MetadataProcessed))

	return nil
}

// processStrmFileQueueAsync 异步处理STRM文件队列（并发处理），可以在目录扫描时就开始处理
func (s *StrmGeneratorService) processStrmFileQueueAsync(taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint, scanDoneChan chan bool) error {
	// 设置并发数
	const defaultConcurrency = 50
	// TODO: 未来可考虑从任务配置或全局配置中读取并发参数
	concurrency := defaultConcurrency

	s.logger.Info("启动STRM文件异步处理协程",
		zap.Int("并发数", concurrency))

	// 创建任务和结果通道
	// TODO: 未来可考虑将通道大小设置为可配置参数，避免超大目录处理时内存占用过多
	const channelSize = 1000
	jobChan := make(chan FileEntry, channelSize)            // 缓冲队列，用于接收扫描到的文件
	resultChan := make(chan FileProcessResult, channelSize) // 结果队列

	// 启动工作协程池
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for entry := range jobChan {
				// 处理媒体文件，生成STRM文件
				processed := s.processFile(entry.File, entry.FileType, taskInfo, strmConfig, taskLogID, entry.SourcePath, entry.TargetPath)

				// 发送结果
				resultChan <- FileProcessResult{
					Entry:     entry,
					Processed: processed,
					FileType:  entry.FileType,
					Success:   processed.Success,
				}
			}
		}()
	}

	// 启动结果收集协程
	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for result := range resultChan {
			// 确定使用的目标路径
			targetPath := result.Processed.TargetPath

			// 记录文件历史
			s.recordFileHistory(
				taskInfo.ID,
				taskLogID,
				result.Entry.File,
				result.Entry.SourcePath,
				targetPath,
				result.FileType,
				result.Success,
			)

			// 统计结果
			s.stats.Mutex.Lock()
			if result.Success {
				s.stats.GeneratedFiles++
			} else {
				s.stats.SkippedFiles++
			}
			s.stats.Mutex.Unlock()

			// 定期更新任务日志
			if s.stats.GeneratedFiles%100 == 0 || s.stats.SkippedFiles%100 == 0 {
				s.stats.Mutex.RLock()
				updateData := map[string]interface{}{
					"generated_file": s.stats.GeneratedFiles,
					"skip_file":      s.stats.SkippedFiles,
				}
				s.stats.Mutex.RUnlock()

				if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
					s.logger.Error("更新任务日志进度失败", zap.Error(updateErr))
				}
			}
		}
	}()

	// 启动队列监听协程，将文件发送到工作协程
	go func() {
		scanDone := false

		// 持续监听直到扫描结束且队列为空
		for !scanDone || s.queue.hasStrmFiles() {
			// 先检查是否有文件可处理
			if s.queue.hasStrmFiles() {
				// 获取并删除队列中的一批文件
				batch := s.queue.getAndRemoveStrmFileBatch(100)
				if len(batch) > 0 {
					s.logger.Debug("提交一批STRM文件进行处理", zap.Int("数量", len(batch)))
					for _, entry := range batch {
						jobChan <- entry
					}
				}
			}

			// 检查扫描是否结束
			select {
			case _, ok := <-scanDoneChan:
				if !ok {
					// 通道关闭，扫描结束
					scanDone = true
				}
			default:
				// 通道未关闭，休息一下再检查
				time.Sleep(50 * time.Millisecond)
			}
		}

		// 关闭任务通道，表示没有更多任务
		close(jobChan)

		// 等待所有工作协程完成
		wg.Wait()

		// 关闭结果通道
		close(resultChan)

		// 等待结果收集完成
		resultWg.Wait()
	}()

	// 等待所有处理完成
	resultWg.Wait()

	// 标记 STRM 处理完成
	s.stats.Mutex.Lock()
	s.stats.StrmProcessingDone = true
	s.stats.Mutex.Unlock()

	s.logger.Info("STRM 文件生成队列处理完成",
		zap.Int("生成文件数", s.stats.GeneratedFiles))

	return nil
}

// hasStrmFiles 检查队列中是否有STRM文件
func (q *FileProcessQueue) hasStrmFiles() bool {
	q.FilesMutex.RLock()
	defer q.FilesMutex.RUnlock()
	return len(q.StrmFiles) > 0
}

// getAndRemoveStrmFileBatch 获取并移除一批STRM文件
func (q *FileProcessQueue) getAndRemoveStrmFileBatch(batchSize int) []FileEntry {
	q.FilesMutex.Lock()
	defer q.FilesMutex.Unlock()

	if len(q.StrmFiles) == 0 {
		return []FileEntry{}
	}

	// 确定批次大小
	size := batchSize
	if size > len(q.StrmFiles) {
		size = len(q.StrmFiles)
	}

	// 获取批次
	batch := make([]FileEntry, size)
	copy(batch, q.StrmFiles[:size])

	// 移除已获取的文件
	q.StrmFiles = q.StrmFiles[size:]

	return batch
}
