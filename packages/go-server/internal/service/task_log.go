package service

import (
	"alist2strm/internal/model"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskLogService struct {
	db     *gorm.DB
	logger *zap.Logger
}

var (
	taskLogService *TaskLogService
	taskLogOnce    sync.Once
)

// GetTaskLogService 获取TaskLogService单例
func GetTaskLogService(db *gorm.DB, logger *zap.Logger) *TaskLogService {
	taskLogOnce.Do(func() {
		taskLogService = &TaskLogService{
			db:     db,
			logger: logger,
		}
	})
	return taskLogService
}

// Create 创建任务日志
func (s *TaskLogService) Create(req *model.TaskLogCreateRequest) (*model.TaskLog, error) {
	// 验证任务是否存在
	var task model.Task
	if err := s.db.First(&task, req.TaskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务不存在")
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	// 解析开始时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %w", err)
	}

	taskLog := &model.TaskLog{
		TaskID:    req.TaskID,
		Status:    req.Status,
		Message:   req.Message,
		StartTime: startTime,
	}

	if err := s.db.Create(taskLog).Error; err != nil {
		s.logger.Error("创建任务日志失败", zap.Error(err))
		return nil, fmt.Errorf("创建任务日志失败: %w", err)
	}

	s.logger.Info("创建任务日志成功",
		zap.Uint("id", taskLog.ID),
		zap.Uint("taskId", taskLog.TaskID),
		zap.String("status", taskLog.Status))

	return taskLog, nil
}

// Update 更新任务日志
func (s *TaskLogService) Update(id uint, req *model.TaskLogUpdateRequest) (*model.TaskLog, error) {
	var taskLog model.TaskLog
	if err := s.db.First(&taskLog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务日志不存在")
		}
		return nil, fmt.Errorf("查询任务日志失败: %w", err)
	}

	updates := make(map[string]interface{})

	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Message != "" {
		updates["message"] = req.Message
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, fmt.Errorf("结束时间格式错误: %w", err)
		}
		updates["end_time"] = endTime
	}
	if req.TotalFile != nil {
		updates["total_file"] = *req.TotalFile
	}
	if req.GeneratedFile != nil {
		updates["generated_file"] = *req.GeneratedFile
	}
	if req.SkipFile != nil {
		updates["skip_file"] = *req.SkipFile
	}
	if req.MetadataCount != nil {
		updates["metadata_count"] = *req.MetadataCount
	}
	if req.SubtitleCount != nil {
		updates["subtitle_count"] = *req.SubtitleCount
	}

	if len(updates) > 0 {
		if err := s.db.Model(&taskLog).Updates(updates).Error; err != nil {
			s.logger.Error("更新任务日志失败", zap.Error(err))
			return nil, fmt.Errorf("更新任务日志失败: %w", err)
		}
	}

	// 重新查询更新后的数据
	if err := s.db.First(&taskLog, id).Error; err != nil {
		return nil, fmt.Errorf("查询更新后的任务日志失败: %w", err)
	}

	s.logger.Info("更新任务日志成功", zap.Uint("id", id))
	return &taskLog, nil
}

// GetByID 根据ID获取任务日志
func (s *TaskLogService) GetByID(id uint) (*model.TaskLog, error) {
	var taskLog model.TaskLog
	if err := s.db.First(&taskLog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("任务日志不存在")
		}
		return nil, fmt.Errorf("查询任务日志失败: %w", err)
	}
	return &taskLog, nil
}

// List 获取任务日志列表
func (s *TaskLogService) List(req *model.TaskLogQueryRequest) ([]*model.TaskLog, int64, error) {
	query := s.db.Model(&model.TaskLog{})

	// 过滤条件
	if req.TaskID != nil {
		query = query.Where("task_id = ?", *req.TaskID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, 0, fmt.Errorf("开始日期格式错误: %w", err)
		}
		query = query.Where("start_time >= ?", startDate)
	}
	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, 0, fmt.Errorf("结束日期格式错误: %w", err)
		}
		// 结束日期包含当天
		endDate = endDate.Add(24*time.Hour - time.Second)
		query = query.Where("start_time <= ?", endDate)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计任务日志数量失败: %w", err)
	}

	// 排序
	sortBy := "created_at"
	sortOrder := "desc"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 分页
	page := 1
	pageSize := 10
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	var taskLogs []*model.TaskLog
	if err := query.Find(&taskLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询任务日志列表失败: %w", err)
	}

	return taskLogs, total, nil
}

