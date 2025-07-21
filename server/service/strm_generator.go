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
	"github.com/MccRay-s/alist2strm/model/webhook"
	"github.com/MccRay-s/alist2strm/repository"
	"go.uber.org/zap"
)

// StrmConfig STRM 配置结构
type StrmConfig struct {
	DefaultSuffix string `json:"defaultSuffix"` // 默认媒体文件后缀
	ReplaceSuffix bool   `json:"replaceSuffix"` // 是否替换后缀
	URLEncode     bool   `json:"urlEncode"`     // 是否URL编码
	MinFileSize   int64  `json:"minFileSize"`   // 最小文件大小(MB)，用于过滤小文件，0表示不过滤
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
	alistService      *AListService
	cloudDriveService *CloudDriveService
	logger            *zap.Logger
	mu                sync.RWMutex
	queue             *FileProcessQueue // 文件处理队列
	stats             *ProcessingStats  // 处理统计
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
	s.cloudDriveService = GetCloudDriveService()

	// 初始化队列和统计信息
	s.queue = &FileProcessQueue{
		StrmFiles:     make([]FileEntry, 0),
		DownloadFiles: make([]FileEntry, 0),
	}
	s.stats = &ProcessingStats{}

	logger.Info("STRM 生成服务初始化完成")
}

// ProcessFileChangeEvent handles a single file change event from a webhook.
// It's designed to be called asynchronously for individual file creations.
func (s *StrmGeneratorService) ProcessFileChangeEvent(taskInfo *task.Task, event *webhook.FileChangeEvent) error {
	s.logger.Info("Processing file change event from webhook",
		zap.String("task", taskInfo.Name),
		zap.String("action", event.Action),
		zap.String("sourceFile", event.SourceFile),
	)

	// This function currently only handles file creation events.
	if event.IsDir {
		s.logger.Info("Event is for a directory, skipping.", zap.String("dir", event.SourceFile))
		// In the future, this could trigger a limited recursive scan.
		return nil
	}

	// 1. Load STRM config
	strmConfig, err := s.loadStrmConfig()
	if err != nil {
		return fmt.Errorf("failed to load strm config: %w", err)
	}

	// Normalize path separators to / for consistency
	sourceFileNormalized := strings.ReplaceAll(event.SourceFile, "\\", "/")
	taskSourcePathNormalized := strings.ReplaceAll(taskInfo.SourcePath, "\\", "/")

	// 2. Get the full file info (AListFile) for the new file.
	// We do this by listing the parent directory and finding the matching file.
	parentDir := filepath.Dir(sourceFileNormalized)
	files, err := s.listFiles(taskInfo, parentDir)
	if err != nil {
		return fmt.Errorf("failed to list files in parent directory '%s': %w", parentDir, err)
	}

	var aListFile *AListFile
	fileName := filepath.Base(sourceFileNormalized)
	for i := range files {
		if files[i].Name == fileName {
			aListFile = &files[i]
			break
		}
	}

	if aListFile == nil {
		return fmt.Errorf("could not find file info for '%s' in parent directory listing", sourceFileNormalized)
	}

	// 3. Determine the target path for the new file.
	// This is done by replacing the task's source path prefix with the task's target path.
	// Use the normalized paths for this operation.
	relativePath := strings.TrimPrefix(sourceFileNormalized, taskSourcePathNormalized)
	relativePath = strings.TrimPrefix(relativePath, "/") // Ensure no leading slash
	targetPath := filepath.Join(taskInfo.TargetPath, relativePath)

	// 4. Determine file type and process accordingly.
	fileType := s.determineFileType(aListFile, taskInfo, strmConfig)

	var success bool
	var errorMessage string

	s.logger.Info("Determined file type for webhook event",
		zap.String("file", aListFile.Name),
		zap.String("type", getFileTypeString(fileType)),
	)

	switch fileType {
	case FileTypeMedia:
		if !s.isMediaFileSizeValid(aListFile, strmConfig) {
			s.logger.Info("Skipping media file from webhook due to size constraints", zap.String("file", aListFile.Name))
			return nil
		}
		// Pass the original, non-normalized path to generateStrmFile, as it handles its own normalization.
		success, errorMessage, _ = s.generateStrmFile(aListFile, strmConfig, taskInfo, event.SourceFile, targetPath)
	case FileTypeMetadata, FileTypeSubtitle:
		// Pass the original, non-normalized path to downloadFile.
		success, errorMessage = s.downloadFile(aListFile, event.SourceFile, targetPath, taskInfo)
	default:
		s.logger.Info("Skipping file with unhandled type from webhook", zap.String("file", aListFile.Name))
		return nil // Not an error, just skipping.
	}

	if !success {
		return fmt.Errorf("failed to process file from webhook '%s': %s", event.SourceFile, errorMessage)
	}

	s.logger.Info("Successfully processed file from webhook", zap.String("file", event.SourceFile))
	return nil
}

