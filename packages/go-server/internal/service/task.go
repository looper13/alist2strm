package service

import (
	"alist2strm/internal/model"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskService struct {
	db     *gorm.DB
	logger *zap.Logger
}

type CreateTaskRequest struct {
	Name               string `json:"name" binding:"required"`
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

func NewTaskService(db *gorm.DB, logger *zap.Logger) *TaskService {
	return &TaskService{
		db:     db,
		logger: logger,
	}
}

// Create 创建新任务
func (s *TaskService) Create(req *CreateTaskRequest) (*model.Task, error) {
	task := &model.Task{
		Name:               req.Name,
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
func (s *TaskService) Update(id uint, req *UpdateTaskRequest) (*model.Task, error) {
	task, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		task.Name = *req.Name
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
func (s *TaskService) Delete(id uint) error {
	task, err := s.GetByID(id)
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
func (s *TaskService) GetByID(id uint) (*model.Task, error) {
	var task model.Task
	if err := s.db.First(&task, id).Error; err != nil {
		s.logger.Error("获取任务失败", zap.Error(err))
		return nil, err
	}
	return &task, nil
}

// List 获取任务列表
func (s *TaskService) List() ([]model.Task, error) {
	var tasks []model.Task
	if err := s.db.Find(&tasks).Error; err != nil {
		s.logger.Error("获取任务列表失败", zap.Error(err))
		return nil, err
	}
	return tasks, nil
}

// SetStatus 设置任务状态（启用/禁用）
func (s *TaskService) SetStatus(id uint, enabled bool) error {
	task, err := s.GetByID(id)
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
	task, err := s.GetByID(id)
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
