package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FileHistoryService struct {
	db     *gorm.DB
	logger *zap.Logger
}

var (
	fileHistoryService *FileHistoryService
	fileHistoryOnce    sync.Once
)

// GetFileHistoryService 获取 FileHistoryService 单例
func GetFileHistoryService() *FileHistoryService {
	fileHistoryOnce.Do(func() {
		fileHistoryService = &FileHistoryService{
			db:     model.DB,
			logger: utils.Logger,
		}
	})
	return fileHistoryService
}

// Create 创建文件历史记录
func (s *FileHistoryService) Create(req *model.FileHistoryCreateRequest) (*model.FileHistory, error) {
	fileHistory := &model.FileHistory{
		TaskID:         req.TaskID,
		TaskLogID:      req.TaskLogID,
		FileName:       req.FileName,
		SourcePath:     req.SourcePath,
		SourceURL:      req.SourceURL,
		TargetFilePath: req.TargetFilePath,
		FileSize:       req.FileSize,
		FileType:       req.FileType,
		FileSuffix:     req.FileSuffix,
		FileCategory:   req.FileCategory,
		IsMainFile:     req.IsMainFile,
		MainFileID:     req.MainFileID,
		MediaInfo:      req.MediaInfo,
		Hash:           req.Hash,
		Metadata:       req.Metadata,
		Tags:           req.Tags,
	}

	// 设置默认值
	if fileHistory.FileCategory == "" {
		fileHistory.FileCategory = model.FileCategoryMain
	}

	now := time.Now()
	fileHistory.LastProcessedAt = &now

	if err := s.db.Create(fileHistory).Error; err != nil {
		utils.Error("创建文件历史记录失败", zap.Error(err))
		return nil, err
	}

	utils.Info("创建文件历史记录成功", zap.String("fileName", fileHistory.FileName))
	return fileHistory, nil
}

// GetByID 根据ID获取文件历史记录
func (s *FileHistoryService) GetByID(id uint) (*model.FileHistory, error) {
	var fileHistory model.FileHistory
	if err := s.db.First(&fileHistory, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("文件历史记录不存在")
		}
		utils.Error("获取文件历史记录失败", zap.Error(err))
		return nil, err
	}
	return &fileHistory, nil
}

// List 获取文件历史记录列表
func (s *FileHistoryService) List(req *model.FileHistoryQueryRequest) ([]*model.FileHistoryResponse, int64, error) {
	query := s.db.Model(&model.FileHistory{})

	// 构建查询条件
	if req.TaskID != nil {
		query = query.Where("task_id = ?", *req.TaskID)
	}
	if req.TaskLogID != nil {
		query = query.Where("task_log_id = ?", *req.TaskLogID)
	}
	if req.FileName != "" {
		query = query.Where("file_name LIKE ?", "%"+req.FileName+"%")
	}
	if req.FileCategory != nil {
		query = query.Where("file_category = ?", *req.FileCategory)
	}
	if req.ProcessingStatus != nil {
		query = query.Where("processing_status = ?", *req.ProcessingStatus)
	}
	if req.IsValid != nil {
		query = query.Where("is_valid = ?", *req.IsValid)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.Error("获取文件历史记录总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 排序和分页
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	var fileHistories []model.FileHistory
	if err := query.Find(&fileHistories).Error; err != nil {
		utils.Error("获取文件历史记录列表失败", zap.Error(err))
		return nil, 0, err
	}

	// 转换为响应格式
	var responses []*model.FileHistoryResponse
	for _, fh := range fileHistories {
		response := &model.FileHistoryResponse{
			FileHistory: fh,
		}
		responses = append(responses, response)
	}

	return responses, total, nil
}

// GetStatistics 获取文件历史统计信息
func (s *FileHistoryService) GetStatistics() (*model.FileHistoryStatistics, error) {
	stats := &model.FileHistoryStatistics{}

	// 总文件数
	if err := s.db.Model(&model.FileHistory{}).Count(&stats.TotalFiles).Error; err != nil {
		return nil, err
	}

	// 按文件类别统计
	s.db.Model(&model.FileHistory{}).Where("file_category = ?", model.FileCategoryMain).Count(&stats.MainFiles)
	s.db.Model(&model.FileHistory{}).Where("file_category = ?", model.FileCategoryMetadata).Count(&stats.MetadataFiles)
	s.db.Model(&model.FileHistory{}).Where("file_category = ?", model.FileCategorySubtitle).Count(&stats.SubtitleFiles)

	// 按处理状态统计
	s.db.Model(&model.FileHistory{}).Where("processing_status = ?", model.ProcessingStatusSuccess).Count(&stats.SuccessFiles)
	s.db.Model(&model.FileHistory{}).Where("processing_status = ?", model.ProcessingStatusFailed).Count(&stats.FailedFiles)
	s.db.Model(&model.FileHistory{}).Where("processing_status = ?", model.ProcessingStatusPending).Count(&stats.PendingFiles)
	s.db.Model(&model.FileHistory{}).Where("processing_status = ?", model.ProcessingStatusSkipped).Count(&stats.SkippedFiles)

	// 按有效性统计
	s.db.Model(&model.FileHistory{}).Where("is_valid = ?", true).Count(&stats.ValidFiles)
	s.db.Model(&model.FileHistory{}).Where("is_valid = ?", false).Count(&stats.InvalidFiles)

	// 文件大小统计
	var totalSize sql.NullInt64
	s.db.Model(&model.FileHistory{}).Select("SUM(file_size)").Scan(&totalSize)
	if totalSize.Valid {
		stats.TotalSize = totalSize.Int64
		if stats.TotalFiles > 0 {
			stats.AverageFileSize = stats.TotalSize / stats.TotalFiles
		}
	}

	return stats, nil
}

// MarkAsProcessed 标记文件为已处理
func (s *FileHistoryService) MarkAsProcessed(id uint, status model.ProcessingStatus, message string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"processing_status":  status,
		"processing_message": message,
		"last_processed_at":  &now,
	}

	if err := s.db.Model(&model.FileHistory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.Error("标记文件处理状态失败", zap.Error(err))
		return err
	}

	return nil
}

// MarkAsValidated 标记文件为已验证
func (s *FileHistoryService) MarkAsValidated(id uint, isValid bool, message string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"is_valid":           isValid,
		"validation_message": message,
		"last_checked_at":    &now,
	}

	if err := s.db.Model(&model.FileHistory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.Error("标记文件验证状态失败", zap.Error(err))
		return err
	}

	return nil
}

