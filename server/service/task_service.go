package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/MccRay-s/alist2strm/model/task"
	taskRequest "github.com/MccRay-s/alist2strm/model/task/request"
	taskResponse "github.com/MccRay-s/alist2strm/model/task/response"
	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/utils"
)

type TaskService struct{}

// 包级别的全局实例
var Task = &TaskService{}

// Create 创建任务
func (s *TaskService) Create(req *taskRequest.TaskCreateReq) error {
	// 创建任务
	newTask := &task.Task{
		Name:               req.Name,
		MediaType:          req.MediaType,
		SourcePath:         req.SourcePath,
		TargetPath:         req.TargetPath,
		FileSuffix:         req.FileSuffix,
		Overwrite:          req.Overwrite,
		Enabled:            req.Enabled,
		Cron:               req.Cron,
		Running:            false, // 新创建的任务默认不运行
		DownloadMetadata:   req.DownloadMetadata,
		DownloadSubtitle:   req.DownloadSubtitle,
		MetadataExtensions: req.MetadataExtensions,
		SubtitleExtensions: req.SubtitleExtensions,
	}

	// 设置默认值
	if newTask.MetadataExtensions == "" {
		newTask.MetadataExtensions = "nfo,jpg,png"
	}
	if newTask.SubtitleExtensions == "" {
		newTask.SubtitleExtensions = "srt,ass,ssa"
	}

	err := repository.Task.Create(newTask)
	if err != nil {
		return err
	}

	// 如果任务启用并设置了cron表达式，添加到调度器
	if newTask.Enabled && newTask.Cron != "" {
		scheduler := GetTaskScheduler()
		if err := scheduler.AddTask(newTask); err != nil {
			utils.Warn("添加任务到调度器失败", "task_id", newTask.ID, "error", err.Error())
		} else {
			utils.Info("任务已添加到调度器", "task_id", newTask.ID, "name", newTask.Name)
		}
	}

	return nil
}