// Delete 删除任务日志
func (s *TaskLogService) Delete(id uint) error {
	result := s.db.Delete(&model.TaskLog{}, id)
	if result.Error != nil {
		s.logger.Error("删除任务日志失败", zap.Error(result.Error))
		return fmt.Errorf("删除任务日志失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("任务日志不存在")
	}

	s.logger.Info("删除任务日志成功", zap.Uint("id", id))
	return nil
}

// DeleteByTaskID 删除指定任务的所有日志
func (s *TaskLogService) DeleteByTaskID(taskID uint) error {
	result := s.db.Where("task_id = ?", taskID).Delete(&model.TaskLog{})
	if result.Error != nil {
		s.logger.Error("删除任务日志失败", zap.Error(result.Error))
		return fmt.Errorf("删除任务日志失败: %w", result.Error)
	}

	s.logger.Info("删除任务日志成功",
		zap.Uint("taskId", taskID),
		zap.Int64("deletedCount", result.RowsAffected))
	return nil
}

// GetStatistics 获取任务日志统计信息
func (s *TaskLogService) GetStatistics(taskID *uint) (*model.TaskLogStatistics, error) {
	query := s.db.Model(&model.TaskLog{})
	if taskID != nil {
		query = query.Where("task_id = ?", *taskID)
	}

	stats := &model.TaskLogStatistics{}

	// 总日志数
	if err := query.Count(&stats.TotalLogs).Error; err != nil {
		return nil, fmt.Errorf("统计总日志数失败: %w", err)
	}

	// 各状态日志数
	if err := query.Where("status = ?", model.TaskLogStatusRunning).Count(&stats.RunningLogs).Error; err != nil {
		return nil, fmt.Errorf("统计运行中日志数失败: %w", err)
	}
	if err := query.Where("status = ?", model.TaskLogStatusCompleted).Count(&stats.CompletedLogs).Error; err != nil {
		return nil, fmt.Errorf("统计已完成日志数失败: %w", err)
	}
	if err := query.Where("status = ?", model.TaskLogStatusFailed).Count(&stats.FailedLogs).Error; err != nil {
		return nil, fmt.Errorf("统计失败日志数失败: %w", err)
	}

	// 文件统计
	var fileStats struct {
		TotalFiles     int64 `gorm:"column:total_files"`
		GeneratedFiles int64 `gorm:"column:generated_files"`
		SkippedFiles   int64 `gorm:"column:skipped_files"`
		MetadataFiles  int64 `gorm:"column:metadata_files"`
		SubtitleFiles  int64 `gorm:"column:subtitle_files"`
	}

	queryFileStats := query.Select(
		"COALESCE(SUM(total_file), 0) as total_files",
		"COALESCE(SUM(generated_file), 0) as generated_files",
		"COALESCE(SUM(skip_file), 0) as skipped_files",
		"COALESCE(SUM(metadata_count), 0) as metadata_files",
		"COALESCE(SUM(subtitle_count), 0) as subtitle_files",
	)

	if err := queryFileStats.Scan(&fileStats).Error; err != nil {
		return nil, fmt.Errorf("统计文件数量失败: %w", err)
	}

	stats.TotalFiles = fileStats.TotalFiles
	stats.GeneratedFiles = fileStats.GeneratedFiles
	stats.SkippedFiles = fileStats.SkippedFiles
	stats.MetadataFiles = fileStats.MetadataFiles
	stats.SubtitleFiles = fileStats.SubtitleFiles

	return stats, nil
}

// GetLatestByTaskID 获取指定任务的最新日志
func (s *TaskLogService) GetLatestByTaskID(taskID uint) (*model.TaskLog, error) {
	var taskLog model.TaskLog
	if err := s.db.Where("task_id = ?", taskID).Order("created_at desc").First(&taskLog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有日志记录不算错误
		}
		return nil, fmt.Errorf("查询最新任务日志失败: %w", err)
	}
	return &taskLog, nil
}

// MarkAsCompleted 标记任务日志为已完成
func (s *TaskLogService) MarkAsCompleted(id uint, totalFile, generatedFile, skipFile, metadataCount, subtitleCount int) error {
	updates := map[string]interface{}{
		"status":         model.TaskLogStatusCompleted,
		"end_time":       time.Now(),
		"total_file":     totalFile,
		"generated_file": generatedFile,
		"skip_file":      skipFile,
		"metadata_count": metadataCount,
		"subtitle_count": subtitleCount,
	}

	if err := s.db.Model(&model.TaskLog{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("标记任务日志为已完成失败", zap.Error(err))
		return fmt.Errorf("标记任务日志为已完成失败: %w", err)
	}

	s.logger.Info("标记任务日志为已完成成功",
		zap.Uint("id", id),
		zap.Int("totalFile", totalFile),
		zap.Int("generatedFile", generatedFile),
		zap.Int("skipFile", skipFile),
		zap.Int("metadataCount", metadataCount),
		zap.Int("subtitleCount", subtitleCount))

	return nil
}

// MarkAsFailed 标记任务日志为失败
func (s *TaskLogService) MarkAsFailed(id uint, message string) error {
	updates := map[string]interface{}{
		"status":   model.TaskLogStatusFailed,
		"end_time": time.Now(),
		"message":  message,
	}

	if err := s.db.Model(&model.TaskLog{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("标记任务日志为失败失败", zap.Error(err))
		return fmt.Errorf("标记任务日志为失败失败: %w", err)
	}

	s.logger.Info("标记任务日志为失败成功",
		zap.Uint("id", id),
		zap.String("message", message))

	return nil
}

// StartTask 开始任务（创建运行中的日志记录）
func (s *TaskLogService) StartTask(taskID uint, message string) (*model.TaskLog, error) {
	req := &model.TaskLogCreateRequest{
		TaskID:    taskID,
		Status:    model.TaskLogStatusRunning,
		Message:   message,
		StartTime: time.Now().Format(time.RFC3339),
	}

	return s.Create(req)
}

// CleanOldLogs 清理旧的日志记录
func (s *TaskLogService) CleanOldLogs(days int) (int64, error) {
	if days <= 0 {
		return 0, errors.New("天数必须大于0")
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	result := s.db.Where("created_at < ?", cutoffTime).Delete(&model.TaskLog{})
	if result.Error != nil {
		s.logger.Error("清理旧日志失败", zap.Error(result.Error))
		return 0, fmt.Errorf("清理旧日志失败: %w", result.Error)
	}

	s.logger.Info("清理旧日志成功",
		zap.Int("days", days),
		zap.Int64("deletedCount", result.RowsAffected))

	return result.RowsAffected, nil
}
