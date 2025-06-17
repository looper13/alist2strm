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

// processDirectory 处理目录（递归）
func (s *StrmGeneratorService) processDirectory(taskInfo *task.Task, strmConfig *StrmConfig, taskLogID uint, sourcePath, targetPath string) (int, int, int, int, int, error) {
	totalFiles := 0
	generatedFiles := 0
	skippedFiles := 0
	metadataProcessedCount := 0
	subtitleProcessedCount := 0

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

	// 第一阶段：收集和分类所有文件
	var mediaFileEntries []FileEntry
	var subtitleFileEntries []FileEntry
	var metadataFileEntries []FileEntry
	var otherFileEntries []FileEntry
	var directoryFiles []*AListFile

	// 先对文件进行分类
	for _, file := range files {
		totalFiles++

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
			otherFileEntries = append(otherFileEntries, entry)
		}
	}
	// 第二阶段：筛选需要处理的字幕文件（需要与媒体文件匹配）
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
			skippedFiles++
		}
	}

	// 第三阶段：只对STRM文件（媒体文件）使用并发生成
	if len(mediaFileEntries) > 0 {
		// 计算需要处理的STRM文件数
		totalStrmFiles := len(mediaFileEntries)
		s.logger.Info("准备并发生成STRM文件",
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

		// 提交任务
		go func() {
			// 提交媒体文件
			for _, entry := range mediaFileEntries {
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
			if result.Success {
				generatedFiles++
			} else {
				skippedFiles++
			}
		}
	}

	// 第四阶段：串行处理字幕和元数据文件（下载任务）
	s.logger.Info("开始串行处理下载任务",
		zap.Int("匹配字幕", len(matchedSubtitleEntries)),
		zap.Int("元数据文件", len(metadataFileEntries)))

	// 处理字幕文件
	for _, entry := range matchedSubtitleEntries {
		processed := s.processFile(entry.File, entry.FileType, taskInfo, strmConfig, taskLogID, entry.SourcePath, entry.TargetPath)
		s.recordFileHistory(taskInfo.ID, taskLogID, entry.File, entry.SourcePath, processed.TargetPath, entry.FileType, processed.Success)

		if processed.Success {
			subtitleProcessedCount++
		} else {
			skippedFiles++
		}
	}

	// 处理元数据文件
	for _, entry := range metadataFileEntries {
		processed := s.processFile(entry.File, entry.FileType, taskInfo, strmConfig, taskLogID, entry.SourcePath, entry.TargetPath)
		s.recordFileHistory(taskInfo.ID, taskLogID, entry.File, entry.SourcePath, processed.TargetPath, entry.FileType, processed.Success)

		if processed.Success {
			metadataProcessedCount++
		} else {
			skippedFiles++
		}
	}

	// 第五阶段：处理其他文件
	for range otherFileEntries {
		skippedFiles++
	}

	// 最后递归处理目录
	for _, dirFile := range directoryFiles {
		currentSourcePath := filepath.Join(sourcePath, dirFile.Name)
		currentTargetPath := filepath.Join(targetPath, dirFile.Name)

		subTotal, subGenerated, subSkipped, subMetadata, subSubtitle, err := s.processDirectory(
			taskInfo, strmConfig, taskLogID, currentSourcePath, currentTargetPath)
		if err != nil {
			return totalFiles, generatedFiles, skippedFiles, metadataProcessedCount, subtitleProcessedCount, err
		}

		totalFiles += subTotal
		generatedFiles += subGenerated
		skippedFiles += subSkipped
		metadataProcessedCount += subMetadata
		subtitleProcessedCount += subSubtitle
	}

	return totalFiles, generatedFiles, skippedFiles, metadataProcessedCount, subtitleProcessedCount, nil
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
	// 添加下载前延迟，避免网盘风控
	time.Sleep(300 * time.Millisecond)

	// 检查目标文件是否已存在
	if !s.shouldOverwrite(targetPath, taskConfig) {
		return false, "文件已存在且不允许覆盖"
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
