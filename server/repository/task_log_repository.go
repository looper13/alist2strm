package repository

import (
	"errors"
	"time"

	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/model/tasklog"
	taskLogRequest "github.com/MccRay-s/alist2strm/model/tasklog/request"
	"gorm.io/gorm"
)

type TaskLogRepository struct{}

// 包级别的全局实例
var TaskLog = &TaskLogRepository{}

// Create 创建任务日志
func (r *TaskLogRepository) Create(taskLog *tasklog.TaskLog) error {
	return database.DB.Create(taskLog).Error
}

// GetByID 根据ID获取任务日志
func (r *TaskLogRepository) GetByID(id uint) (*tasklog.TaskLog, error) {
	var tl tasklog.TaskLog
	err := database.DB.Where("id = ?", id).First(&tl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tl, nil
}

// Update 更新任务日志
func (r *TaskLogRepository) Update(taskLog *tasklog.TaskLog) error {
	return database.DB.Save(taskLog).Error
}

// UpdatePartial 部分更新任务日志
func (r *TaskLogRepository) UpdatePartial(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&tasklog.TaskLog{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除任务日志
func (r *TaskLogRepository) Delete(id uint) error {
	return database.DB.Delete(&tasklog.TaskLog{}, id).Error
}

// GetTotalExecutionsCount 获取总执行次数
func (r *TaskLogRepository) GetTotalExecutionsCount() (int64, error) {
	var count int64
	err := database.DB.Model(&tasklog.TaskLog{}).Count(&count).Error
	return count, err
}

// 已使用新的 GetSuccessCount 和 GetFailedCount 方法替代

// GetSuccessCount 获取成功执行总次数，可指定时间范围
func (r *TaskLogRepository) GetSuccessCount(timeRange string) (int64, error) {
	var count int64
	query := database.DB.Model(&tasklog.TaskLog{}).Where("status = ?", tasklog.TaskLogStatusCompleted)

	// 根据时间范围添加筛选条件
	query = r.addTimeRangeFilter(query, timeRange)

	err := query.Count(&count).Error
	return count, err
}

// GetFailedCount 获取失败执行总次数，可指定时间范围
func (r *TaskLogRepository) GetFailedCount(timeRange string) (int64, error) {
	var count int64
	query := database.DB.Model(&tasklog.TaskLog{}).Where("status = ?", tasklog.TaskLogStatusFailed)

	// 根据时间范围添加筛选条件
	query = r.addTimeRangeFilter(query, timeRange)

	err := query.Count(&count).Error
	return count, err
}

// ListByTaskID 根据任务ID获取任务日志列表（分页）
func (r *TaskLogRepository) ListByTaskID(req *taskLogRequest.TaskLogListReq) ([]tasklog.TaskLog, int64, error) {
	var taskLogs []tasklog.TaskLog
	var total int64

	query := database.DB.Model(&tasklog.TaskLog{}).Where("task_id = ?", req.TaskID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按创建时间倒序
	err := query.Scopes(req.Paginate()).Order("created_at DESC").Find(&taskLogs).Error
	return taskLogs, total, err
}

// DeleteByTaskID 删除指定任务的所有日志
func (r *TaskLogRepository) DeleteByTaskID(taskID uint) error {
	return database.DB.Where("task_id = ?", taskID).Delete(&tasklog.TaskLog{}).Error
}

// GetLatestByTaskID 获取任务的最新日志
func (r *TaskLogRepository) GetLatestByTaskID(taskID uint, limit int) ([]tasklog.TaskLog, int64, error) {
	var logs []tasklog.TaskLog
	var total int64

	// 获取总数
	if err := database.DB.Model(&tasklog.TaskLog{}).Where("task_id = ?", taskID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取最新的日志记录
	if err := database.DB.Where("task_id = ?", taskID).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// UpdateEndTime 更新任务日志结束时间和持续时间
func (r *TaskLogRepository) UpdateEndTime(id uint, endTime time.Time, duration int64) error {
	updates := map[string]interface{}{
		"end_time": endTime,
		"duration": duration,
	}
	return database.DB.Model(&tasklog.TaskLog{}).Where("id = ?", id).Updates(updates).Error
}

// GetRunningLogByTaskID 获取任务正在运行的日志
func (r *TaskLogRepository) GetRunningLogByTaskID(taskID uint) (*tasklog.TaskLog, error) {
	var tl tasklog.TaskLog
	err := database.DB.Where("task_id = ? AND status = ?", taskID, "running").First(&tl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tl, nil
}

// addTimeRangeFilter 根据时间范围添加筛选条件
func (r *TaskLogRepository) addTimeRangeFilter(query *gorm.DB, timeRange string) *gorm.DB {
	now := time.Now()

	switch timeRange {
	case "day":
		// 当日数据
		today := now.Format("2006-01-02")
		query = query.Where("DATE(created_at) = ?", today)
	case "month":
		// 当月数据（从当月1号到当前时间）
		firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		query = query.Where("created_at >= ? AND created_at <= ?", firstDayOfMonth, now)
	case "year":
		// 当年数据（从1月1号到当前时间）
		firstDayOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		query = query.Where("created_at >= ? AND created_at <= ?", firstDayOfYear, now)
	}

	return query
}
