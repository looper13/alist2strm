package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"errors"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService struct {
	db     *gorm.DB
	logger *zap.Logger
}

var (
	taskService *TaskService
	taskOnce    sync.Once
)

type CreateTaskRequest struct {
	Name               string `json:"name" binding:"required"`
	MediaType          string `json:"mediaType" binding:"required,oneof=movie tv"`
	SourcePath         string `json:"sourcePath" binding:"required"`
	TargetPath         string `json:"targetPath" binding:"required"`
	FileSuffix         string `json:"fileSuffix" binding:"required"`
	Overwrite          bool   `json:"overwrite"`
	Enabled            bool   `json:"enabled"`
	Cron               string `json:"cron"`
	DownloadMetadata   bool   `json:"downloadMetadata"`
	DownloadSubtitle   bool   `json:"downloadSubtitle"`
	MetadataExtensions string `json:"metadataExtensions"`
	SubtitleExtensions string `json:"subtitleExtensions"`
}

type UpdateTaskRequest struct {
	Name               *string `json:"name"`
	MediaType          *string `json:"mediaType" binding:"omitempty,oneof=movie tv"`
	SourcePath         *string `json:"sourcePath"`
	TargetPath         *string `json:"targetPath"`
	FileSuffix         *string `json:"fileSuffix"`
	Overwrite          *bool   `json:"overwrite"`
	Enabled            *bool   `json:"enabled"`
	Cron               *string `json:"cron"`
	DownloadMetadata   *bool   `json:"downloadMetadata"`
	DownloadSubtitle   *bool   `json:"downloadSubtitle"`
	MetadataExtensions *string `json:"metadataExtensions"`
	SubtitleExtensions *string `json:"subtitleExtensions"`
}

// GetTaskService 获取 TaskService 单例
func GetTaskService() *TaskService {
	taskOnce.Do(func() {
		taskService = &TaskService{
			db:     model.DB,
			logger: utils.Logger,
		}
	})
	return taskService
}

// Create 创建新任务
func (s *TaskService) CreateTask(req *CreateTaskRequest) (*model.Task, error) {
	task := &model.Task{
		Name:               req.Name,
		MediaType:          req.MediaType,
		SourcePath:         req.SourcePath,
		TargetPath:         req.TargetPath,
		FileSuffix:         req.FileSuffix,
		Overwrite:          req.Overwrite,
		Enabled:            req.Enabled,
		Cron:               req.Cron,
		DownloadMetadata:   req.DownloadMetadata,
		DownloadSubtitle:   req.DownloadSubtitle,
		MetadataExtensions: req.MetadataExtensions,
		SubtitleExtensions: req.SubtitleExtensions,
	}

	if err := s.db.Create(task).Error; err != nil {
		s.logger.Error("创建任务失败", zap.Error(err))
		return nil, err
	}

	return task, nil
}

// Update 更新任务
func (s *TaskService) UpdateTask(id uint, req *UpdateTaskRequest) (*model.Task, error) {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.MediaType != nil {
		task.MediaType = *req.MediaType
	}
	if req.SourcePath != nil {
		task.SourcePath = *req.SourcePath
	}
	if req.TargetPath != nil {
		task.TargetPath = *req.TargetPath
	}
	if req.FileSuffix != nil {
		task.FileSuffix = *req.FileSuffix
	}
	if req.Overwrite != nil {
		task.Overwrite = *req.Overwrite
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}
	if req.Cron != nil {
		task.Cron = *req.Cron
	}
	if req.DownloadMetadata != nil {
		task.DownloadMetadata = *req.DownloadMetadata
	}
	if req.DownloadSubtitle != nil {
		task.DownloadSubtitle = *req.DownloadSubtitle
	}
	if req.MetadataExtensions != nil {
		task.MetadataExtensions = *req.MetadataExtensions
	}
	if req.SubtitleExtensions != nil {
		task.SubtitleExtensions = *req.SubtitleExtensions
	}

	if err := s.db.Save(task).Error; err != nil {
		s.logger.Error("更新任务失败", zap.Error(err))
		return nil, err
	}

	return task, nil
}

// Delete 删除任务
func (s *TaskService) DeleteTask(id uint) error {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return err
	}

	if task.Running {
		return errors.New("无法删除正在运行的任务")
	}

	if err := s.db.Delete(task).Error; err != nil {
		s.logger.Error("删除任务失败", zap.Error(err))
		return err
	}

	return nil
}

// GetByID 通过ID获取任务
func (s *TaskService) GetTaskByID(id uint) (*model.Task, error) {
	var task model.Task
	if err := s.db.First(&task, id).Error; err != nil {
		s.logger.Error("获取任务失败", zap.Error(err))
		return nil, err
	}
	return &task, nil
}

// List 获取任务列表
func (s *TaskService) ListTasks(name string) ([]model.Task, error) {
	var tasks []model.Task
	query := s.db

	// 如果提供了name参数，添加模糊查询条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Find(&tasks).Error; err != nil {
		s.logger.Error("获取任务列表失败",
			zap.Error(err),
			zap.String("name", name))
		return nil, err
	}

	return tasks, nil
}

// SetStatus 设置任务状态（启用/禁用）
func (s *TaskService) SetTaskStatus(id uint, enabled bool) error {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return err
	}

	task.Enabled = enabled
	if err := s.db.Save(task).Error; err != nil {
		s.logger.Error("设置任务状态失败", zap.Error(err))
		return err
	}

	return nil
}

// ResetRunningStatus 重置任务的运行状态
func (s *TaskService) ResetRunningStatus(id uint) error {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return err
	}

	task.Running = false
	if err := s.db.Save(task).Error; err != nil {
		s.logger.Error("重置任务运行状态失败", zap.Error(err))
		return err
	}

	return nil
}
