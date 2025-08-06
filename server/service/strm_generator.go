package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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
	urlEncodeCache    *URLEncodeCache   // URL编码缓存
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
			stats:          &ProcessingStats{},
			urlEncodeCache: NewURLEncodeCache(),
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
	s.urlEncodeCache = NewURLEncodeCache()

	logger.Info("STRM 生成服务初始化完成")
}

// ProcessFileChangeEvent 处理来自 webhook 的单个文件变更事件。
// 该方法设计为可异步调用，用于处理单个文件的创建事件。
func (s *StrmGeneratorService) ProcessFileChangeEvent(taskInfo *task.Task, event *webhook.FileChangeEvent) error {
	s.logger.Info("Processing file change event from webhook",
		zap.String("task", taskInfo.Name),
		zap.String("action", event.Action),
		zap.String("sourceFile", event.SourceFile),
	)

	// 处理目录事件 - 如果是新目录，则扫描其中的文件
	if event.IsDir {
		s.logger.Info("Event is for a directory, checking for files inside.", zap.String("dir", event.SourceFile))

		// 对于目录创建事件，我们应该扫描目录中的文件
		if event.Action == "create" || event.Action == "mkdir" {
			return s.processDirectoryEvent(taskInfo, event.SourceFile)
		}

		// 对于其他目录事件（删除、重命名），暂时跳过
		s.logger.Info("Directory event is not creation, skipping.",
			zap.String("dir", event.SourceFile),
			zap.String("action", event.Action))
		return nil
	}

	// 1. 加载 STRM 配置
	strmConfig, err := s.loadStrmConfig()
	if err != nil {
		return fmt.Errorf("failed to load strm config: %w", err)
	}

	// 标准化路径分隔符为 / 以保持一致性
	sourceFileNormalized := strings.ReplaceAll(event.SourceFile, "\\", "/")
	taskSourcePathNormalized := strings.ReplaceAll(taskInfo.SourcePath, "\\", "/")

	// 2. 获取新文件的完整文件信息（AListFile）。
	// 我们通过列出父目录并找到匹配的文件来实现这一点。
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

	// 3. 确定新文件的目标路径。
	// 这是通过用任务的目标路径替换任务的源路径前缀来完成的。
	// 使用标准化路径进行此操作。
	relativePath := strings.TrimPrefix(sourceFileNormalized, taskSourcePathNormalized)
	relativePath = strings.TrimPrefix(relativePath, "/") // 确保没有前导斜杠
	targetPath := filepath.Join(taskInfo.TargetPath, relativePath)

	// 4. 确定文件类型并进行相应处理。
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
			s.logger.Debug("Skipping media file from webhook due to size constraints", zap.String("file", aListFile.Name))
			return nil
		}
		// 将原始的、非标准化的路径传递给 generateStrmFile，因为它会处理自己的标准化。
		success, errorMessage, _ = s.generateStrmFile(aListFile, strmConfig, taskInfo, event.SourceFile, targetPath)
	case FileTypeMetadata, FileTypeSubtitle:
		// 将原始的、非标准化的路径传递给 downloadFile。
		success, errorMessage = s.downloadFile(aListFile, event.SourceFile, targetPath, taskInfo)
	default:
		s.logger.Info("Skipping file with unhandled type from webhook", zap.String("file", aListFile.Name))
		return nil // 不是错误，只是跳过。
	}

	if !success {
		return fmt.Errorf("failed to process file from webhook '%s': %s", event.SourceFile, errorMessage)
	}

	// 记录成功处理文件的文件历史
	// 使用 0 作为 taskLogID，因为这是 webhook 事件，不是批处理任务
	s.recordFileHistory(taskInfo.ID, 0, aListFile, event.SourceFile, targetPath, fileType, true)

	s.logger.Info("Successfully processed file from webhook", zap.String("file", event.SourceFile))
	return nil
}

// processDirectoryEvent 通过扫描目录中的文件来处理目录创建事件
func (s *StrmGeneratorService) processDirectoryEvent(taskInfo *task.Task, dirPath string) error {
	s.logger.Info("Processing directory creation event",
		zap.String("task", taskInfo.Name),
		zap.String("dirPath", dirPath))

	// 加载 STRM 配置
	strmConfig, err := s.loadStrmConfig()
	if err != nil {
		return fmt.Errorf("failed to load strm config: %w", err)
	}

	// 列出新目录中的文件
	files, err := s.listFiles(taskInfo, dirPath)
	if err != nil {
		return fmt.Errorf("failed to list files in directory '%s': %w", dirPath, err)
	}

	if len(files) == 0 {
		s.logger.Info("Directory is empty, nothing to process", zap.String("dirPath", dirPath))
		return nil
	}

	s.logger.Info("Found files in new directory",
		zap.String("dirPath", dirPath),
		zap.Int("fileCount", len(files)))

	// 标准化路径分隔符以保持一致性
	dirPathNormalized := strings.ReplaceAll(dirPath, "\\", "/")
	taskSourcePathNormalized := strings.ReplaceAll(taskInfo.SourcePath, "\\", "/")

	// 计算目标目录路径
	relativePath := strings.TrimPrefix(dirPathNormalized, taskSourcePathNormalized)
	relativePath = strings.TrimPrefix(relativePath, "/")
	targetDirPath := filepath.Join(taskInfo.TargetPath, relativePath)

	// 如果目标目录不存在则创建
	if err := s.safeMkdirAll(targetDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory '%s': %w", targetDirPath, err)
	}

	// 处理目录中的每个文件
	var processedCount int
	var errorCount int

	for _, file := range files {
		if file.IsDir {
			// 递归处理子目录
			subDirPath := filepath.Join(dirPath, file.Name)

			// 检查子目录路径长度是否合理
			// 如果路径过长，我们可能需要以不同方式处理
			if len(subDirPath) > 4000 { // 防止问题的保守限制
				s.logger.Warn("Subdirectory path is very long, skipping to prevent issues",
					zap.String("subDir", subDirPath),
					zap.Int("pathLength", len(subDirPath)))
				errorCount++
				continue
			}

			if err := s.processDirectoryEvent(taskInfo, subDirPath); err != nil {
				s.logger.Error("Failed to process subdirectory",
					zap.String("subDir", subDirPath),
					zap.Error(err))
				errorCount++
			}
			continue
		}

		// 处理单个文件
		sourceFilePath := filepath.Join(dirPath, file.Name)
		targetFilePath := filepath.Join(targetDirPath, file.Name)

		// 确定文件类型
		fileType := s.determineFileType(&file, taskInfo, strmConfig)

		var success bool
		var errorMessage string

		switch fileType {
		case FileTypeMedia:
			if !s.isMediaFileSizeValid(&file, strmConfig) {
				s.logger.Debug("Skipping media file due to size constraints",
					zap.String("file", file.Name))
				continue
			}
			success, errorMessage, _ = s.generateStrmFile(&file, strmConfig, taskInfo, sourceFilePath, targetFilePath)
		case FileTypeMetadata, FileTypeSubtitle:
			success, errorMessage = s.downloadFile(&file, sourceFilePath, targetFilePath, taskInfo)
		default:
			s.logger.Debug("Skipping file with unhandled type",
				zap.String("file", file.Name),
				zap.String("type", getFileTypeString(fileType)))
			continue
		}

		if success {
			processedCount++
			s.logger.Debug("Successfully processed file from directory",
				zap.String("file", sourceFilePath))

			// 记录成功处理文件的文件历史
			// 使用 0 作为 taskLogID，因为这是 webhook 事件，不是批处理任务
			s.recordFileHistory(taskInfo.ID, 0, &file, sourceFilePath, targetFilePath, fileType, true)
		} else {
			errorCount++
			s.logger.Error("Failed to process file from directory",
				zap.String("file", sourceFilePath),
				zap.String("error", errorMessage))
		}
	}

	s.logger.Info("Completed processing directory",
		zap.String("dirPath", dirPath),
		zap.Int("processedFiles", processedCount),
		zap.Int("errorCount", errorCount))

	if errorCount > 0 {
		return fmt.Errorf("processed directory with %d errors out of %d files", errorCount, len(files))
	}

	return nil
}

