package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ValidationService struct {
	db                  *gorm.DB
	logger              *zap.Logger
	fileHistoryService  *FileHistoryService
	notificationService *NotificationService
	httpClient          *http.Client
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	Enabled            bool `json:"enabled"`
	CheckInterval      int  `json:"checkInterval"`      // 检查间隔（小时）
	BatchSize          int  `json:"batchSize"`          // 批处理大小
	TimeoutSeconds     int  `json:"timeoutSeconds"`     // 超时时间（秒）
	MaxRetries         int  `json:"maxRetries"`         // 最大重试次数
	EnableNotification bool `json:"enableNotification"` // 启用通知
	AutoCleanInvalid   bool `json:"autoCleanInvalid"`   // 自动清理失效文件
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid      bool   `json:"isValid"`
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
	ResponseTime int64  `json:"responseTime"` // 毫秒
}

// ValidationTaskProgress 验证任务进度
type ValidationTaskProgress struct {
	TaskID         uint                       `json:"taskId"`
	Status         model.ValidationTaskStatus `json:"status"`
	TotalFiles     int                        `json:"totalFiles"`
	ProcessedFiles int                        `json:"processedFiles"`
	ValidFiles     int                        `json:"validFiles"`
	InvalidFiles   int                        `json:"invalidFiles"`
	Progress       int                        `json:"progress"`
	StartedAt      *time.Time                 `json:"startedAt"`
	EstimatedEnd   *time.Time                 `json:"estimatedEnd"`
}

var (
	validationService *ValidationService
	validationOnce    sync.Once
)

// GetValidationService 获取 ValidationService 单例
func GetValidationService() *ValidationService {
	validationOnce.Do(func() {
		validationService = &ValidationService{
			db:                  model.DB,
			logger:              utils.Logger,
			fileHistoryService:  GetFileHistoryService(),
			notificationService: GetNotificationService(),
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	})
	return validationService
}

// StartValidationTask 启动验证任务
func (s *ValidationService) StartValidationTask(req *model.ValidationTaskCreateRequest) (*model.ValidationTask, error) {
	// 检查是否有正在运行的任务
	var runningTask model.ValidationTask
	err := s.db.Where("status = ?", model.ValidationTaskStatusRunning).First(&runningTask).Error
	if err == nil {
		return nil, fmt.Errorf("已有验证任务正在运行，任务ID: %d", runningTask.ID)
	}

	// 创建新的验证任务
	task := &model.ValidationTask{
		Type:           req.Type,
		Status:         model.ValidationTaskStatusPending,
		Config:         req.Config,
		TotalFiles:     0,
		ProcessedFiles: 0,
		ValidFiles:     0,
		InvalidFiles:   0,
		Progress:       0,
	}

	if err := s.db.Create(task).Error; err != nil {
		s.logger.Error("创建验证任务失败", zap.Error(err))
		return nil, err
	}

	// 异步执行验证任务
	go s.executeValidationTask(task.ID)

	s.logger.Info("验证任务创建成功", zap.Uint("taskId", task.ID), zap.String("type", string(task.Type)))
	return task, nil
}

// executeValidationTask 执行验证任务
func (s *ValidationService) executeValidationTask(taskID uint) {
	// 更新任务状态为运行中
	now := time.Now()
	s.db.Model(&model.ValidationTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     model.ValidationTaskStatusRunning,
		"started_at": &now,
	})

	// 获取需要验证的文件列表
	files, err := s.getFilesToValidate(taskID)
	if err != nil {
		s.updateTaskStatus(taskID, model.ValidationTaskStatusFailed, fmt.Sprintf("获取文件列表失败: %v", err))
		return
	}

	totalFiles := len(files)
	s.db.Model(&model.ValidationTask{}).Where("id = ?", taskID).Update("total_files", totalFiles)

	if totalFiles == 0 {
		s.updateTaskStatus(taskID, model.ValidationTaskStatusCompleted, "没有需要验证的文件")
		return
	}

	s.logger.Info("开始执行验证任务", zap.Uint("taskId", taskID), zap.Int("totalFiles", totalFiles))

	// 执行验证
	validCount := 0
	invalidCount := 0
	invalidFiles := make([]map[string]interface{}, 0)

	for i, file := range files {
		// 验证文件
		result := s.validateStrmFile(file.TargetFilePath)

		// 更新文件历史记录
		s.fileHistoryService.MarkAsValidated(file.ID, result.IsValid, result.ErrorMessage)

		// 统计结果
		if result.IsValid {
			validCount++
		} else {
			invalidCount++
			invalidFiles = append(invalidFiles, map[string]interface{}{
				"id":           file.ID,
				"fileName":     file.FileName,
				"sourcePath":   file.SourcePath,
				"targetPath":   file.TargetFilePath,
				"errorMessage": result.ErrorMessage,
			})
		}

		// 更新进度
		processedFiles := i + 1
		progress := int(float64(processedFiles) / float64(totalFiles) * 100)
		s.db.Model(&model.ValidationTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"processed_files": processedFiles,
			"valid_files":     validCount,
			"invalid_files":   invalidCount,
			"progress":        progress,
		})

		// 每处理100个文件记录一次日志
		if processedFiles%100 == 0 || processedFiles == totalFiles {
			s.logger.Info("验证任务进度",
				zap.Uint("taskId", taskID),
				zap.Int("processed", processedFiles),
				zap.Int("total", totalFiles),
				zap.Int("progress", progress))
		}
	}

	// 完成任务
	completedAt := time.Now()
	s.db.Model(&model.ValidationTask{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":       model.ValidationTaskStatusCompleted,
		"completed_at": &completedAt,
		"message":      fmt.Sprintf("验证完成：共 %d 个文件，有效 %d 个，失效 %d 个", totalFiles, validCount, invalidCount),
	})

	s.logger.Info("验证任务完成",
		zap.Uint("taskId", taskID),
		zap.Int("totalFiles", totalFiles),
		zap.Int("validFiles", validCount),
		zap.Int("invalidFiles", invalidCount))

	// 发送通知
	if invalidCount > 0 {
		s.sendInvalidFileNotification(invalidFiles, validCount, invalidCount, totalFiles)
	}
}

