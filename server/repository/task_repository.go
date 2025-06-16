package repository

import (
	"errors"
	"time"

	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/model/task"
	taskRequest "github.com/MccRay-s/alist2strm/model/task/request"
	"gorm.io/gorm"
)

type TaskRepository struct{}

// 包级别的全局实例
var Task = &TaskRepository{}

// Create 创建任务
func (r *TaskRepository) Create(task *task.Task) error {
	return database.DB.Create(task).Error
}

// GetByID 根据ID获取任务
func (r *TaskRepository) GetByID(id uint) (*task.Task, error) {
	var t task.Task
	err := database.DB.Where("id = ?", id).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// Update 更新任务
func (r *TaskRepository) Update(task *task.Task) error {
	return database.DB.Save(task).Error
}

// Delete 删除任务
func (r *TaskRepository) Delete(id uint) error {
	return database.DB.Delete(&task.Task{}, id).Error
}

// List 获取任务列表（分页）
func (r *TaskRepository) List(req *taskRequest.TaskListReq) ([]task.Task, int64, error) {
	var tasks []task.Task
	var total int64

	query := database.DB.Model(&task.Task{})

	// 添加名称筛选
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 添加媒体类型筛选
	if req.MediaType != "" {
		query = query.Where("media_type = ?", req.MediaType)
	}

	// 添加启用状态筛选
	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}

	// 添加运行状态筛选
	if req.Running != nil {
		query = query.Where("running = ?", *req.Running)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Scopes(req.Paginate()).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

// ListAll 获取所有任务（不分页）
func (r *TaskRepository) ListAll(req *taskRequest.TaskAllReq) ([]task.Task, error) {
	var tasks []task.Task

	query := database.DB.Model(&task.Task{})

	// 添加名称筛选
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 添加媒体类型筛选
	if req.MediaType != "" {
		query = query.Where("media_type = ?", req.MediaType)
	}

	// 添加启用状态筛选
	if req.Enabled != nil {
		query = query.Where("enabled = ?", *req.Enabled)
	}

	// 添加运行状态筛选
	if req.Running != nil {
		query = query.Where("running = ?", *req.Running)
	}

	// 查询所有数据，按创建时间排序
	err := query.Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// UpdateRunningStatus 更新任务运行状态
func (r *TaskRepository) UpdateRunningStatus(id uint, running bool) error {
	updates := map[string]interface{}{
		"running": running,
	}

	// 如果设置为运行状态，更新最后运行时间
	if running {
		now := time.Now()
		updates["last_run_at"] = &now
	}

	return database.DB.Model(&task.Task{}).Where("id = ?", id).Updates(updates).Error
}

// ResetRunningStatus 重置所有任务运行状态
func (r *TaskRepository) ResetRunningStatus() error {
	return database.DB.Model(&task.Task{}).Where("running = ?", true).Update("running", false).Error
}

// GetAllEnabled 获取所有启用且有Cron表达式的任务
func (r *TaskRepository) GetAllEnabled() ([]task.Task, error) {
	var tasks []task.Task
	// 查询启用且cron表达式不为空的任务
	if err := database.DB.Where("enabled = ? AND cron != ?", true, "").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateLastRunAt 更新任务最后执行时间
func (r *TaskRepository) UpdateLastRunAt(id uint, lastRunAt time.Time) error {
	return database.DB.Model(&task.Task{}).Where("id = ?", id).Update("last_run_at", lastRunAt).Error
}
