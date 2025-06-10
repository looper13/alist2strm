package service

import (
	"fmt"
	"time"

	"github.com/MccRay-s/alist2strm/model/tasklog"
	taskLogRequest "github.com/MccRay-s/alist2strm/model/tasklog/request"
	taskLogResponse "github.com/MccRay-s/alist2strm/model/tasklog/response"
	"github.com/MccRay-s/alist2strm/repository"
)

type TaskLogService struct{}

// 包级别的全局实例
var TaskLogServiceInstance = &TaskLogService{}

// CreateTaskLog 创建任务日志
func (s *TaskLogService) CreateTaskLog(req *taskLogRequest.TaskLogCreateReq) (*taskLogResponse.TaskLogCreateResp, error) {
	// 检查任务是否存在
	task, err := repository.Task.GetByID(req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}
	if task == nil {
		return nil, fmt.Errorf("任务不存在")
	}

	// 创建任务日志
	taskLog := &tasklog.TaskLog{
		TaskID:        req.TaskID,
		Status:        req.Status,
		Message:       req.Message,
		StartTime:     time.Now(),
		TotalFile:     req.TotalFile,
		GeneratedFile: req.GeneratedFile,
		SkipFile:      req.SkipFile,
		OverwriteFile: req.OverwriteFile,
		MetadataCount: req.MetadataCount,
		SubtitleCount: req.SubtitleCount,
		FailedCount:   req.FailedCount,
	}

	if err := repository.TaskLog.Create(taskLog); err != nil {
		return nil, fmt.Errorf("创建任务日志失败: %v", err)
	}

	return &taskLogResponse.TaskLogCreateResp{
		ID:        taskLog.ID,
		TaskID:    taskLog.TaskID,
		Status:    taskLog.Status,
		StartTime: taskLog.StartTime,
		Message:   taskLog.Message,
	}, nil
}

// UpdateTaskLog 更新任务日志
func (s *TaskLogService) UpdateTaskLog(req *taskLogRequest.TaskLogUpdateReq) error {
	// 检查任务日志是否存在
	taskLog, err := repository.TaskLog.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("查询任务日志失败: %v", err)
	}
	if taskLog == nil {
		return fmt.Errorf("任务日志不存在")
	}

	// 准备更新数据
	updates := make(map[string]interface{})

	if req.Status != "" {
		updates["status"] = req.Status
		// 如果状态变为完成或失败，设置结束时间
		if req.Status == "completed" || req.Status == "failed" || req.Status == "cancelled" {
			now := time.Now()
			updates["end_time"] = now
			// 如果没有传入duration，计算持续时间
			if req.Duration == nil {
				duration := now.Sub(taskLog.StartTime).Seconds()
				updates["duration"] = int64(duration)
			}
		}
	}

	if req.Message != "" {
		updates["message"] = req.Message
	}

	if req.Duration != nil {
		updates["duration"] = *req.Duration
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

	if req.OverwriteFile != nil {
		updates["overwrite_file"] = *req.OverwriteFile
	}

	if req.MetadataCount != nil {
		updates["metadata_count"] = *req.MetadataCount
	}

	if req.SubtitleCount != nil {
		updates["subtitle_count"] = *req.SubtitleCount
	}

	if req.FailedCount != nil {
		updates["failed_count"] = *req.FailedCount
	}

	if len(updates) == 0 {
		return fmt.Errorf("没有需要更新的字段")
	}

	if err := repository.TaskLog.UpdatePartial(req.ID, updates); err != nil {
		return fmt.Errorf("更新任务日志失败: %v", err)
	}

	return nil
}

// GetTaskLogInfo 获取任务日志信息
func (s *TaskLogService) GetTaskLogInfo(req *taskLogRequest.TaskLogInfoReq) (*taskLogResponse.TaskLogInfoResp, error) {
	taskLog, err := repository.TaskLog.GetByID(req.Uint())
	if err != nil {
		return nil, fmt.Errorf("查询任务日志失败: %v", err)
	}
	if taskLog == nil {
		return nil, fmt.Errorf("任务日志不存在")
	}

	return &taskLogResponse.TaskLogInfoResp{
		TaskLog: *taskLog,
	}, nil
}

// GetTaskLogList 获取任务日志列表
func (s *TaskLogService) GetTaskLogList(req *taskLogRequest.TaskLogListReq) (*taskLogResponse.TaskLogListResp, error) {
	// 检查任务是否存在
	task, err := repository.Task.GetByID(req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}
	if task == nil {
		return nil, fmt.Errorf("任务不存在")
	}

	taskLogs, total, err := repository.TaskLog.ListByTaskID(req)
	if err != nil {
		return nil, fmt.Errorf("查询任务日志列表失败: %v", err)
	}

	return &taskLogResponse.TaskLogListResp{
		List:     taskLogs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