// ProcessFileDeleteEvent 处理来自 webhook 的文件删除事件。
func (s *StrmGeneratorService) ProcessFileDeleteEvent(taskInfo *task.Task, sourceFilePath string) error {
	s.logger.Info("Processing file delete event from webhook",
		zap.String("task", taskInfo.Name),
		zap.String("sourceFile", sourceFilePath),
	)

	// 1. 计算已删除文件的预期目标路径。
	sourceFileNormalized := strings.ReplaceAll(sourceFilePath, "\\", "/")
	taskSourcePathNormalized := strings.ReplaceAll(taskInfo.SourcePath, "\\", "/")

	// 确保文件实际上在任务的路径内
	if !strings.HasPrefix(sourceFileNormalized, taskSourcePathNormalized) {
		s.logger.Debug("Skipping delete event because file is not within task source path",
			zap.String("sourceFile", sourceFilePath),
			zap.String("taskSourcePath", taskInfo.SourcePath))
		return nil
	}

	relativePath := strings.TrimPrefix(sourceFileNormalized, taskSourcePathNormalized)
	relativePath = strings.TrimPrefix(relativePath, "/")

	// 这是媒体文件本应在的路径，我们用它来查找相关文件。
	targetMediaFilePath := filepath.Join(taskInfo.TargetPath, relativePath)
	targetDir := filepath.Dir(targetMediaFilePath)

	// 2. 查找目标目录中所有需要删除的相关文件。
	// 这包括 .strm 文件、.nfo 文件、字幕文件等。
	baseName := strings.TrimSuffix(filepath.Base(targetMediaFilePath), filepath.Ext(targetMediaFilePath))
	filesToDelete, err := filepath.Glob(filepath.Join(targetDir, baseName+".*"))
	if err != nil {
		s.logger.Error("Error finding related files to delete via glob", zap.Error(err), zap.String("pattern", filepath.Join(targetDir, baseName+".*")))
		// 不要返回，尝试手动构建 strm 路径。
	}

	// 3. 手动构建 .strm 文件路径以确保它在列表中。
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

		// 如果尚未存在，则添加到列表中
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

	// 4. 删除文件。
	var lastErr error
	for _, fileToDel := range filesToDelete {
		if _, err := os.Stat(fileToDel); err == nil {
			if err := os.Remove(fileToDel); err != nil {
				s.logger.Error("Failed to delete target file",
					zap.String("file", fileToDel),
					zap.Error(err),
				)
				lastErr = err // 记录最后一个错误
			} else {
				s.logger.Info("Successfully deleted target file", zap.String("file", fileToDel))
			}
		}
	}

	return lastErr
}

// ProcessFileRenameEvent 处理来自 webhook 的文件重命名/移动事件。
func (s *StrmGeneratorService) ProcessFileRenameEvent(taskInfo *task.Task, event *webhook.FileChangeEvent) error {
	s.logger.Info("Processing file rename/move event from webhook",
		zap.String("task", taskInfo.Name),
		zap.String("from", event.SourceFile),
		zap.String("to", event.DestinationFile),
	)

	// 1. 基于源路径删除旧文件
	if err := s.ProcessFileDeleteEvent(taskInfo, event.SourceFile); err != nil {
		// 记录错误但继续，因为源文件可能不是 strm 文件。
		s.logger.Warn("Error during delete part of rename event", zap.Error(err), zap.String("source", event.SourceFile))
	}

	// 2. 基于目标路径创建新文件
	// 我们可以将目标文件视为新的创建事件。
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
	// 跳过的文件总和：STRM文件跳过 + 其他文件跳过
	// 注意：元数据和字幕的跳过不计入总跳过数，因为它们有单独的统计字段
	// 这样确保：总文件数 = 生成文件数 + 跳过文件数 + 元数据文件数 + 字幕文件数
	skippedFiles := s.stats.SkipFile + s.stats.OtherSkipped
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

// StreamingScanner 流式目录扫描器，优化大目录的内存使用
type StreamingScanner struct {
	service       *StrmGeneratorService
	taskInfo      *task.Task
	strmConfig    *StrmConfig
	taskLogID     uint
	maxQueueSize  int             // 最大队列大小，控制内存使用
	processedDirs map[string]bool // 已处理目录缓存，避免重复扫描
	mutex         sync.RWMutex
}

// NewStreamingScanner 创建流式扫描器
func NewStreamingScanner(service *StrmGeneratorService, taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint) *StreamingScanner {
	return &StreamingScanner{
		service:       service,
		taskInfo:      taskInfo,
		strmConfig:    strmConfig,
		taskLogID:     taskLogID,
		maxQueueSize:  1000, // 限制队列大小，避免内存过度使用
		processedDirs: make(map[string]bool),
	}
}

// scanDirectoryRecursive 递归扫描目录，只收集文件信息，不进行处理
func (s *StrmGeneratorService) scanDirectoryRecursive(taskInfo *task.Task, strmConfig *StrmConfig,
	taskLogID uint, sourcePath, targetPath string) error {

	// 使用流式扫描器优化大目录处理
	scanner := NewStreamingScanner(s, taskInfo, strmConfig, taskLogID)
	return scanner.scanWithMemoryControl(sourcePath, targetPath)
}

// scanWithMemoryControl 带内存控制的扫描方法
func (scanner *StreamingScanner) scanWithMemoryControl(sourcePath, targetPath string) error {
	return scanner.scanDirectoryRecursiveInternal(sourcePath, targetPath)
}

// scanDirectoryRecursiveInternal 内部递归扫描实现
func (scanner *StreamingScanner) scanDirectoryRecursiveInternal(sourcePath, targetPath string) error {
	// 检查是否已处理过此目录
	scanner.mutex.RLock()
	if scanner.processedDirs[sourcePath] {
		scanner.mutex.RUnlock()
		return nil
	}
	scanner.mutex.RUnlock()

	// 标记目录为已处理
	scanner.mutex.Lock()
	scanner.processedDirs[sourcePath] = true
	scanner.mutex.Unlock()

	// 根据任务类型获取当前目录的文件列表
	files, err := scanner.service.listFiles(scanner.taskInfo, sourcePath)
	if err != nil {
		return fmt.Errorf("获取目录文件列表失败 [%s]: %w", sourcePath, err)
	}

	scanner.service.logger.Info("扫描目录",
		zap.String("sourcePath", sourcePath),
		zap.String("targetPath", targetPath),
		zap.Int("fileCount", len(files)))

	// 创建目标目录
	if err := scanner.service.safeMkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败 [%s]: %w", targetPath, err)
	}

	// 使用流式处理，避免大目录时内存占用过多
	return scanner.processFilesStreaming(files, sourcePath, targetPath)
}

