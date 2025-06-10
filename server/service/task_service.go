package service

import (
	"errors"
	"time"

	"github.com/MccRay-s/alist2strm/model/task"
	taskRequest "github.com/MccRay-s/alist2strm/model/task/request"
	taskResponse "github.com/MccRay-s/alist2strm/model/task/response"
	"github.com/MccRay-s/alist2strm/repository"
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
		newTask.MetadataExtensions = ".nfo,.jpg,.png"
	}
	if newTask.SubtitleExtensions == "" {
		newTask.SubtitleExtensions = ".srt,.ass,.ssa"
	}

	return repository.Task.Create(newTask)
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

	return repository.Task.Update(task)
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

	return repository.Task.Delete(id)
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
	return repository.Task.Update(task)
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

// ExecuteTask 执行任务
func (s *TaskService) ExecuteTask(id uint, req *taskRequest.TaskExecuteReq) (*taskResponse.TaskExecuteResp, error) {
	// 检查任务是否存在
	task, err := repository.Task.GetByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("任务不存在")
	}

	// 检查任务是否已启用
	if !task.Enabled {
		return nil, errors.New("任务已禁用，无法执行")
	}

	// 检查任务是否正在运行
	if task.Running {
		return nil, errors.New("任务正在运行中")
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
		// 同步执行
		result := s.executeTaskSync(task)
		return result, nil
	} else {
		// 异步执行
		go s.executeTaskAsync(task)
		resp.Status = "running"
		resp.Message = "任务已提交执行，请稍后查看执行结果"
		return resp, nil
	}
}

// executeTaskSync 同步执行任务
func (s *TaskService) executeTaskSync(task *task.Task) *taskResponse.TaskExecuteResp {
	// 设置任务为运行状态
	repository.Task.UpdateRunningStatus(task.ID, true)
	defer repository.Task.UpdateRunningStatus(task.ID, false)

	startTime := time.Now()
	result := &taskResponse.TaskExecuteResp{
		TaskID:    task.ID,
		TaskName:  task.Name,
		IsSync:    true,
		Status:    "completed",
		StartTime: startTime.Format("2006-01-02 15:04:05"),
		EndTime:   time.Now().Format("2006-01-02 15:04:05"),
		Duration:  time.Since(startTime).String(),
		Message:   "任务执行完成",
	}

	// TODO: 实现具体的任务执行逻辑
	// 这里可以根据任务类型执行不同的处理逻辑
	// 示例逻辑：
	result.TotalCount = 100                   // 模拟总文件数
	result.SuccessCount = 85                  // 模拟成功数
	result.FailedCount = 5                    // 模拟失败数
	result.SkippedCount = 8                   // 模拟跳过数
	result.OverwriteCount = 2                 // 模拟覆盖数
	result.SubtitleCount = 15                 // 模拟字幕文件数
	result.MetadataCount = 20                 // 模拟元数据文件数
	result.ErrorFiles = 2                     // 模拟错误文件数
	result.ProcessedBytes = 1024 * 1024 * 500 // 模拟处理500MB

	return result
}

// executeTaskAsync 异步执行任务
func (s *TaskService) executeTaskAsync(task *task.Task) {
	// 设置任务为运行状态
	repository.Task.UpdateRunningStatus(task.ID, true)
	defer repository.Task.UpdateRunningStatus(task.ID, false)

	// TODO: 实现异步执行逻辑
	// 可以创建任务日志记录执行过程
	// 执行完成后更新任务状态和结果
}