// validateStrmFile 验证 strm 文件
func (s *ValidationService) validateStrmFile(strmPath string) ValidationResult {
	result := ValidationResult{
		IsValid:      false,
		StatusCode:   0,
		ErrorMessage: "",
		ResponseTime: 0,
	}

	startTime := time.Now()

	// 检查文件是否存在
	if _, err := os.Stat(strmPath); os.IsNotExist(err) {
		result.ErrorMessage = "strm 文件不存在"
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result
	}

	// 读取 strm 文件内容
	file, err := os.Open(strmPath)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("无法打开 strm 文件: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var streamURL string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && (strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://")) {
			streamURL = line
			break
		}
	}

	if streamURL == "" {
		result.ErrorMessage = "strm 文件中未找到有效的流媒体 URL"
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result
	}

	// 验证 URL 格式
	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("URL 格式无效: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result
	}

	// 检查 URL 可访问性
	req, err := http.NewRequest("HEAD", parsedURL.String(), nil)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("创建请求失败: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("网络请求失败: %v", err)
		result.ResponseTime = time.Since(startTime).Milliseconds()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ResponseTime = time.Since(startTime).Milliseconds()

	// 检查响应状态码
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.IsValid = true
	} else {
		result.ErrorMessage = fmt.Sprintf("HTTP 状态码: %d", resp.StatusCode)
	}

	return result
}

// getFilesToValidate 获取需要验证的文件列表
func (s *ValidationService) getFilesToValidate(taskID uint) ([]model.FileHistory, error) {
	var task model.ValidationTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}

	query := s.db.Model(&model.FileHistory{}).Where("file_suffix = ?", "strm")

	// 根据验证类型调整查询条件
	switch task.Type {
	case model.ValidationTaskTypeFull:
		// 全量验证，不添加额外条件
	case model.ValidationTaskTypeIncremental:
		// 增量验证，只验证未验证过的或距离上次验证时间较长的文件
		cutoffTime := time.Now().AddDate(0, 0, -7) // 7天前
		query = query.Where("last_checked_at IS NULL OR last_checked_at < ?", cutoffTime)
	case model.ValidationTaskTypeManual:
		// 手动验证，可以在配置中指定具体条件
		// 这里暂时使用全量验证的逻辑
	}

	var files []model.FileHistory
	if err := query.Find(&files).Error; err != nil {
		return nil, err
	}

	return files, nil
}

// updateTaskStatus 更新任务状态
func (s *ValidationService) updateTaskStatus(taskID uint, status model.ValidationTaskStatus, message string) {
	updates := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	if status == model.ValidationTaskStatusCompleted || status == model.ValidationTaskStatusFailed {
		now := time.Now()
		updates["completed_at"] = &now
	}

	s.db.Model(&model.ValidationTask{}).Where("id = ?", taskID).Updates(updates)
}