// processFilesStreaming 流式处理文件，优化内存使用
func (scanner *StreamingScanner) processFilesStreaming(files []AListFile, sourcePath, targetPath string) error {
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
		fileType := scanner.service.determineFileType(&file, scanner.taskInfo, scanner.strmConfig)

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
			if scanner.service.isMediaFileSizeValid(&file, scanner.strmConfig) {
				mediaFileEntries = append(mediaFileEntries, entry)
			} else {
				// 媒体文件大小不满足要求，计入跳过文件
				scanner.service.stats.Mutex.Lock()
				scanner.service.stats.SkipFile++
				scanner.service.stats.Mutex.Unlock()
				scanner.service.logger.Debug("跳过不满足大小要求的媒体文件",
					zap.String("fileName", file.Name),
					zap.String("path", sourcePath),
					zap.Int64("fileSize", file.Size),
					zap.Int64("minSize", scanner.strmConfig.MinFileSize*1024*1024))
			}
		case FileTypeSubtitle:
			subtitleFileEntries = append(subtitleFileEntries, entry)
		case FileTypeMetadata:
			metadataFileEntries = append(metadataFileEntries, entry)
		default:
			// 其他文件类型不处理，但计入跳过文件
			scanner.service.stats.Mutex.Lock()
			scanner.service.stats.OtherSkipped++
			scanner.service.stats.Mutex.Unlock()
		}

		// 内存控制：当媒体文件队列过大时，先处理一批
		if len(mediaFileEntries) >= scanner.maxQueueSize {
			scanner.flushMediaFiles(mediaFileEntries)
			mediaFileEntries = mediaFileEntries[:0] // 清空切片但保留容量
		}
	}

	// 处理剩余的媒体文件
	if len(mediaFileEntries) > 0 {
		scanner.flushMediaFiles(mediaFileEntries)
	}

	// 文件分类完成后，增加总文件计数（只统计文件，不包含文件夹）
	scanner.service.stats.Mutex.Lock()
	scanner.service.stats.TotalFiles += currentDirectoryFileCount
	totalFiles := scanner.service.stats.TotalFiles
	scanner.service.stats.Mutex.Unlock()

	// 减少数据库更新频率，提高性能
	if totalFiles%500 == 0 {
		updateData := map[string]interface{}{
			"total_file": totalFiles,
		}
		if updateErr := repository.TaskLog.UpdatePartial(scanner.taskLogID, updateData); updateErr != nil {
			scanner.service.logger.Error("更新任务日志总文件数失败", zap.Error(updateErr))
		}
	}

	scanner.service.logger.Debug("当前目录文件统计",
		zap.String("sourcePath", sourcePath),
		zap.Int("当前目录文件数", currentDirectoryFileCount),
		zap.Int("累计总文件数", totalFiles),
		zap.Int("文件夹数", len(directoryFiles)))

	// 处理字幕和元数据文件
	return scanner.processSubtitleAndMetadata(subtitleFileEntries, metadataFileEntries, directoryFiles, sourcePath, targetPath)
}

// flushMediaFiles 批量处理媒体文件，避免队列过大
func (scanner *StreamingScanner) flushMediaFiles(mediaFileEntries []FileEntry) {
	if len(mediaFileEntries) == 0 {
		return
	}

	scanner.service.queue.FilesMutex.Lock()
	scanner.service.queue.StrmFiles = append(scanner.service.queue.StrmFiles, mediaFileEntries...)
	scanner.service.queue.FilesMutex.Unlock()

	scanner.service.logger.Debug("批量添加媒体文件到队列",
		zap.Int("文件数", len(mediaFileEntries)))
}