// ProcessFileDeleteEvent handles a file deletion event from a webhook.
func (s *StrmGeneratorService) ProcessFileDeleteEvent(taskInfo *task.Task, sourceFilePath string) error {
	s.logger.Info("Processing file delete event from webhook",
		zap.String("task", taskInfo.Name),
		zap.String("sourceFile", sourceFilePath),
	)

	// 1. Calculate the expected target path for the deleted file.
	sourceFileNormalized := strings.ReplaceAll(sourceFilePath, "\\", "/")
	taskSourcePathNormalized := strings.ReplaceAll(taskInfo.SourcePath, "\\", "/")

	// Ensure the file is actually within the task's path
	if !strings.HasPrefix(sourceFileNormalized, taskSourcePathNormalized) {
		s.logger.Debug("Skipping delete event because file is not within task source path",
			zap.String("sourceFile", sourceFilePath),
			zap.String("taskSourcePath", taskInfo.SourcePath))
		return nil
	}

	relativePath := strings.TrimPrefix(sourceFileNormalized, taskSourcePathNormalized)
	relativePath = strings.TrimPrefix(relativePath, "/")

	// This is the path where the media file would have been, which we use to find siblings.
	targetMediaFilePath := filepath.Join(taskInfo.TargetPath, relativePath)
	targetDir := filepath.Dir(targetMediaFilePath)

	// 2. Find all related files in the target directory to delete.
	// This includes the .strm file, .nfo, subtitles, etc.
	baseName := strings.TrimSuffix(filepath.Base(targetMediaFilePath), filepath.Ext(targetMediaFilePath))
	filesToDelete, err := filepath.Glob(filepath.Join(targetDir, baseName+".*"))
	if err != nil {
		s.logger.Error("Error finding related files to delete via glob", zap.Error(err), zap.String("pattern", filepath.Join(targetDir, baseName+".*")))
		// Don't return, try to construct strm path manually.
	}

	// 3. Manually construct the .strm file path to ensure it's on the list.
	strmConfig, err := s.loadStrmConfig()
	if err != nil {
		s.logger.Warn("Could not load strm config for delete event, .strm file might be missed if glob failed", zap.Error(err))
	} else {
		var strmFileName string
		originalFileName := filepath.Base(sourceFilePath)
		if strmConfig.ReplaceSuffix {
			nameWithoutExt := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
			strmFileName = nameWithoutExt + ".strm"
		} else {
			strmFileName = originalFileName + ".strm"
		}
		strmFilePath := filepath.Join(targetDir, strmFileName)

		// Add to list if not already present
		found := false
		for _, f := range filesToDelete {
			if f == strmFilePath {
				found = true
				break
			}
		}
		if !found {
			filesToDelete = append(filesToDelete, strmFilePath)
		}
	}

	if len(filesToDelete) == 0 {
		s.logger.Info("No corresponding target files found to delete.", zap.String("sourceFile", sourceFilePath))
		return nil
	}

	// 4. Delete the files.
	var lastErr error
	for _, fileToDel := range filesToDelete {
		if _, err := os.Stat(fileToDel); err == nil {
			if err := os.Remove(fileToDel); err != nil {
				s.logger.Error("Failed to delete target file",
					zap.String("file", fileToDel),
					zap.Error(err),
				)
				lastErr = err // Keep track of the last error
			} else {
				s.logger.Info("Successfully deleted target file", zap.String("file", fileToDel))
			}
		}
	}

	return lastErr
}