// BatchDelete 批量删除文件历史记录
func (s *FileHistoryService) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("没有提供要删除的ID")
	}

	if err := s.db.Delete(&model.FileHistory{}, ids).Error; err != nil {
		utils.Error("批量删除文件历史记录失败", zap.Error(err))
		return err
	}

	utils.Info("批量删除文件历史记录成功", zap.Int("count", len(ids)))
	return nil
}

// Update 更新文件历史记录
func (s *FileHistoryService) Update(id uint, req *model.FileHistoryUpdateRequest) error {
	// 构建更新字段
	updates := make(map[string]interface{})

	if req.FileName != nil {
		updates["file_name"] = *req.FileName
	}
	if req.SourcePath != nil {
		updates["source_path"] = *req.SourcePath
	}
	if req.SourceURL != nil {
		updates["source_url"] = *req.SourceURL
	}
	if req.TargetFilePath != nil {
		updates["target_file_path"] = *req.TargetFilePath
	}
	if req.FileSize != nil {
		updates["file_size"] = *req.FileSize
	}
	if req.FileType != nil {
		updates["file_type"] = *req.FileType
	}
	if req.FileSuffix != nil {
		updates["file_suffix"] = *req.FileSuffix
	}
	if req.FileCategory != nil {
		updates["file_category"] = *req.FileCategory
	}
	if req.ProcessingStatus != nil {
		updates["processing_status"] = *req.ProcessingStatus
	}
	if req.ProcessingMessage != nil {
		updates["processing_message"] = *req.ProcessingMessage
	}
	if req.IsValid != nil {
		updates["is_valid"] = *req.IsValid
	}
	if req.ValidationMessage != nil {
		updates["validation_message"] = *req.ValidationMessage
	}
	if req.Hash != nil {
		updates["hash"] = *req.Hash
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}

	if len(updates) == 0 {
		return fmt.Errorf("没有提供要更新的字段")
	}

	if err := s.db.Model(&model.FileHistory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.Error("更新文件历史记录失败", zap.Error(err))
		return err
	}

	utils.Info("更新文件历史记录成功", zap.Uint("id", id))
	return nil
}