// processSubtitleAndMetadata 处理字幕和元数据文件
func (scanner *StreamingScanner) processSubtitleAndMetadata(subtitleFileEntries, metadataFileEntries []FileEntry, directoryFiles []*AListFile, sourcePath, targetPath string) error {
	// 筛选需要处理的字幕文件（需要与媒体文件匹配）
	// 注意：由于我们采用流式处理，这里需要检查全局媒体文件，而不仅仅是当前批次
	var matchedSubtitleEntries []FileEntry
	for _, subEntry := range subtitleFileEntries {
		// 简化匹配逻辑，假设同目录下的字幕文件都需要处理
		// 实际匹配可以在下载时进行更精确的检查
		matchedSubtitleEntries = append(matchedSubtitleEntries, subEntry)
	}

	// 下载文件先收集，等扫描结束后再处理
	// 先筛选出需要下载的文件（本地不存在的文件）
	var needDownloadEntries []FileEntry

	// 批量检查文件是否存在，减少系统调用
	allEntriesToCheck := append(matchedSubtitleEntries, metadataFileEntries...)
	existenceMap := scanner.service.batchCheckFileExistence(allEntriesToCheck)

	// 检查匹配的字幕文件是否已存在于本地
	for _, entry := range matchedSubtitleEntries {
		if !existenceMap[entry.TargetPath] {
			needDownloadEntries = append(needDownloadEntries, entry)
		} else {
			scanner.service.logger.Debug("字幕文件已存在，跳过下载",
				zap.String("fileName", entry.File.Name),
				zap.String("targetPath", entry.TargetPath))

			// 记录文件历史（已存在的文件）
			scanner.service.recordFileHistory(scanner.taskInfo.ID, scanner.taskLogID, entry.File, entry.SourcePath, entry.TargetPath, entry.FileType, true)

			// 更新统计信息
			scanner.service.stats.Mutex.Lock()
			scanner.service.stats.SubtitleSkipped++ // 计入已跳过的字幕文件
			scanner.service.stats.Mutex.Unlock()
		}
	}

	// 检查元数据文件是否已存在于本地
	for _, entry := range metadataFileEntries {
		if !existenceMap[entry.TargetPath] {
			needDownloadEntries = append(needDownloadEntries, entry)
		} else {
			scanner.service.logger.Debug("元数据文件已存在，跳过下载",
				zap.String("fileName", entry.File.Name),
				zap.String("targetPath", entry.TargetPath))

			// 记录文件历史（已存在的文件）
			scanner.service.recordFileHistory(scanner.taskInfo.ID, scanner.taskLogID, entry.File, entry.SourcePath, entry.TargetPath, entry.FileType, true)

			// 更新统计信息
			scanner.service.stats.Mutex.Lock()
			scanner.service.stats.MetadataSkipped++ // 计入已跳过的元数据文件
			scanner.service.stats.Mutex.Unlock()
		}
	}

	// 只将需要下载的文件添加到下载队列
	if len(needDownloadEntries) > 0 {
		scanner.service.queue.FilesMutex.Lock()
		scanner.service.queue.DownloadFiles = append(scanner.service.queue.DownloadFiles, needDownloadEntries...)
		scanner.service.queue.FilesMutex.Unlock()

		// 获取已跳过的文件数（已包含在总跳过文件数中）
		skippedExistingFiles := len(matchedSubtitleEntries) + len(metadataFileEntries) - len(needDownloadEntries)

		scanner.service.stats.Mutex.RLock()
		scanner.service.logger.Info("添加文件到下载队列",
			zap.Int("需下载文件数", len(needDownloadEntries)),
			zap.Int("跳过已存在文件数", skippedExistingFiles),
			zap.Int("已跳过字幕文件", scanner.service.stats.SubtitleSkipped),
			zap.Int("已跳过元数据文件", scanner.service.stats.MetadataSkipped),
			zap.Int("跳过其他文件数", scanner.service.stats.OtherSkipped))
		scanner.service.stats.Mutex.RUnlock()
	}

	// 递归处理子目录
	for _, dirFile := range directoryFiles {
		currentSourcePath := filepath.Join(sourcePath, dirFile.Name)
		currentTargetPath := filepath.Join(targetPath, dirFile.Name)

		// 递归处理子目录
		if err := scanner.scanDirectoryRecursiveInternal(currentSourcePath, currentTargetPath); err != nil {
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

// truncateFileName 截断文件名以避免在Linux系统中因文件名过长导致创建失败
// Linux文件系统通常限制文件名最大长度为255字节
func (s *StrmGeneratorService) truncateFileName(fileName string) string {
	const maxFileNameLength = 200

	// 如果文件名长度在限制内，直接返回
	if len(fileName) <= maxFileNameLength {
		return fileName
	}

	// 获取文件扩展名
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	// 为了避免截断后的文件名冲突，我们添加一个短哈希后缀
	// 计算原文件名的哈希值，取前8位作为后缀
	hash := fmt.Sprintf("%08x", hashString(fileName))
	hashSuffix := "_" + hash[:8]

	// 计算可用于文件名主体的最大长度（需要为哈希后缀和扩展名预留空间）
	maxNameLength := maxFileNameLength - len(ext) - len(hashSuffix)

	// 如果扩展名和哈希后缀本身就超过了限制，这种情况很少见，但需要处理
	if maxNameLength <= 0 {
		// 如果扩展名过长，截断整个文件名，不添加哈希后缀
		s.logger.Warn("文件扩展名过长，截断整个文件名",
			zap.String("原文件名", fileName),
			zap.String("扩展名", ext),
			zap.Int("扩展名长度", len(ext)))
		return fileName[:maxFileNameLength]
	}

	// 截断文件名主体，添加哈希后缀，保留扩展名
	truncatedName := nameWithoutExt[:maxNameLength] + hashSuffix + ext

	s.logger.Debug("文件名过长，已截断并添加哈希后缀",
		zap.String("原文件名", fileName),
		zap.String("截断后文件名", truncatedName),
		zap.Int("原长度", len(fileName)),
		zap.Int("截断后长度", len(truncatedName)),
		zap.Int("最大允许长度", maxFileNameLength),
		zap.String("哈希后缀", hashSuffix))

	return truncatedName
}

// truncatePathLength 处理整个路径长度限制，包括目录名和文件名
// 主要处理每个目录名和文件名的长度限制（255字节）
func (s *StrmGeneratorService) truncatePathLength(fullPath string) string {
	// 首先处理路径中每个组件的长度限制
	pathComponents := strings.Split(fullPath, string(filepath.Separator))
	var truncatedComponents []string
	var pathChanged bool

	for i, component := range pathComponents {
		if component == "" {
			// 保留空组件（如Unix绝对路径的开头）
			truncatedComponents = append(truncatedComponents, component)
			continue
		}

		// 检查组件长度是否超过255字节
		if len(component) > 255 {
			truncatedComponent := s.truncateFileName(component)
			truncatedComponents = append(truncatedComponents, truncatedComponent)
			pathChanged = true

			s.logger.Debug("截断路径组件",
				zap.String("原组件", component),
				zap.String("截断后组件", truncatedComponent),
				zap.Int("原长度", len(component)),
				zap.Int("截断后长度", len(truncatedComponent)),
				zap.Int("组件索引", i))
		} else {
			truncatedComponents = append(truncatedComponents, component)
		}
	}

	// 重新组装路径
	newPath := strings.Join(truncatedComponents, string(filepath.Separator))

	if pathChanged {
		s.logger.Info("路径组件已截断",
			zap.String("原路径", fullPath),
			zap.String("新路径", newPath),
			zap.Int("原长度", len(fullPath)),
			zap.Int("新长度", len(newPath)))
	}

	// 检查整体路径长度限制
	var maxPathLength int
	switch runtime.GOOS {
	case "windows":
		maxPathLength = 260
	case "linux":
		maxPathLength = 4096
	case "darwin": // macOS
		maxPathLength = 1024
	default:
		maxPathLength = 260 // 默认使用最保守的限制
	}

	// 如果整体路径仍然过长，进行进一步处理
	if len(newPath) > maxPathLength {
		return s.truncateDirectoryPath(filepath.Dir(newPath), filepath.Base(newPath), maxPathLength)
	}

	return newPath
}

// truncateDirectoryPath 截断目录路径以适应路径长度限制
func (s *StrmGeneratorService) truncateDirectoryPath(dir, fileName string, maxPathLength int) string {
	// 为文件名和路径分隔符预留空间
	reservedLength := len(fileName) + 1 // +1 for path separator
	maxDirLength := maxPathLength - reservedLength

	if maxDirLength <= 0 {
		// 如果文件名本身就太长，只能使用根目录
		s.logger.Error("文件名过长，无法创建有效路径",
			zap.String("文件名", fileName),
			zap.Int("文件名长度", len(fileName)),
			zap.Int("最大路径长度", maxPathLength))
		return fileName
	}

	// 如果目录路径在限制内，直接返回
	if len(dir) <= maxDirLength {
		return filepath.Join(dir, fileName)
	}

	// 需要截断目录路径
	// 策略：保留路径的开头和结尾部分，中间用哈希替代
	pathParts := strings.Split(dir, string(filepath.Separator))

	// 如果只有一个目录层级，直接截断
	if len(pathParts) <= 1 {
		truncatedDir := dir[:maxDirLength]
		result := filepath.Join(truncatedDir, fileName)
		s.logger.Debug("截断单层目录路径",
			zap.String("原目录", dir),
			zap.String("截断后目录", truncatedDir),
			zap.String("最终路径", result))
		return result
	}

	// 多层目录：保留第一层和最后一层，中间层用哈希替代
	firstPart := pathParts[0]
	lastPart := pathParts[len(pathParts)-1]

	// 计算中间部分的哈希
	middleParts := strings.Join(pathParts[1:len(pathParts)-1], string(filepath.Separator))
	hash := fmt.Sprintf("%08x", hashString(middleParts))
	hashDir := "_" + hash[:8]

	// 构建新的目录路径
	var newDir string
	if firstPart == "" {
		// Unix 绝对路径
		newDir = string(filepath.Separator) + filepath.Join(hashDir, lastPart)
	} else {
		newDir = filepath.Join(firstPart, hashDir, lastPart)
	}

	// 检查新路径长度
	newPath := filepath.Join(newDir, fileName)
	if len(newPath) <= maxPathLength {
		s.logger.Debug("成功截断多层目录路径",
			zap.String("原目录", dir),
			zap.String("新目录", newDir),
			zap.String("最终路径", newPath),
			zap.String("哈希目录", hashDir))
		return newPath
	}

	// 如果还是太长，进一步截断最后一层目录
	maxLastPartLength := maxDirLength - len(firstPart) - len(hashDir) - 2 // -2 for separators
	if maxLastPartLength > 0 && len(lastPart) > maxLastPartLength {
		truncatedLastPart := lastPart[:maxLastPartLength]
		if firstPart == "" {
			newDir = string(filepath.Separator) + filepath.Join(hashDir, truncatedLastPart)
		} else {
			newDir = filepath.Join(firstPart, hashDir, truncatedLastPart)
		}
		newPath = filepath.Join(newDir, fileName)

		s.logger.Debug("进一步截断最后一层目录",
			zap.String("原最后层", lastPart),
			zap.String("截断后最后层", truncatedLastPart),
			zap.String("最终路径", newPath))
	}

	return newPath
}

// safeMkdirAll 安全创建目录，处理目录名长度限制
func (s *StrmGeneratorService) safeMkdirAll(path string, perm os.FileMode) error {
	// 如果路径为空或已存在，直接返回
	if path == "" {
		return nil
	}

	// 检查目录是否已存在
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return nil
	}

	// 处理路径长度和目录名长度
	safePath := s.truncatePathLength(path)

	// 分解路径，逐级检查和创建目录
	return s.createDirectoryRecursively(safePath, perm)
}

// createDirectoryRecursively 递归创建目录，确保每个目录名都符合长度限制
func (s *StrmGeneratorService) createDirectoryRecursively(path string, perm os.FileMode) error {
	// 获取父目录
	parent := filepath.Dir(path)

	// 如果父目录不是根目录且不存在，先创建父目录
	if parent != path && parent != "." && parent != "/" {
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			if err := s.createDirectoryRecursively(parent, perm); err != nil {
				return err
			}
		}
	}

	// 检查当前目录名长度
	dirName := filepath.Base(path)
	if len(dirName) > 255 {
		// 目录名过长，需要截断
		truncatedDirName := s.truncateFileName(dirName)
		newPath := filepath.Join(parent, truncatedDirName)

		s.logger.Warn("目录名过长，已截断",
			zap.String("原目录名", dirName),
			zap.String("截断后目录名", truncatedDirName),
			zap.String("原路径", path),
			zap.String("新路径", newPath))

		path = newPath
	}

	// 创建目录
	if err := os.Mkdir(path, perm); err != nil && !os.IsExist(err) {
		return fmt.Errorf("创建目录失败 [%s]: %w", path, err)
	}

	return nil
}

// hashString 计算字符串的简单哈希值
func hashString(s string) uint32 {
	h := uint32(0)
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
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

		// 使用优化的URL编码（带缓存）
		dirPath, fileName = s.optimizedURLEncode(dirPath, fileName, strmConfig.URLEncode)
		if strmConfig.URLEncode {
			s.logger.Debug("进行了URL编码",
				zap.String("原路径", filepath.Dir(sourcePath)),
				zap.String("编码后路径", dirPath),
				zap.String("原文件名", file.Name),
				zap.String("编码后文件名", fileName))
		}

		fileURL = s.alistService.GetFileURL(dirPath, fileName, file.Sign)
	case "clouddrive":
		// 处理路径和文件名的 URL 编码
		dirPath := filepath.Dir(sourcePath)
		fileName := file.Name

		// 使用优化的URL编码（带缓存）
		dirPath, fileName = s.optimizedURLEncode(dirPath, fileName, strmConfig.URLEncode)
		if strmConfig.URLEncode {
			s.logger.Debug("进行了URL编码",
				zap.String("原路径", filepath.Dir(sourcePath)),
				zap.String("编码后路径", dirPath),
				zap.String("原文件名", file.Name),
				zap.String("编码后文件名", fileName))
		}

		fileURL = s.alistService.GetFileURL(dirPath, fileName, file.Sign)
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

	// 检查并处理路径长度，包括目录名和文件名
	strmFilePath = s.truncatePathLength(strmFilePath)

	// 检查是否需要覆盖现有文件
	if !s.shouldOverwrite(strmFilePath, taskConfig) {
		return false, "文件已存在且不允许覆盖", strmFilePath
	}

	// 确保目标目录存在
	if err := s.safeMkdirAll(filepath.Dir(strmFilePath), 0755); err != nil {
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

	// 检查并处理路径长度，包括目录名和文件名
	originalPath := targetPath
	targetPath = s.truncatePathLength(targetPath)

	// 如果路径被截断，记录日志
	if targetPath != originalPath {
		s.logger.Debug("下载文件路径已更新",
			zap.String("原路径", originalPath),
			zap.String("新路径", targetPath))
	}

	// 获取目标目录路径
	targetDir := filepath.Dir(targetPath)

	// 确保目标目录存在
	if err := s.safeMkdirAll(targetDir, 0755); err != nil {
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

	// 使用优化的URL编码（带缓存）
	dirPath, fileName = s.optimizedURLEncode(dirPath, fileName, strmConfig.URLEncode)

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

// 优化的HTTP客户端，复用连接
var optimizedHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	},
}

// downloadFileFromURL 从 URL 下载文件（优化版本）
func (s *StrmGeneratorService) downloadFileFromURL(fileURL, targetPath string) error {
	// 发送 GET 请求，使用优化的客户端
	resp, err := optimizedHTTPClient.Get(fileURL)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载文件失败，状态码: %d", resp.StatusCode)
	}

	// 确保目标目录存在
	if err := s.safeMkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 创建目标文件
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer file.Close()

	// 使用缓冲区复制内容，提高I/O效率
	buffer := make([]byte, 32*1024) // 32KB缓冲区
	_, err = io.CopyBuffer(file, resp.Body, buffer)
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

// FileHistoryBatch 批量文件历史记录
type FileHistoryBatch struct {
	records []filehistory.FileHistory
	mutex   sync.Mutex
}

// recordFileHistory 记录文件历史（优化版本，支持批量处理）
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
			s.logger.Debug("更新现有文件历史记录",
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
			s.logger.Debug("文件历史记录已存在（数据库约束），跳过创建",
				zap.String("fileName", file.Name),
				zap.String("sourcePath", sourcePath),
				zap.String("hash", hash))
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
		// 添加随机延迟(0.5-2秒)，防止网盘风控，优化延迟时间
		if i > 0 {
			// 使用更高效的随机数生成，减少延迟时间
			randomDelay := time.Duration(500+(i*47)%1500) * time.Millisecond // 使用简单的伪随机
			if i%10 == 0 {                                                   // 每10个文件记录一次延迟日志
				s.logger.Debug("等待随机延迟", zap.Duration("delay", randomDelay))
			}
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

		// 每处理 50 个文件更新一次数据库，减少数据库操作频率
		if (i+1)%50 == 0 {
			s.stats.Mutex.RLock()
			// 计算数据库中需要的汇总数值
			subtitleCount := s.stats.SubtitleDownloaded + s.stats.SubtitleSkipped
			metadataCount := s.stats.MetadataDownloaded + s.stats.MetadataSkipped
			// 跳过文件数只包含STRM文件跳过和其他文件跳过，元数据和字幕有单独的统计字段
			skipFileCount := s.stats.SkipFile + s.stats.OtherSkipped

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

// processStrmFileQueueAsync 异步处理STRM文件队列（高级并发处理）
func (s *StrmGeneratorService) processStrmFileQueueAsync(taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint, scanDoneChan chan bool) error {
	// 动态计算最优并发数
	concurrency := s.calculateOptimalConcurrency()

	s.logger.Info("启动高级STRM文件异步处理协程",
		zap.Int("并发数", concurrency),
		zap.Int("CPU核心数", runtime.NumCPU()))

	// 创建工作窃取队列系统
	workerPool := s.createWorkerPool(concurrency)

	// 启动高性能工作池
	workerPool.startWorkers(s, taskInfo, strmConfig, taskLogID)

	// 启动高性能结果收集协程
	resultCollector := s.createResultCollector(workerPool.resultChan, taskInfo, taskLogID)
	go resultCollector.start()

	// 启动智能队列分发器
	dispatcher := s.createSmartDispatcher(workerPool, scanDoneChan, concurrency)
	go dispatcher.start()

	// 等待所有处理完成
	dispatcher.wait()

	// 等待结果收集器完成
	resultCollector.wait()

	// 标记 STRM 处理完成
	s.stats.Mutex.Lock()
	s.stats.StrmProcessingDone = true
	s.stats.Mutex.Unlock()

	// 清理URL编码缓存，释放内存
	s.urlEncodeCache.Clear()

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

// URLEncodeCache URL编码缓存
type URLEncodeCache struct {
	cache sync.Map // 使用sync.Map提供并发安全的缓存
}

// NewURLEncodeCache 创建新的URL编码缓存
func NewURLEncodeCache() *URLEncodeCache {
	return &URLEncodeCache{}
}

// GetEncodedPath 获取编码后的路径，带缓存
func (c *URLEncodeCache) GetEncodedPath(path string) string {
	if cached, ok := c.cache.Load(path); ok {
		return cached.(string)
	}

	// 编码路径
	encoded := c.encodePath(path)
	c.cache.Store(path, encoded)
	return encoded
}

// GetEncodedFileName 获取编码后的文件名，带缓存
func (c *URLEncodeCache) GetEncodedFileName(fileName string) string {
	if cached, ok := c.cache.Load(fileName); ok {
		return cached.(string)
	}

	encoded := url.PathEscape(fileName)
	c.cache.Store(fileName, encoded)
	return encoded
}

// encodePath 编码路径的内部方法
func (c *URLEncodeCache) encodePath(dirPath string) string {
	encodedDirPath := filepath.ToSlash(dirPath) // 统一使用正斜杠
	if encodedDirPath == "." || encodedDirPath == "" {
		return encodedDirPath
	}

	pathParts := strings.Split(encodedDirPath, "/")
	for i, part := range pathParts {
		if part != "" { // 跳过空字符串
			pathParts[i] = url.PathEscape(part)
		}
	}
	return strings.Join(pathParts, "/")
}

// Clear 清理缓存（可选，用于内存管理）
func (c *URLEncodeCache) Clear() {
	c.cache = sync.Map{}
}

// optimizedURLEncode 优化的URL编码函数，使用缓存减少重复计算
func (s *StrmGeneratorService) optimizedURLEncode(dirPath, fileName string, needEncode bool) (string, string) {
	if !needEncode {
		return dirPath, fileName
	}

	// 使用缓存获取编码结果
	encodedDirPath := s.urlEncodeCache.GetEncodedPath(dirPath)
	encodedFileName := s.urlEncodeCache.GetEncodedFileName(fileName)

	return encodedDirPath, encodedFileName
}

// calculateOptimalConcurrency 动态计算最优并发数
func (s *StrmGeneratorService) calculateOptimalConcurrency() int {
	cpuCount := runtime.NumCPU()

	// 基于CPU核心数和系统负载动态调整
	baseConcurrency := cpuCount * 4 // I/O密集型任务，可以超过CPU核心数

	// 根据可用内存调整
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 如果内存使用率过高，降低并发数
	if memStats.Sys > 1024*1024*1024 { // 超过1GB系统内存使用
		baseConcurrency = cpuCount * 2
	}

	// 设置合理的范围
	if baseConcurrency < 10 {
		baseConcurrency = 10
	} else if baseConcurrency > 50 {
		baseConcurrency = 50
	}

	return baseConcurrency
}

// WorkerPool 高性能工作池
type WorkerPool struct {
	workers    []*Worker
	resultChan chan FileProcessResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	closed     int32         // 使用原子操作防止重复关闭
	initDone   chan struct{} // 初始化完成信号
}

// Worker 工作协程
type Worker struct {
	id         int
	localQueue []FileEntry // 本地队列，实现工作窃取
	mutex      sync.RWMutex
	service    *StrmGeneratorService
	taskInfo   *task.Task
	strmConfig *StrmConfig
	taskLogID  uint
	resultChan chan FileProcessResult // 结果通道引用
}

// createWorkerPool 创建高性能工作池
func (s *StrmGeneratorService) createWorkerPool(concurrency int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:    make([]*Worker, concurrency),
		resultChan: make(chan FileProcessResult, concurrency*10),
		ctx:        ctx,
		cancel:     cancel,
		initDone:   make(chan struct{}),
	}

	return pool
}

// startWorkers 启动工作协程
func (pool *WorkerPool) startWorkers(service *StrmGeneratorService, taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint) {
	// 第一步：创建所有Worker实例
	for i := 0; i < len(pool.workers); i++ {
		worker := &Worker{
			id:         i,
			localQueue: make([]FileEntry, 0, 20), // 增加本地队列容量
			service:    service,
			taskInfo:   taskInfo,
			strmConfig: strmConfig,
			taskLogID:  taskLogID,
			resultChan: pool.resultChan,
		}
		pool.workers[i] = worker
	}

	// 发送初始化完成信号
	close(pool.initDone)

	// 第二步：启动所有协程（此时所有Worker都已初始化）
	for i := 0; i < len(pool.workers); i++ {
		pool.wg.Add(1)
		go pool.workers[i].run(pool.ctx, &pool.wg, pool.workers, pool.initDone)
	}
}

// run 工作协程主循环，支持工作窃取
func (w *Worker) run(ctx context.Context, wg *sync.WaitGroup, allWorkers []*Worker, initDone chan struct{}) {
	defer wg.Done()

	// 等待所有Worker初始化完成
	<-initDone

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 尝试从本地队列获取任务
			if job := w.getLocalJob(); job != nil {
				w.processJob(*job)
			} else {
				// 尝试从其他工作协程窃取任务
				if stolenJob := w.stealWork(allWorkers); stolenJob != nil {
					w.processJob(*stolenJob)
				} else {
					// 没有任务，短暂休眠
					time.Sleep(time.Millisecond)
				}
			}
		}
	}
}

// processJob 处理单个任务
func (w *Worker) processJob(entry FileEntry) {
	processed := w.service.processFile(
		entry.File,
		entry.FileType,
		w.taskInfo,
		w.strmConfig,
		w.taskLogID,
		entry.SourcePath,
		entry.TargetPath,
	)

	w.resultChan <- FileProcessResult{
		Entry:     entry,
		Processed: processed,
		FileType:  entry.FileType,
		Success:   processed.Success,
	}
}

// addLocalJob 添加任务到本地队列
func (w *Worker) addLocalJob(job FileEntry) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.localQueue = append(w.localQueue, job)
}

// getLocalJob 从本地队列获取任务
func (w *Worker) getLocalJob() *FileEntry {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.localQueue) == 0 {
		return nil
	}

	job := w.localQueue[0]
	w.localQueue = w.localQueue[1:]
	return &job
}

// stealWork 从其他工作协程窃取任务
func (w *Worker) stealWork(allWorkers []*Worker) *FileEntry {
	for _, worker := range allWorkers {
		// 添加 nil 检查，防止竞态条件
		if worker == nil || worker.id == w.id {
			continue
		}

		worker.mutex.Lock()
		if len(worker.localQueue) > 1 { // 只有当对方有多个任务时才窃取
			// 从队列末尾窃取任务
			job := worker.localQueue[len(worker.localQueue)-1]
			worker.localQueue = worker.localQueue[:len(worker.localQueue)-1]
			worker.mutex.Unlock()
			return &job
		}
		worker.mutex.Unlock()
	}
	return nil
}

// Close 关闭工作池
func (pool *WorkerPool) Close() {
	// 使用原子操作防止重复关闭
	if !atomic.CompareAndSwapInt32(&pool.closed, 0, 1) {
		return // 已经关闭过了
	}

	pool.cancel()
	pool.wg.Wait()
	close(pool.resultChan)
}

// ResultCollector 高性能结果收集器
type ResultCollector struct {
	resultChan     chan FileProcessResult
	service        *StrmGeneratorService
	taskInfo       *task.Task
	taskLogID      uint
	batchSize      int
	updateInterval time.Duration
	lastUpdateTime time.Time
	pendingResults []FileProcessResult
	mutex          sync.Mutex
	done           chan struct{} // 用于通知收集器完成
}

// createResultCollector 创建结果收集器
func (s *StrmGeneratorService) createResultCollector(resultChan chan FileProcessResult, taskInfo *task.Task, taskLogID uint) *ResultCollector {
	return &ResultCollector{
		resultChan:     resultChan,
		service:        s,
		taskInfo:       taskInfo,
		taskLogID:      taskLogID,
		batchSize:      100,             // 批量处理大小
		updateInterval: 5 * time.Second, // 更新间隔
		lastUpdateTime: time.Now(),
		pendingResults: make([]FileProcessResult, 0, 100),
		done:           make(chan struct{}), // 完成通知通道
	}
}

// start 启动结果收集
func (rc *ResultCollector) start() {
	ticker := time.NewTicker(rc.updateInterval)
	defer ticker.Stop()
	defer close(rc.done) // 结束时发送完成信号

	for {
		select {
		case result, ok := <-rc.resultChan:
			if !ok {
				// 处理剩余结果
				rc.processPendingResults()
				return
			}
			rc.addResult(result)

		case <-ticker.C:
			// 定期处理批量结果
			rc.processPendingResults()
		}
	}
}

// wait 等待结果收集器完成
func (rc *ResultCollector) wait() {
	<-rc.done
}

// addResult 添加结果到批处理队列
func (rc *ResultCollector) addResult(result FileProcessResult) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	rc.pendingResults = append(rc.pendingResults, result)

	// 如果达到批处理大小，立即处理
	if len(rc.pendingResults) >= rc.batchSize {
		rc.processBatch()
	}
}

