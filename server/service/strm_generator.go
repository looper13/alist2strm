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
	GeneratedFile          int          // 成功生成的 STRM 文件数 (与 TaskLog 字段保持一致)
	SkipFile               int          // 跳过的 STRM 文件数 (与 TaskLog 字段保持一致)
	OverwriteFile          int          // 覆盖的文件数 (与 TaskLog 字段保持一致)
	MetadataDownloaded     int          // 已下载的元数据文件数
	MetadataSkipped        int          // 已跳过的元数据文件数
	SubtitleDownloaded     int          // 已下载的字幕文件数
	SubtitleSkipped        int          // 已跳过的字幕文件数
	OtherSkipped           int          // 跳过的其他类型文件数
	FailedCount            int          // 处理失败的文件数 (与 TaskLog 字段保持一致)
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

	// 启动下载文件处理协程（串行）- 仅在队列有数据时启动
	var downloadProcessingErr error
	s.queue.FilesMutex.RLock()
	hasDownloadFiles := len(s.queue.DownloadFiles) > 0
	s.queue.FilesMutex.RUnlock()

	if hasDownloadFiles {
		wg.Add(1)
		go func() {
			defer wg.Done()
			downloadProcessingErr = s.processDownloadFileQueue(taskInfo, strmConfig, taskLogID)
		}()
	} else {
		// 无需下载，直接标记下载处理完成
		s.stats.Mutex.Lock()
		s.stats.DownloadProcessingDone = true
		s.stats.Mutex.Unlock()
		s.logger.Info("下载队列为空，跳过下载处理")
	}

	// 等待所有处理都完成
	wg.Wait()

	// 更新任务日志
	s.stats.Mutex.RLock()
	// 计算统计数据
	generatedFiles := s.stats.GeneratedFile
	// 所有跳过的文件总和：STRM文件跳过 + 元数据文件跳过 + 字幕文件跳过 + 其他文件跳过
	skippedFiles := s.stats.SkipFile + s.stats.MetadataSkipped + s.stats.SubtitleSkipped + s.stats.OtherSkipped
	// 元数据处理总数：下载 + 跳过
	metadataFiles := s.stats.MetadataDownloaded + s.stats.MetadataSkipped
	// 字幕处理总数：下载 + 跳过
	subtitleFiles := s.stats.SubtitleDownloaded + s.stats.SubtitleSkipped
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

	s.stats.Mutex.RLock()
	metadataDownloaded := s.stats.MetadataDownloaded
	metadataSkipped := s.stats.MetadataSkipped
	subtitleDownloaded := s.stats.SubtitleDownloaded
	subtitleSkipped := s.stats.SubtitleSkipped
	otherSkipped := s.stats.OtherSkipped
	failedCount := s.stats.FailedCount
	s.stats.Mutex.RUnlock()

	// 只包含 TaskLog 模型中存在的字段
	updateData := map[string]interface{}{
		"status":              status,
		"message":             message,
		"end_time":            &endTime,
		"duration":            durationSeconds,
		"total_file":          totalFiles,
		"generated_file":      generatedFiles,
		"skip_file":           skippedFiles,
		"metadata_count":      metadataFiles,
		"subtitle_count":      subtitleFiles,
		"metadata_downloaded": metadataDownloaded,
		"subtitle_downloaded": subtitleDownloaded,
		"failed_count":        failedCount,
	}

	// 额外的统计信息保留在通知中，但不更新到数据库
	notifyData := map[string]interface{}{
		"status":              status,
		"message":             message,
		"end_time":            &endTime,
		"duration":            durationSeconds,
		"total_file":          totalFiles,
		"generated_file":      generatedFiles,
		"skip_file":           skippedFiles,
		"metadata_count":      metadataFiles,
		"subtitle_count":      subtitleFiles,
		"metadata_downloaded": metadataDownloaded,
		"subtitle_downloaded": subtitleDownloaded,
		"metadata_skipped":    metadataSkipped,
		"subtitle_skipped":    subtitleSkipped,
		"other_skipped":       otherSkipped,
		"failed_count":        failedCount,
	}

	if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
		s.logger.Error("更新任务日志失败", zap.Error(updateErr))
	}

	// 发送通知 - 使用包含额外详细统计信息的notifyData
	notifyErr := s.sendNotification(taskInfo, taskLogID, status, durationSeconds, notifyData)
	if notifyErr != nil {
		s.logger.Error("发送任务通知失败", zap.Error(notifyErr))
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
			s.stats.OtherSkipped++
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
			s.stats.SubtitleSkipped++ // 将未匹配的字幕文件计入已跳过的字幕文件
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
	// 先筛选出需要下载的文件（本地不存在的文件）
	var needDownloadEntries []FileEntry

	// 检查匹配的字幕文件是否已存在于本地
	for _, entry := range matchedSubtitleEntries {
		if !s.fileExistsLocally(entry.TargetPath) {
			needDownloadEntries = append(needDownloadEntries, entry)
		} else {
			s.logger.Info("字幕文件已存在，跳过下载",
				zap.String("fileName", entry.File.Name),
				zap.String("targetPath", entry.TargetPath))

			// 记录文件历史（已存在的文件）
			s.recordFileHistory(taskInfo.ID, taskLogID, entry.File, entry.SourcePath, entry.TargetPath, entry.FileType, true)

			// 更新统计信息
			s.stats.Mutex.Lock()
			s.stats.SubtitleSkipped++ // 计入已跳过的字幕文件
			s.stats.Mutex.Unlock()
		}
	}

	// 检查元数据文件是否已存在于本地
	for _, entry := range metadataFileEntries {
		if !s.fileExistsLocally(entry.TargetPath) {
			needDownloadEntries = append(needDownloadEntries, entry)
		} else {
			s.logger.Info("元数据文件已存在，跳过下载",
				zap.String("fileName", entry.File.Name),
				zap.String("targetPath", entry.TargetPath))

			// 记录文件历史（已存在的文件）
			s.recordFileHistory(taskInfo.ID, taskLogID, entry.File, entry.SourcePath, entry.TargetPath, entry.FileType, true)

			// 更新统计信息
			s.stats.Mutex.Lock()
			s.stats.MetadataSkipped++ // 计入已跳过的元数据文件
			s.stats.Mutex.Unlock()
		}
	}

	// 只将需要下载的文件添加到下载队列
	if len(needDownloadEntries) > 0 {
		s.queue.FilesMutex.Lock()
		s.queue.DownloadFiles = append(s.queue.DownloadFiles, needDownloadEntries...)
		s.queue.FilesMutex.Unlock()

		// 获取已跳过的文件数（已包含在总跳过文件数中）
		skippedExistingFiles := len(matchedSubtitleEntries) + len(metadataFileEntries) - len(needDownloadEntries)

		s.stats.Mutex.RLock()
		s.logger.Info("添加文件到下载队列",
			zap.Int("需下载文件数", len(needDownloadEntries)),
			zap.Int("跳过已存在文件数", skippedExistingFiles),
			zap.Int("已跳过字幕文件", s.stats.SubtitleSkipped),
			zap.Int("已跳过元数据文件", s.stats.MetadataSkipped),
			zap.Int("跳过其他文件数", s.stats.OtherSkipped))
		s.stats.Mutex.RUnlock()
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
		zap.Int("需下载文件总数", totalDownloadFiles))

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
				s.stats.SubtitleDownloaded++ // 成功下载的字幕文件
			} else if entry.FileType == FileTypeMetadata {
				s.stats.MetadataDownloaded++ // 成功下载的元数据文件
			}
		} else {
			s.stats.FailedCount++ // 处理失败的文件
			// 下载失败的文件也应计入相应的跳过类别
			if entry.FileType == FileTypeSubtitle {
				s.stats.SubtitleSkipped++ // 下载失败的字幕文件计入已跳过
			} else if entry.FileType == FileTypeMetadata {
				s.stats.MetadataSkipped++ // 下载失败的元数据文件计入已跳过
			}
		}
		s.stats.Mutex.Unlock()

		// 每处理 10 个文件更新一次数据库
		if (i+1)%10 == 0 {
			s.stats.Mutex.RLock()
			// 计算数据库中需要的汇总数值
			subtitleCount := s.stats.SubtitleDownloaded + s.stats.SubtitleSkipped
			metadataCount := s.stats.MetadataDownloaded + s.stats.MetadataSkipped
			skipFileCount := s.stats.SkipFile + s.stats.MetadataSkipped + s.stats.SubtitleSkipped + s.stats.OtherSkipped

			updateData := map[string]interface{}{
				"subtitle_count": subtitleCount,
				"metadata_count": metadataCount,
				"skip_file":      skipFileCount,
			}
			s.stats.Mutex.RUnlock()

			if updateErr := repository.TaskLog.UpdatePartial(taskLogID, updateData); updateErr != nil {
				s.logger.Error("更新任务日志进度失败", zap.Error(updateErr))
			}

			s.stats.Mutex.RLock()
			s.logger.Info("下载队列处理进度",
				zap.Int("已处理", i+1),
				zap.Int("总数", totalDownloadFiles),
				zap.Int("已下载字幕", s.stats.SubtitleDownloaded),
				zap.Int("已跳过字幕", s.stats.SubtitleSkipped),
				zap.Int("已下载元数据", s.stats.MetadataDownloaded),
				zap.Int("已跳过元数据", s.stats.MetadataSkipped))
			s.stats.Mutex.RUnlock()
		}
	}

	// 标记下载处理完成
	s.stats.Mutex.Lock()
	s.stats.DownloadProcessingDone = true
	s.stats.Mutex.Unlock()

	s.stats.Mutex.RLock()
	s.logger.Info("下载文件队列处理完成",
		zap.Int("已下载字幕", s.stats.SubtitleDownloaded),
		zap.Int("已跳过字幕", s.stats.SubtitleSkipped),
		zap.Int("已下载元数据", s.stats.MetadataDownloaded),
		zap.Int("已跳过元数据", s.stats.MetadataSkipped))
	s.stats.Mutex.RUnlock()

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
				s.stats.GeneratedFile++ // 成功生成的STRM文件
			} else {
				s.stats.SkipFile++ // 跳过的STRM文件
			}
			s.stats.Mutex.Unlock()

			// 定期更新任务日志
			if s.stats.GeneratedFile%100 == 0 || s.stats.SkipFile%100 == 0 {
				s.stats.Mutex.RLock()
				// 计算需要更新到数据库的总计数
				skipFileCount := s.stats.SkipFile + s.stats.MetadataSkipped + s.stats.SubtitleSkipped + s.stats.OtherSkipped

				updateData := map[string]interface{}{
					"generated_file": s.stats.GeneratedFile,
					"skip_file":      skipFileCount,
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

	s.stats.Mutex.RLock()
	s.logger.Info("STRM 文件生成队列处理完成",
		zap.Int("生成文件数", s.stats.GeneratedFile),
		zap.Int("跳过文件数", s.stats.SkipFile))
	s.stats.Mutex.RUnlock()

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

// fileExistsLocally 检查文件是否已存在于本地
func (s *StrmGeneratorService) fileExistsLocally(filePath string) bool {
	// 使用 os.Stat 检查文件是否存在
	_, err := os.Stat(filePath)
	return err == nil
}