// GetTaskInfo 获取任务信息
func (s *TaskService) GetTaskInfo(req *taskRequest.TaskInfoReq) (*taskResponse.TaskInfo, error) {
	task, err := repository.Task.GetByID(uint(req.ID))
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	resp := &taskResponse.TaskInfo{
		ID:                 task.ID,
		CreatedAt:          task.CreatedAt,
		UpdatedAt:          task.UpdatedAt,
		Name:               task.Name,
		MediaType:          task.MediaType,
		SourcePath:         task.SourcePath,
		TargetPath:         task.TargetPath,
		FileSuffix:         task.FileSuffix,
		Overwrite:          task.Overwrite,
		Enabled:            task.Enabled,
		Cron:               task.Cron,
		Running:            task.Running,
		LastRunAt:          task.LastRunAt,
		DownloadMetadata:   task.DownloadMetadata,
		DownloadSubtitle:   task.DownloadSubtitle,
		MetadataExtensions: task.MetadataExtensions,
		SubtitleExtensions: task.SubtitleExtensions,
	}

	return resp, nil
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(req *taskRequest.TaskUpdateReq) error {
	// 获取任务信息
	task, err := repository.Task.GetByID(req.ID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}

	// 检查任务是否正在运行
	if task.Running {
		return errors.New("任务正在运行，无法修改")
	}

	hasUpdate := false

	// 更新字段
	if req.Name != "" {
		task.Name = req.Name
		hasUpdate = true
	}
	if req.MediaType != "" {
		task.MediaType = req.MediaType
		hasUpdate = true
	}
	if req.SourcePath != "" {
		task.SourcePath = req.SourcePath
		hasUpdate = true
	}
	if req.TargetPath != "" {
		task.TargetPath = req.TargetPath
		hasUpdate = true
	}
	if req.FileSuffix != "" {
		task.FileSuffix = req.FileSuffix
		hasUpdate = true
	}
	if req.Overwrite != nil {
		task.Overwrite = *req.Overwrite
		hasUpdate = true
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
		hasUpdate = true
	}
	if req.Cron != "" {
		task.Cron = req.Cron
		hasUpdate = true
	}
	if req.DownloadMetadata != nil {
		task.DownloadMetadata = *req.DownloadMetadata
		hasUpdate = true
	}
	if req.DownloadSubtitle != nil {
		task.DownloadSubtitle = *req.DownloadSubtitle
		hasUpdate = true
	}
	if req.MetadataExtensions != "" {
		task.MetadataExtensions = req.MetadataExtensions
		hasUpdate = true
	}
	if req.SubtitleExtensions != "" {
		task.SubtitleExtensions = req.SubtitleExtensions
		hasUpdate = true
	}

	// 如果没有任何更新，返回错误
	if !hasUpdate {
		return errors.New("请提供要更新的信息")
	}

	err = repository.Task.Update(task)
	if err != nil {
		return err
	}

	// 更新任务调度
	scheduler := GetTaskScheduler()
	if err := scheduler.UpdateTask(task); err != nil {
		utils.Warn("更新任务调度失败", "task_id", task.ID, "error", err.Error())
	} else {
		utils.Info("任务调度已更新", "task_id", task.ID, "name", task.Name)
	}

	return nil
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(id uint) error {
	// 检查任务是否存在
	task, err := repository.Task.GetByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}

	// 检查任务是否正在运行
	if task.Running {
		return errors.New("任务正在运行，无法删除")
	}

	// 从队列中移除任务（如果存在）
	GetTaskQueue().RemoveTaskFromQueue(id)

	err = repository.Task.Delete(id)
	if err != nil {
		return err
	}

	// 从调度器中移除任务
	scheduler := GetTaskScheduler()
	scheduler.RemoveTask(id)
	utils.Info("任务已删除", "task_id", id)

	return nil
}

// GetTaskList 获取任务列表（分页）
func (s *TaskService) GetTaskList(req *taskRequest.TaskListReq) (*taskResponse.TaskListResp, error) {
	tasks, total, err := repository.Task.List(req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	taskInfos := make([]taskResponse.TaskInfo, len(tasks))
	for i, t := range tasks {
		taskInfos[i] = taskResponse.TaskInfo{
			ID:                 t.ID,
			CreatedAt:          t.CreatedAt,
			UpdatedAt:          t.UpdatedAt,
			Name:               t.Name,
			MediaType:          t.MediaType,
			SourcePath:         t.SourcePath,
			TargetPath:         t.TargetPath,
			FileSuffix:         t.FileSuffix,
			Overwrite:          t.Overwrite,
			Enabled:            t.Enabled,
			Cron:               t.Cron,
			Running:            t.Running,
			LastRunAt:          t.LastRunAt,
			DownloadMetadata:   t.DownloadMetadata,
			DownloadSubtitle:   t.DownloadSubtitle,
			MetadataExtensions: t.MetadataExtensions,
			SubtitleExtensions: t.SubtitleExtensions,
		}
	}

	resp := &taskResponse.TaskListResp{
		List:     taskInfos,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	return resp, nil
}

// GetAllTasks 获取所有任务（不分页）
func (s *TaskService) GetAllTasks(req *taskRequest.TaskAllReq) ([]taskResponse.TaskInfo, error) {
	tasks, err := repository.Task.ListAll(req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	taskInfos := make([]taskResponse.TaskInfo, len(tasks))
	for i, t := range tasks {
		taskInfos[i] = taskResponse.TaskInfo{
			ID:                 t.ID,
			CreatedAt:          t.CreatedAt,
			UpdatedAt:          t.UpdatedAt,
			Name:               t.Name,
			MediaType:          t.MediaType,
			SourcePath:         t.SourcePath,
			TargetPath:         t.TargetPath,
			FileSuffix:         t.FileSuffix,
			Overwrite:          t.Overwrite,
			Enabled:            t.Enabled,
			Cron:               t.Cron,
			Running:            t.Running,
			LastRunAt:          t.LastRunAt,
			DownloadMetadata:   t.DownloadMetadata,
			DownloadSubtitle:   t.DownloadSubtitle,
			MetadataExtensions: t.MetadataExtensions,
			SubtitleExtensions: t.SubtitleExtensions,
		}
	}

	return taskInfos, nil
}

// ToggleTaskEnabled 切换任务启用状态
func (s *TaskService) ToggleTaskEnabled(id uint) error {
	// 检查任务是否存在
	task, err := repository.Task.GetByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}

	// 如果任务正在运行，不允许禁用
	if task.Running && task.Enabled {
		return errors.New("任务正在运行，无法禁用")
	}

	// 切换启用状态
	task.Enabled = !task.Enabled

	err = repository.Task.Update(task)
	if err != nil {
		return err
	}

	// 更新任务调度
	scheduler := GetTaskScheduler()
	if err := scheduler.UpdateTask(task); err != nil {
		utils.Warn("更新任务调度失败", "task_id", task.ID, "error", err.Error())
	} else {
		if task.Enabled {
			utils.Info("任务已启用", "task_id", task.ID, "name", task.Name)
		} else {
			utils.Info("任务已禁用", "task_id", task.ID, "name", task.Name)
		}
	}

	return nil
}

// ResetTaskStatus 重置任务运行状态
func (s *TaskService) ResetTaskStatus(id uint) error {
	// 检查任务是否存在
	task, err := repository.Task.GetByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("任务不存在")
	}

	// 重置运行状态为false
	return repository.Task.UpdateRunningStatus(id, false)
}

// checkTaskExecutable 检查任务是否可执行
// 返回任务信息和错误（如果有）
func (s *TaskService) checkTaskExecutable(taskID uint) (*task.Task, error) {
	// 获取任务信息
	taskInfo, err := repository.Task.GetByID(taskID)
	if err != nil {
		return nil, err
	}
	if taskInfo == nil {
		return nil, errors.New("任务不存在")
	}

	// 检查任务是否已启用
	if !taskInfo.Enabled {
		return nil, errors.New("任务已禁用，无法执行")
	}

	// 检查任务是否正在运行
	if taskInfo.Running {
		return nil, errors.New("任务正在运行中")
	}

	return taskInfo, nil
}

// ExecuteTask 执行任务
func (s *TaskService) ExecuteTask(id uint, req *taskRequest.TaskExecuteReq) (*taskResponse.TaskExecuteResp, error) {
	// 使用辅助方法检查任务是否可执行
	task, err := s.checkTaskExecutable(id)
	if err != nil {
		return nil, err
	}

	// 创建执行结果响应
	startTime := time.Now()
	resp := &taskResponse.TaskExecuteResp{
		TaskID:    id,
		TaskName:  task.Name,
		IsSync:    req.Sync,
		Status:    "running",
		StartTime: startTime.Format("2006-01-02 15:04:05"),
	}

	if req.Sync {
		// 同步执行 STRM 生成
		execResult, err := s.ExecuteStrmGeneration(id)
		if err != nil {
			resp.Status = "failed"
			resp.Message = err.Error()
		} else {
			resp.Status = "completed"
			resp.Message = "任务执行完成"
		}

		// 复制详细执行结果
		if execResult != nil {
			resp.EndTime = execResult.EndTime
			resp.Duration = execResult.Duration
			resp.TotalCount = execResult.TotalCount
			resp.SuccessCount = execResult.SuccessCount
			resp.FailedCount = execResult.FailedCount
			resp.SkippedCount = execResult.SkippedCount
			resp.MetadataCount = execResult.MetadataCount
			resp.SubtitleCount = execResult.SubtitleCount
		} else {
			resp.EndTime = time.Now().Format("2006-01-02 15:04:05")
			resp.Duration = time.Since(startTime).String()
		}

		return resp, err
	} else {
		// 异步执行
		err := s.ExecuteStrmGenerationAsync(id)
		if err != nil {
			return nil, err
		}
		resp.Status = "running"
		resp.Message = "任务已提交执行，请稍后查看执行结果"
		return resp, nil
	}
}

// ExecuteStrmGeneration 执行 STRM 文件生成任务
func (s *TaskService) ExecuteStrmGeneration(taskID uint) (*taskResponse.TaskExecuteResp, error) {
	// 使用辅助方法检查任务是否可执行
	taskInfo, err := s.checkTaskExecutable(taskID)
	if err != nil {
		return nil, err
	}

	// 获取 STRM 生成服务
	strmService := GetStrmGeneratorService()
	if strmService == nil {
		return nil, errors.New("STRM 生成服务未初始化")
	}

	// 记录开始时间
	startTime := time.Now()

	// 更新任务运行状态
	if err := repository.Task.UpdateRunningStatus(taskID, true); err != nil {
		return nil, fmt.Errorf("更新任务运行状态失败: %w", err)
	}

	// 更新最后执行时间
	if err := repository.Task.UpdateLastRunAt(taskID, startTime); err != nil {
		utils.Error("更新任务最后执行时间失败", "task_id", taskID, "error", err.Error())
	}

	// 准备响应对象
	resp := &taskResponse.TaskExecuteResp{
		TaskID:    taskID,
		TaskName:  taskInfo.Name,
		IsSync:    true,
		Status:    "running",
		StartTime: startTime.Format("2006-01-02 15:04:05"),
	}

	// 启动 STRM 文件生成
	err = strmService.GenerateStrmFiles(taskID)

	// 更新任务运行状态
	if updateErr := repository.Task.UpdateRunningStatus(taskID, false); updateErr != nil {
		utils.Error("更新任务运行状态失败", "task_id", taskID, "error", updateErr.Error())
	}

	// 记录结束时间
	endTime := time.Now()
	resp.EndTime = endTime.Format("2006-01-02 15:04:05")
	resp.Duration = endTime.Sub(startTime).String()

	// 获取最新的任务日志记录，填充详细执行结果
	taskLogs, total, logErr := repository.TaskLog.GetLatestByTaskID(taskID, 1)
	if logErr == nil && total > 0 && len(taskLogs) > 0 {
		latestLog := taskLogs[0]
		resp.TotalCount = latestLog.TotalFile
		resp.SuccessCount = latestLog.GeneratedFile
		resp.SkippedCount = latestLog.SkipFile
		resp.MetadataCount = latestLog.MetadataCount
		resp.SubtitleCount = latestLog.SubtitleCount

		// 计算失败文件数
		resp.FailedCount = resp.TotalCount - resp.SuccessCount - resp.SkippedCount
	}

	if err != nil {
		resp.Status = "failed"
		return resp, err
	}

	resp.Status = "completed"
	return resp, nil
}

// ExecuteStrmGenerationAsync 异步执行 STRM 文件生成任务
func (s *TaskService) ExecuteStrmGenerationAsync(taskID uint) error {
	// 验证任务是否可执行
	taskInfo, err := s.checkTaskExecutable(taskID)
	if err != nil {
		return err
	}

	utils.Info("准备异步执行任务", "task_id", taskID, "name", taskInfo.Name)

	// 将任务添加到队列
	GetTaskQueue().AddTask(taskID)
	return nil
}