// processPendingResults 处理待处理的结果
func (rc *ResultCollector) processPendingResults() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	if len(rc.pendingResults) > 0 {
		rc.processBatch()
	}
}

// processBatch 批量处理结果
func (rc *ResultCollector) processBatch() {
	if len(rc.pendingResults) == 0 {
		return
	}

	// 批量处理文件历史记录
	var successResults []FileProcessResult
	var generatedCount, skippedCount int

	for _, result := range rc.pendingResults {
		// 记录文件历史
		rc.service.recordFileHistory(
			rc.taskInfo.ID,
			rc.taskLogID,
			result.Entry.File,
			result.Entry.SourcePath,
			result.Processed.TargetPath,
			result.FileType,
			result.Success,
		)

		if result.Success {
			generatedCount++
			successResults = append(successResults, result)
		} else {
			skippedCount++
		}
	}

	// 批量更新统计信息
	rc.service.stats.Mutex.Lock()
	rc.service.stats.GeneratedFile += generatedCount
	rc.service.stats.SkipFile += skippedCount
	totalGenerated := rc.service.stats.GeneratedFile
	totalSkipped := rc.service.stats.SkipFile
	rc.service.stats.Mutex.Unlock()

	// 更新数据库
	if time.Since(rc.lastUpdateTime) > rc.updateInterval {
		rc.updateDatabase(totalGenerated, totalSkipped)
		rc.lastUpdateTime = time.Now()
	}

	// 清空待处理队列
	rc.pendingResults = rc.pendingResults[:0]

	rc.service.logger.Debug("批量处理结果完成",
		zap.Int("成功", generatedCount),
		zap.Int("跳过", skippedCount),
		zap.Int("总成功", totalGenerated),
		zap.Int("总跳过", totalSkipped))
}

