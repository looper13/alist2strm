package model

import (
	"time"
)

// ValidationTaskType 验证任务类型枚举
type ValidationTaskType string

const (
	ValidationTaskTypeFull        ValidationTaskType = "full"
	ValidationTaskTypeIncremental ValidationTaskType = "incremental"
	ValidationTaskTypeManual      ValidationTaskType = "manual"
)

// ValidationTaskStatus 验证任务状态枚举
type ValidationTaskStatus string

const (
	ValidationTaskStatusPending   ValidationTaskStatus = "pending"
	ValidationTaskStatusRunning   ValidationTaskStatus = "running"
	ValidationTaskStatusCompleted ValidationTaskStatus = "completed"
	ValidationTaskStatusFailed    ValidationTaskStatus = "failed"
	ValidationTaskStatusCancelled ValidationTaskStatus = "cancelled"
)

// ValidationTask 失效检测任务模型
type ValidationTask struct {
	ID             uint                 `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
	Type           ValidationTaskType   `gorm:"not null;index" json:"type"`
	Status         ValidationTaskStatus `gorm:"not null;default:pending;index" json:"status"`
	StartedAt      *time.Time           `json:"startedAt"`
	CompletedAt    *time.Time           `json:"completedAt"`
	TotalFiles     int                  `gorm:"not null;default:0" json:"totalFiles"`
	ProcessedFiles int                  `gorm:"not null;default:0" json:"processedFiles"`
	ValidFiles     int                  `gorm:"not null;default:0" json:"validFiles"`
	InvalidFiles   int                  `gorm:"not null;default:0" json:"invalidFiles"`
	Progress       int                  `gorm:"not null;default:0" json:"progress"` // 进度百分比
	Message        string               `gorm:"type:text" json:"message"`
	Config         string               `gorm:"type:text" json:"config"` // JSON格式
}

// ValidationTaskCreateRequest 创建验证任务请求
type ValidationTaskCreateRequest struct {
	Type   ValidationTaskType `json:"type" binding:"required"`
	Config string             `json:"config"`
}

// ValidationTaskQueryRequest 查询验证任务请求
type ValidationTaskQueryRequest struct {
	Page      int                   `form:"page" binding:"min=1"`
	PageSize  int                   `form:"pageSize" binding:"min=1,max=100"`
	Type      *ValidationTaskType   `form:"type"`
	Status    *ValidationTaskStatus `form:"status"`
	StartDate string                `form:"startDate"`
	EndDate   string                `form:"endDate"`
	SortBy    string                `form:"sortBy"`
	SortOrder string                `form:"sortOrder"`
}

// ValidationStatistics 验证统计
type ValidationStatistics struct {
	TotalTasks      int64   `json:"totalTasks"`
	RunningTasks    int64   `json:"runningTasks"`
	CompletedTasks  int64   `json:"completedTasks"`
	FailedTasks     int64   `json:"failedTasks"`
	TotalFiles      int64   `json:"totalFiles"`
	ValidFiles      int64   `json:"validFiles"`
	InvalidFiles    int64   `json:"invalidFiles"`
	AverageProgress float64 `json:"averageProgress"`
}