// ProcessFileRenameEvent handles a file rename/move event from a webhook.
func (s *StrmGeneratorService) ProcessFileRenameEvent(taskInfo *task.Task, event *webhook.FileChangeEvent) error {
	s.logger.Info("Processing file rename/move event from webhook",
		zap.String("task", taskInfo.Name),
		zap.String("from", event.SourceFile),
		zap.String("to", event.DestinationFile),
	)

	// 1. Delete the old file(s) based on the source path
	if err := s.ProcessFileDeleteEvent(taskInfo, event.SourceFile); err != nil {
		// Log error but continue, as the source might not have been a strm file.
		s.logger.Warn("Error during delete part of rename event", zap.Error(err), zap.String("source", event.SourceFile))
	}

	// 2. Create the new file(s) based on the destination path
	// We can treat the destination file as a new creation event.
	createEvent := &webhook.FileChangeEvent{
		Action:     "create",
		IsDir:      event.IsDir,
		SourceFile: event.DestinationFile,
	}
	if err := s.ProcessFileChangeEvent(taskInfo, createEvent); err != nil {
		return fmt.Errorf("error during create part of rename event for '%s': %w", event.DestinationFile, err)
	}

	return nil
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

	// 如果任务成功完成，则刷新 Emby 媒体库
	if status == tasklog.TaskLogStatusCompleted && (generatedFiles > 0 || metadataDownloaded > 0 || subtitleDownloaded > 0) {
		s.logger.Info("开始刷新 Emby 媒体库", zap.String("taskName", taskInfo.Name))
		if refreshErr := Emby.RefreshAllLibraries(); refreshErr != nil {
			s.logger.Error("刷新 Emby 媒体库失败", zap.Error(refreshErr))
			// 仅记录错误，不影响任务状态
		} else {
			s.logger.Info("刷新 Emby 媒体库请求成功", zap.String("taskName", taskInfo.Name))
		}
	}

	return err
}

// loadStrmConfig 加载 STRM 配置
func (s *StrmGeneratorService) loadStrmConfig() (*StrmConfig, error) {
	config, err := repository.Config.GetByCode("STRM")
	if err != nil || config == nil {
		var errorMessage string
		if err != nil {
			s.logger.Error("获取 STRM 配置失败", zap.Error(err))
			errorMessage = fmt.Sprintf("获取 STRM 配置失败: %v", err)
		} else {
			s.logger.Error("STRM 配置未找到")
			errorMessage = "STRM 配置未找到"
		}
		return nil, fmt.Errorf("获取 STRM 配置失败: %s", errorMessage)
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

// listFiles 根据任务配置类型获取文件列表
func (s *StrmGeneratorService) listFiles(taskInfo *task.Task, path string) ([]AListFile, error) {
	switch taskInfo.ConfigType {
	case "alist":
		if s.alistService == nil {
			return nil, fmt.Errorf("AList service is not initialized")
		}
		return s.alistService.ListFiles(path)
	case "local":
		return s.listLocalFiles(path)
	case "clouddrive":
		if s.cloudDriveService == nil {
			return nil, fmt.Errorf("CloudDrive service is not initialized")
		}
		return s.cloudDriveService.ListFiles(path)
	default:
		return nil, fmt.Errorf("unsupported ConfigType: %s", taskInfo.ConfigType)
	}
}

// listLocalFiles 从本地文件系统读取文件列表
func (s *StrmGeneratorService) listLocalFiles(path string) ([]AListFile, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read local directory %s: %w", path, err)
	}

	var files []AListFile
	for _, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			// 记录错误但继续处理其他文件
			s.logger.Warn("Failed to get file info for entry, skipping", zap.String("entry", entry.Name()), zap.Error(err))
			continue
		}

		// 将 os.FileInfo 转换为 AListFile 结构
		file := AListFile{
			Name:     info.Name(),
			Size:     info.Size(),
			IsDir:    info.IsDir(),
			Modified: info.ModTime(),
			// 以下字段对于本地文件可能不适用，使用零值
			Sign:  "",
			Thumb: "",
			Type:  0,
			HashInfo: struct {
				Sha1 string `json:"sha1"`
			}{},
		}
		files = append(files, file)
	}
	return files, nil
}