// updateDatabase 更新数据库
func (rc *ResultCollector) updateDatabase(totalGenerated, totalSkipped int) {
	rc.service.stats.Mutex.RLock()
	// 跳过文件数只包含STRM文件跳过和其他文件跳过，元数据和字幕有单独的统计字段
	skipFileCount := rc.service.stats.SkipFile + rc.service.stats.OtherSkipped
	rc.service.stats.Mutex.RUnlock()

	updateData := map[string]interface{}{
		"generated_file": totalGenerated,
		"skip_file":      skipFileCount,
	}

	if err := repository.TaskLog.UpdatePartial(rc.taskLogID, updateData); err != nil {
		rc.service.logger.Error("批量更新任务日志失败", zap.Error(err))
	}
}

// BatchFileChecker 批量文件检查器
type BatchFileChecker struct {
	batchSize      int
	maxConcurrency int
}

// NewBatchFileChecker 创建批量文件检查器
func NewBatchFileChecker() *BatchFileChecker {
	return &BatchFileChecker{
		batchSize:      100, // 每批处理100个文件
		maxConcurrency: 20,  // 最大并发数
	}
}

// CheckExistence 批量检查文件存在性，优化大量文件的检查效率
func (checker *BatchFileChecker) CheckExistence(entries []FileEntry) map[string]bool {
	existenceMap := make(map[string]bool, len(entries))

	// 如果文件数量较少，直接并发检查
	if len(entries) <= checker.batchSize {
		return checker.concurrentCheck(entries)
	}

	// 大量文件时，分批处理以控制内存和系统资源使用
	for i := 0; i < len(entries); i += checker.batchSize {
		end := i + checker.batchSize
		if end > len(entries) {
			end = len(entries)
		}

		batch := entries[i:end]
		batchResult := checker.concurrentCheck(batch)

		// 合并结果
		for path, exists := range batchResult {
			existenceMap[path] = exists
		}

		// 短暂休眠，避免过度占用系统资源
		if i+checker.batchSize < len(entries) {
			time.Sleep(time.Millisecond)
		}
	}

	return existenceMap
}