// sendInvalidFileNotification 发送失效文件通知
func (s *ValidationService) sendInvalidFileNotification(invalidFiles []map[string]interface{}, validCount, invalidCount, totalCount int) {
	// 统计主要失效原因
	reasonCounts := make(map[string]int)
	for _, file := range invalidFiles {
		if errorMsg, ok := file["errorMessage"].(string); ok {
			reasonCounts[errorMsg]++
		}
	}

	// 找出最主要的失效原因
	var mainReason string
	maxCount := 0
	for reason, count := range reasonCounts {
		if count > maxCount {
			maxCount = count
			mainReason = reason
		}
	}

	// 构建通知内容
	payload := map[string]interface{}{
		"TotalFiles":   totalCount,
		"ValidFiles":   validCount,
		"InvalidFiles": invalidCount,
		"MainReason":   mainReason,
		"Details":      invalidFiles,
	}

	// 添加到通知队列
	notificationReq := &model.NotificationQueueCreateRequest{
		Type:     model.NotificationTypeEmby,
		Event:    model.NotificationEventFileInvalid,
		Priority: model.NotificationPriorityNormal,
		Payload:  payload,
	}

	s.notificationService.AddToQueue(notificationReq)

	// 同时发送 Telegram 通知
	telegramReq := &model.NotificationQueueCreateRequest{
		Type:     model.NotificationTypeTelegram,
		Event:    model.NotificationEventFileInvalid,
		Priority: model.NotificationPriorityNormal,
		Payload:  payload,
	}

	s.notificationService.AddToQueue(telegramReq)
}

// GetTaskProgress 获取任务进度
func (s *ValidationService) GetTaskProgress(taskID uint) (*ValidationTaskProgress, error) {
	var task model.ValidationTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}

	progress := &ValidationTaskProgress{
		TaskID:         task.ID,
		Status:         task.Status,
		TotalFiles:     task.TotalFiles,
		ProcessedFiles: task.ProcessedFiles,
		ValidFiles:     task.ValidFiles,
		InvalidFiles:   task.InvalidFiles,
		Progress:       task.Progress,
		StartedAt:      task.StartedAt,
	}

	// 计算预计完成时间
	if task.Status == model.ValidationTaskStatusRunning && task.ProcessedFiles > 0 {
		elapsed := time.Since(*task.StartedAt)
		avgTimePerFile := elapsed / time.Duration(task.ProcessedFiles)
		remainingFiles := task.TotalFiles - task.ProcessedFiles
		estimatedRemaining := avgTimePerFile * time.Duration(remainingFiles)
		estimatedEnd := time.Now().Add(estimatedRemaining)
		progress.EstimatedEnd = &estimatedEnd
	}

	return progress, nil
}

// GetRunningTask 获取正在运行的任务
func (s *ValidationService) GetRunningTask() (*model.ValidationTask, error) {
	var task model.ValidationTask
	err := s.db.Where("status = ?", model.ValidationTaskStatusRunning).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

// CancelTask 取消任务
func (s *ValidationService) CancelTask(taskID uint) error {
	result := s.db.Model(&model.ValidationTask{}).Where("id = ? AND status = ?", taskID, model.ValidationTaskStatusRunning).
		Update("status", model.ValidationTaskStatusCancelled)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("任务不存在或不在运行状态")
	}

	s.logger.Info("验证任务已取消", zap.Uint("taskId", taskID))
	return nil
}

// CleanInvalidFiles 清理失效文件
func (s *ValidationService) CleanInvalidFiles() (int64, error) {
	// 获取失效的 strm 文件
	var invalidFiles []model.FileHistory
	err := s.db.Where("file_suffix = ? AND is_valid = ?", "strm", false).Find(&invalidFiles).Error
	if err != nil {
		return 0, err
	}

	cleanedCount := int64(0)

	for _, file := range invalidFiles {
		// 删除物理文件
		if _, err := os.Stat(file.TargetFilePath); err == nil {
			if err := os.Remove(file.TargetFilePath); err != nil {
				s.logger.Warn("删除失效文件失败",
					zap.String("filePath", file.TargetFilePath),
					zap.Error(err))
				continue
			}
		}

		// 删除目录（如果为空）
		dir := filepath.Dir(file.TargetFilePath)
		if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
			os.Remove(dir)
		}

		cleanedCount++
	}

	// 从数据库中删除失效文件记录
	result := s.db.Where("file_suffix = ? AND is_valid = ?", "strm", false).Delete(&model.FileHistory{})
	if result.Error != nil {
		return cleanedCount, result.Error
	}

	s.logger.Info("清理失效文件完成",
		zap.Int64("cleanedFiles", cleanedCount),
		zap.Int64("deletedRecords", result.RowsAffected))

	return cleanedCount, nil
}