// scanDirectoryRecursive 递归扫描目录，只收集文件信息，不进行处理
func (s *StrmGeneratorService) scanDirectoryRecursive(taskInfo *task.Task, strmConfig *StrmConfig,
	taskLogID uint, sourcePath, targetPath string) error {

	// 根据任务类型获取当前目录的文件列表
	files, err := s.listFiles(taskInfo, sourcePath)
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

	// 先对文件进行分类
	var currentDirectoryFileCount int // 当前目录的文件数量（不包含文件夹）

	for _, file := range files {
		if file.IsDir {
			directoryFiles = append(directoryFiles, &file)
			continue
		}

		// 只有在这里才增加文件计数（跳过了文件夹）
		currentDirectoryFileCount++

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
			// 对于媒体文件，需要检查文件大小是否满足最小要求
			if s.isMediaFileSizeValid(&file, strmConfig) {
				mediaFileEntries = append(mediaFileEntries, entry)
			} else {
				// 媒体文件大小不满足要求，计入跳过文件
				s.stats.Mutex.Lock()
				s.stats.SkipFile++
				s.stats.Mutex.Unlock()
				s.logger.Debug("跳过不满足大小要求的媒体文件",
					zap.String("fileName", file.Name),
					zap.String("path", sourcePath),
					zap.Int64("fileSize", file.Size),
					zap.Int64("minSize", strmConfig.MinFileSize*1024*1024))
			}
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

	// 文件分类完成后，增加总文件计数（只统计文件，不包含文件夹）
	s.stats.Mutex.Lock()
	s.stats.TotalFiles += currentDirectoryFileCount
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

	s.logger.Debug("当前目录文件统计",
		zap.String("sourcePath", sourcePath),
		zap.Int("当前目录文件数", currentDirectoryFileCount),
		zap.Int("累计总文件数", totalFiles),
		zap.Int("文件夹数", len(directoryFiles)))

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

// isMediaFileSizeValid 检查媒体文件大小是否满足最小大小要求
func (s *StrmGeneratorService) isMediaFileSizeValid(file *AListFile, strmConfig *StrmConfig) bool {
	// 如果 minFileSize 为 0，表示不过滤
	if strmConfig.MinFileSize <= 0 {
		return true
	}

	// 将 MB 转换为字节进行比较
	minFileSizeBytes := strmConfig.MinFileSize * 1024 * 1024

	if file.Size < minFileSizeBytes {
		s.logger.Debug("媒体文件大小不满足最小要求，跳过处理",
			zap.String("文件名", file.Name),
			zap.Int64("文件大小(字节)", file.Size),
			zap.Int64("最小大小要求(字节)", minFileSizeBytes),
			zap.String("文件大小(友好显示)", humanizeSize(file.Size)),
			zap.Int64("最小大小要求(MB)", strmConfig.MinFileSize))
		return false
	}

	return true
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
	var fileURL string

	// 根据任务类型构建不同的URL
	switch taskConfig.ConfigType {
	case "alist":
		// 处理路径和文件名的 URL 编码
		dirPath := filepath.Dir(sourcePath)
		fileName := file.Name

		// 根据 URLEncode 配置决定是否需要对路径进行编码
		if strmConfig.URLEncode {
			dirPath = strings.ReplaceAll(dirPath, "\\", "/")
			pathParts := strings.Split(dirPath, "/")
			for i, part := range pathParts {
				pathParts[i] = url.PathEscape(part)
			}
			dirPath = strings.Join(pathParts, "/")
			fileName = url.PathEscape(fileName)
			s.logger.Debug("进行了URL编码",
				zap.String("原路径", filepath.Dir(sourcePath)),
				zap.String("编码后路径", dirPath),
				zap.String("原文件名", file.Name),
				zap.String("编码后文件名", fileName))
		}

		if taskConfig.ConfigType == "alist" {
			fileURL = s.alistService.GetFileURL(dirPath, fileName, file.Sign)
		}
	case "clouddrive":
		// 处理路径和文件名的 URL 编码
		dirPath := filepath.Dir(sourcePath)
		fileName := file.Name

		// 根据 URLEncode 配置决定是否需要对路径进行编码
		if strmConfig.URLEncode {
			dirPath = strings.ReplaceAll(dirPath, "\\", "/")
			pathParts := strings.Split(dirPath, "/")
			for i, part := range pathParts {
				pathParts[i] = url.PathEscape(part)
			}
			dirPath = strings.Join(pathParts, "/")
			fileName = url.PathEscape(fileName)
			s.logger.Debug("进行了URL编码",
				zap.String("原路径", filepath.Dir(sourcePath)),
				zap.String("编码后路径", dirPath),
				zap.String("原文件名", file.Name),
				zap.String("编码后文件名", fileName))
		}

		fileURL = s.cloudDriveService.GetFileURL(dirPath, fileName, file.Sign)
	case "local":
		// 本地文件直接使用源路径
		fileURL = sourcePath
	default:
		return false, fmt.Sprintf("不支持的 ConfigType 用于生成 STRM URL: %s", taskConfig.ConfigType), ""
	}

	if fileURL == "" {
		return false, fmt.Sprintf("无法为类型 %s 生成文件URL，请检查相关配置是否完整", taskConfig.ConfigType), ""
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

	// 对于本地文件，执行文件复制
	if taskConfig.ConfigType == "local" {
		in, err := os.Open(sourcePath)
		if err != nil {
			return false, fmt.Sprintf("打开源文件失败: %v", err)
		}
		defer in.Close()

		out, err := os.Create(targetPath)
		if err != nil {
			return false, fmt.Sprintf("创建目标文件失败: %v", err)
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return false, fmt.Sprintf("复制文件失败: %v", err)
		}
		s.logger.Info("复制本地文件成功",
			zap.String("sourceFile", sourcePath),
			zap.String("targetPath", targetPath))
		return true, ""
	}

	// --- 对于远程文件 (alist, clouddrive)，执行下载 ---

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
		dirPath = strings.ReplaceAll(dirPath, "\\", "/")
		pathParts := strings.Split(dirPath, "/")
		for i, part := range pathParts {
			pathParts[i] = url.PathEscape(part)
		}
		dirPath = strings.Join(pathParts, "/")

		// 对文件名编码
		fileName = url.PathEscape(fileName)
	}

	var fileURL string
	switch taskConfig.ConfigType {
	case "alist":
		fileURL = s.alistService.GetFileURL(dirPath, fileName, file.Sign)
	case "clouddrive":
		fileURL = s.cloudDriveService.GetFileURL(dirPath, fileName, file.Sign)
	default:
		return false, fmt.Sprintf("不支持的 ConfigType 用于下载: %s", taskConfig.ConfigType)
	}

	if fileURL == "" {
		return false, fmt.Sprintf("无法为类型 %s 生成文件下载URL，请检查相关配置", taskConfig.ConfigType)
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
		// hash = file.HashInfo.Sha1
		// 测试为 null 的结果
		hash = ""
	}

	// 记录文件类型和是否有Hash值
	s.logger.Debug("处理文件历史记录",
		zap.String("fileName", file.Name),
		zap.String("fileType", fileTypeStr),
		zap.Bool("hasHash", hash != ""),
		zap.String("hash", hash))

	// 查找现有记录 - 支持两种查找方式
	var existingRecord *filehistory.FileHistory
	var findErr error

	// 1. 优先通过 Hash 查找（如果有 hash）
	if hash != "" {
		existingRecord, findErr = repository.FileHistory.GetByHash(hash)
		if findErr != nil {
			s.logger.Debug("通过Hash查找文件历史记录失败",
				zap.String("hash", hash),
				zap.Error(findErr))
		}
	}

	// 2. 如果通过 Hash 没找到记录或者没有 hash，则通过文件路径和属性查找
	if existingRecord == nil {
		existingRecord, findErr = repository.FileHistory.GetByFileAttributes(sourcePath, file.Name, file.Size, fileTypeStr)
		if findErr != nil {
			s.logger.Debug("通过文件属性查找文件历史记录失败",
				zap.String("sourcePath", sourcePath),
				zap.String("fileName", file.Name),
				zap.Int64("fileSize", file.Size),
				zap.String("fileType", fileTypeStr),
				zap.Error(findErr))
		} else if existingRecord != nil {
			existingHashStr := ""
			if existingRecord.Hash != nil {
				existingHashStr = *existingRecord.Hash
			}
			s.logger.Debug("通过文件属性找到现有记录",
				zap.String("fileName", file.Name),
				zap.String("existingHash", existingHashStr),
				zap.String("newHash", hash))
		}
	}

	// 如果找到现有记录，更新它
	if existingRecord != nil {
		now := time.Now()
		updateData := map[string]interface{}{
			"task_id":     taskID,
			"task_log_id": taskLogID,
			"updated_at":  now,
			"file_size":   file.Size,
			"modified_at": &file.Modified,
		}

		// 处理 hash 字段更新
		if hash != "" {
			// 有新的 hash 值，更新它
			updateData["hash"] = &hash
		} else if existingRecord.Hash != nil {
			// 没有新 hash，但保持现有 hash
			updateData["hash"] = existingRecord.Hash
		}
		// 如果 hash 为空且现有记录也没有 hash，则不更新 hash 字段

		if err := repository.FileHistory.UpdateByID(existingRecord.ID, updateData); err != nil {
			s.logger.Error("更新文件历史记录失败",
				zap.String("fileName", file.Name),
				zap.String("hash", hash),
				zap.Error(err))
		} else {
			existingHashStr := ""
			if existingRecord.Hash != nil {
				existingHashStr = *existingRecord.Hash
			}
			s.logger.Info("更新现有文件历史记录",
				zap.String("fileName", file.Name),
				zap.String("newHash", hash),
				zap.String("existingHash", existingHashStr),
				zap.String("fileType", fileTypeStr),
				zap.Uint("oldTaskID", existingRecord.TaskID),
				zap.Uint("newTaskID", taskID))
		}
		return
	}

	// 没有找到现有记录，创建新记录
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
	}

	// 只有当 hash 不为空时才设置 Hash 字段，否则保持为 nil (数据库中的 NULL)
	if hash != "" {
		fileHistory.Hash = &hash
	}

	if err := repository.FileHistory.Create(fileHistory); err != nil {
		// 检查是否是唯一约束错误（文件已存在）
		if strings.Contains(strings.ToLower(err.Error()), "unique") ||
			strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "constraint") {
			s.logger.Info("文件历史记录已存在（数据库约束），跳过创建",
				zap.String("fileName", file.Name),
				zap.String("sourcePath", sourcePath),
				zap.String("hash", hash),
				zap.String("errorDetails", err.Error()))
		} else {
			s.logger.Error("记录文件历史失败",
				zap.String("fileName", file.Name),
				zap.Error(err))
		}
	} else {
		s.logger.Debug("成功创建新文件历史记录",
			zap.String("fileName", file.Name),
			zap.String("fileType", fileTypeStr),
			zap.String("hash", hash))
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