// concurrentCheck 并发检查一批文件
func (checker *BatchFileChecker) concurrentCheck(entries []FileEntry) map[string]bool {
	existenceMap := make(map[string]bool, len(entries))
	semaphore := make(chan struct{}, checker.maxConcurrency)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, entry := range entries {
		wg.Add(1)
		go func(targetPath string) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 优化的文件存在性检查
			exists := checker.fastFileExists(targetPath)

			mutex.Lock()
			existenceMap[targetPath] = exists
			mutex.Unlock()
		}(entry.TargetPath)
	}

	wg.Wait()
	return existenceMap
}

// fastFileExists 快速文件存在性检查，减少系统调用开销
func (checker *BatchFileChecker) fastFileExists(filePath string) bool {
	// 使用 os.Lstat 而不是 os.Stat，避免跟随符号链接，提高性能
	_, err := os.Lstat(filePath)
	return err == nil
}

// batchCheckFileExistence 批量检查文件是否存在，提高I/O效率
func (s *StrmGeneratorService) batchCheckFileExistence(entries []FileEntry) map[string]bool {
	checker := NewBatchFileChecker()
	return checker.CheckExistence(entries)
}

// SmartDispatcher 智能任务分发器
type SmartDispatcher struct {
	service       *StrmGeneratorService
	workerPool    *WorkerPool
	scanDoneChan  chan bool
	concurrency   int
	batchSize     int
	adaptiveBatch bool
	wg            sync.WaitGroup
}