// Delete 删除文件历史记录
func (s *FileHistoryService) Delete(id uint) error {
	result := s.db.Delete(&model.FileHistory{}, id)
	if result.Error != nil {
		utils.Error("删除文件历史记录失败", zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("文件历史记录不存在")
	}

	utils.Info("删除文件历史记录成功", zap.Uint("id", id))
	return nil
}

// ClearAll 清空所有文件历史记录
func (s *FileHistoryService) ClearAll() error {
	result := s.db.Exec("DELETE FROM file_histories")
	if result.Error != nil {
		utils.Error("清空文件历史记录失败", zap.Error(result.Error))
		return result.Error
	}

	utils.Info("清空文件历史记录成功", zap.Int64("deletedCount", result.RowsAffected))
	return nil
}

// GetValidationSummary 获取验证摘要
func (s *FileHistoryService) GetValidationSummary() (*model.ValidationSummary, error) {
	summary := &model.ValidationSummary{}

	// 总文件数（仅限 strm 文件）
	s.db.Model(&model.FileHistory{}).Where("file_suffix = ?", "strm").Count(&summary.TotalFiles)

	// 有效和无效文件数
	s.db.Model(&model.FileHistory{}).Where("file_suffix = ? AND is_valid = ?", "strm", true).Count(&summary.ValidFiles)
	s.db.Model(&model.FileHistory{}).Where("file_suffix = ? AND is_valid = ?", "strm", false).Count(&summary.InvalidFiles)

	// 待检查文件数（未验证过的文件）
	s.db.Model(&model.FileHistory{}).Where("file_suffix = ? AND last_checked_at IS NULL", "strm").Count(&summary.PendingFiles)

	// 最后检查时间
	var lastCheck time.Time
	if err := s.db.Model(&model.FileHistory{}).Where("file_suffix = ? AND last_checked_at IS NOT NULL", "strm").
		Select("MAX(last_checked_at)").Scan(&lastCheck).Error; err == nil && !lastCheck.IsZero() {
		summary.LastCheckTime = &lastCheck
	}

	// 下次检查时间（计算需要重新验证的文件的最早时间）
	var nextCheck time.Time
	if err := s.db.Model(&model.FileHistory{}).Where("file_suffix = ? AND next_check_at IS NOT NULL", "strm").
		Select("MIN(next_check_at)").Scan(&nextCheck).Error; err == nil && !nextCheck.IsZero() {
		summary.NextCheckTime = &nextCheck
	}

	// 失效原因统计
	type FailureReasonRow struct {
		ValidationMessage string `json:"validation_message"`
		Count             int64  `json:"count"`
	}

	var reasonRows []FailureReasonRow
	s.db.Model(&model.FileHistory{}).
		Where("file_suffix = ? AND is_valid = ? AND validation_message != ''", "strm", false).
		Select("validation_message, COUNT(*) as count").
		Group("validation_message").
		Order("count DESC").
		Limit(10).
		Scan(&reasonRows)

	summary.FailureReasons = make([]model.FailureReason, len(reasonRows))
	for i, row := range reasonRows {
		summary.FailureReasons[i] = model.FailureReason{
			Reason: row.ValidationMessage,
			Count:  row.Count,
		}
	}

	return summary, nil
}

// GetNotificationSummary 获取通知摘要
func (s *FileHistoryService) GetNotificationSummary() (*model.NotificationSummary, error) {
	summary := &model.NotificationSummary{}

	// 总文件数
	s.db.Model(&model.FileHistory{}).Count(&summary.TotalFiles)

	// 已通知文件数
	s.db.Model(&model.FileHistory{}).Where("emby_notified = ? OR telegram_notified = ?", true, true).Count(&summary.NotifiedFiles)

	// Emby 通知数
	s.db.Model(&model.FileHistory{}).Where("emby_notified = ?", true).Count(&summary.EmbyNotified)

	// Telegram 通知数
	s.db.Model(&model.FileHistory{}).Where("telegram_notified = ?", true).Count(&summary.TelegramNotified)

	// 失败通知数
	s.db.Model(&model.FileHistory{}).Where("notification_status = ?", "failed").Count(&summary.FailedNotifications)

	// 最后通知时间
	var lastNotified time.Time
	if err := s.db.Model(&model.FileHistory{}).Where("notification_sent_at IS NOT NULL").
		Select("MAX(notification_sent_at)").Scan(&lastNotified).Error; err == nil && !lastNotified.IsZero() {
		summary.LastNotificationTime = &lastNotified
	}

	return summary, nil
}