// GetValidationStatistics 获取验证统计信息
func (s *ValidationService) GetValidationStatistics() (*model.ValidationStatistics, error) {
	stats := &model.ValidationStatistics{}

	// 任务统计
	s.db.Model(&model.ValidationTask{}).Count(&stats.TotalTasks)
	s.db.Model(&model.ValidationTask{}).Where("status = ?", model.ValidationTaskStatusRunning).Count(&stats.RunningTasks)
	s.db.Model(&model.ValidationTask{}).Where("status = ?", model.ValidationTaskStatusCompleted).Count(&stats.CompletedTasks)
	s.db.Model(&model.ValidationTask{}).Where("status = ?", model.ValidationTaskStatusFailed).Count(&stats.FailedTasks)

	// 文件统计
	s.db.Model(&model.FileHistory{}).Where("file_suffix = ?", "strm").Count(&stats.TotalFiles)
	s.db.Model(&model.FileHistory{}).Where("file_suffix = ? AND is_valid = ?", "strm", true).Count(&stats.ValidFiles)
	s.db.Model(&model.FileHistory{}).Where("file_suffix = ? AND is_valid = ?", "strm", false).Count(&stats.InvalidFiles)

	// 平均进度
	var avgProgress float64
	s.db.Model(&model.ValidationTask{}).Where("status = ?", model.ValidationTaskStatusCompleted).
		Select("AVG(progress)").Scan(&avgProgress)
	stats.AverageProgress = avgProgress

	return stats, nil
}

// ListTasks 获取验证任务列表
func (s *ValidationService) ListTasks(req *model.ValidationTaskQueryRequest) ([]*model.ValidationTask, int64, error) {
	query := s.db.Model(&model.ValidationTask{})

	// 添加过滤条件
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页和排序
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "desc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}

	offset := (req.Page - 1) * req.PageSize
	var tasks []*model.ValidationTask
	err := query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Offset(offset).Limit(req.PageSize).Find(&tasks).Error

	return tasks, total, err
}

// CreateTask 创建验证任务 (别名方法，调用StartValidationTask)
func (s *ValidationService) CreateTask(req *model.ValidationTaskCreateRequest) (*model.ValidationTask, error) {
	return s.StartValidationTask(req)
}

// GetTaskByID 根据ID获取验证任务
func (s *ValidationService) GetTaskByID(id uint) (*model.ValidationTask, error) {
	var task model.ValidationTask
	err := s.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// StartTask 启动验证任务
func (s *ValidationService) StartTask(taskID uint) error {
	// 检查任务是否存在
	var task model.ValidationTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 检查任务状态
	if task.Status == model.ValidationTaskStatusRunning {
		return fmt.Errorf("任务已在运行中")
	}

	if task.Status == model.ValidationTaskStatusCompleted {
		return fmt.Errorf("任务已完成")
	}

	// 检查是否有其他任务正在运行
	var runningTask model.ValidationTask
	err := s.db.Where("status = ? AND id != ?", model.ValidationTaskStatusRunning, taskID).First(&runningTask).Error
	if err == nil {
		return fmt.Errorf("已有其他验证任务正在运行，任务ID: %d", runningTask.ID)
	}

	// 重置任务状态并启动
	now := time.Now()
	updates := map[string]interface{}{
		"status":          model.ValidationTaskStatusPending,
		"started_at":      nil,
		"completed_at":    nil,
		"total_files":     0,
		"processed_files": 0,
		"valid_files":     0,
		"invalid_files":   0,
		"progress":        0,
		"message":         "",
		"updated_at":      now,
	}

	if err := s.db.Model(&task).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 异步执行验证任务
	go s.executeValidationTask(taskID)

	s.logger.Info("验证任务重新启动", zap.Uint("taskId", taskID))
	return nil
}

// DeleteTask 删除验证任务
func (s *ValidationService) DeleteTask(taskID uint) error {
	// 检查任务是否存在
	var task model.ValidationTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 检查任务是否在运行中
	if task.Status == model.ValidationTaskStatusRunning {
		return fmt.Errorf("不能删除正在运行的任务")
	}

	// 删除任务
	if err := s.db.Delete(&task).Error; err != nil {
		return fmt.Errorf("删除任务失败: %v", err)
	}

	s.logger.Info("验证任务已删除", zap.Uint("taskId", taskID))
	return nil
}

// GetStatistics 获取验证统计信息 (别名方法，调用GetValidationStatistics)
func (s *ValidationService) GetStatistics() (*model.ValidationStatistics, error) {
	return s.GetValidationStatistics()
}

// ValidateFile 验证单个文件
func (s *ValidationService) ValidateFile(filePath string) (bool, string, error) {
	// 验证文件
	result := s.validateStrmFile(filePath)

	if result.IsValid {
		return true, "文件验证通过", nil
	} else {
		return false, result.ErrorMessage, nil
	}
}

// CleanupInvalidFiles 清理无效文件 (别名方法，调用CleanInvalidFiles)
func (s *ValidationService) CleanupInvalidFiles() (int64, error) {
	return s.CleanInvalidFiles()
}