// createSmartDispatcher 创建智能分发器
func (s *StrmGeneratorService) createSmartDispatcher(workerPool *WorkerPool, scanDoneChan chan bool, concurrency int) *SmartDispatcher {
	return &SmartDispatcher{
		service:       s,
		workerPool:    workerPool,
		scanDoneChan:  scanDoneChan,
		concurrency:   concurrency,
		batchSize:     concurrency * 2, // 动态批次大小
		adaptiveBatch: true,
	}
}

// start 启动智能分发
func (sd *SmartDispatcher) start() {
	sd.wg.Add(1)
	defer sd.wg.Done()

	scanDone := false
	idleCount := 0

	for !scanDone || sd.service.queue.hasStrmFiles() {
		queueSize := sd.getQueueSize()

		if queueSize > 0 {
			// 动态调整批次大小
			batchSize := sd.calculateBatchSize(queueSize, idleCount)

			// 获取并分发任务
			batch := sd.service.queue.getAndRemoveStrmFileBatch(batchSize)
			if len(batch) > 0 {
				sd.distributeTasks(batch)
				idleCount = 0 // 重置空闲计数
			}
		} else {
			idleCount++
		}

		// 检查扫描状态
		select {
		case _, ok := <-sd.scanDoneChan:
			if !ok {
				scanDone = true
			}
		default:
			// 自适应休眠时间
			sleepTime := sd.calculateSleepTime(queueSize, idleCount)
			time.Sleep(sleepTime)
		}
	}

	// 关闭工作池
	sd.workerPool.Close()
}

// getQueueSize 获取队列大小
func (sd *SmartDispatcher) getQueueSize() int {
	sd.service.queue.FilesMutex.RLock()
	defer sd.service.queue.FilesMutex.RUnlock()
	return len(sd.service.queue.StrmFiles)
}

// calculateBatchSize 动态计算批次大小
func (sd *SmartDispatcher) calculateBatchSize(queueSize, idleCount int) int {
	if !sd.adaptiveBatch {
		return sd.batchSize
	}

	// 基于队列大小和空闲次数动态调整
	baseBatch := sd.concurrency * 2

	if queueSize > 1000 {
		// 大队列，增加批次大小
		baseBatch = sd.concurrency * 4
	} else if queueSize < 100 {
		// 小队列，减少批次大小
		baseBatch = sd.concurrency
	}

	// 如果系统空闲，可以增加批次大小
	if idleCount > 10 {
		baseBatch = baseBatch * 2
	}

	// 限制最大批次大小
	if baseBatch > 500 {
		baseBatch = 500
	}

	return baseBatch
}

// distributeTasks 智能分发任务
func (sd *SmartDispatcher) distributeTasks(batch []FileEntry) {
	// 负载均衡分发到工作协程
	workerCount := len(sd.workerPool.workers)

	for i, entry := range batch {
		workerIndex := i % workerCount
		sd.workerPool.workers[workerIndex].addLocalJob(entry)
	}

	sd.service.logger.Debug("智能分发任务完成",
		zap.Int("任务数", len(batch)),
		zap.Int("工作协程数", workerCount))
}

// calculateSleepTime 计算自适应休眠时间
func (sd *SmartDispatcher) calculateSleepTime(queueSize, idleCount int) time.Duration {
	if queueSize == 0 {
		// 队列为空，逐渐增加休眠时间
		sleepTime := time.Duration(10+idleCount*5) * time.Millisecond
		if sleepTime > 100*time.Millisecond {
			sleepTime = 100 * time.Millisecond
		}
		return sleepTime
	}

	// 有任务时，短暂休眠
	return 10 * time.Millisecond
}

// wait 等待分发完成
func (sd *SmartDispatcher) wait() {
	sd.wg.Wait()
}
