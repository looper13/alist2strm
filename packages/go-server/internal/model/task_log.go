package model

import (
	"time"
)

// TaskLog 任务日志模型
type TaskLog struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	TaskID        uint       `gorm:"not null;index" json:"taskId"`
	Status        string     `gorm:"not null" json:"status"`
	Message       string     `gorm:"type:text" json:"message"`
	StartTime     time.Time  `gorm:"not null" json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
	TotalFile     int        `gorm:"not null;default:0" json:"totalFile"`
	GeneratedFile int        `gorm:"not null;default:0" json:"generatedFile"`
	SkipFile      int        `gorm:"not null;default:0" json:"skipFile"`
	MetadataCount int        `gorm:"not null;default:0" json:"metadataCount"`
	SubtitleCount int        `gorm:"not null;default:0" json:"subtitleCount"`
}

// TaskLogStatus 任务状态常量
const (
	TaskLogStatusPending   = "pending"   // 等待中
	TaskLogStatusRunning   = "running"   // 运行中
	TaskLogStatusCompleted = "completed" // 已完成
	TaskLogStatusFailed    = "failed"    // 失败
	TaskLogStatusCancelled = "cancelled" // 已取消
)

// TaskLogCreateRequest 创建任务日志请求
type TaskLogCreateRequest struct {
	TaskID    uint   `json:"taskId" binding:"required"`
	Status    string `json:"status" binding:"required,oneof=pending running completed failed cancelled"`
	Message   string `json:"message"`
	StartTime string `json:"startTime" binding:"required"`
}

// TaskLogUpdateRequest 更新任务日志请求
type TaskLogUpdateRequest struct {
	Status        string `json:"status,omitempty" binding:"omitempty,oneof=pending running completed failed cancelled"`
	Message       string `json:"message,omitempty"`
	EndTime       string `json:"endTime,omitempty"`
	TotalFile     *int   `json:"totalFile,omitempty"`
	GeneratedFile *int   `json:"generatedFile,omitempty"`
	SkipFile      *int   `json:"skipFile,omitempty"`
	MetadataCount *int   `json:"metadataCount,omitempty"`
	SubtitleCount *int   `json:"subtitleCount,omitempty"`
}

// TaskLogQueryRequest 查询任务日志请求
type TaskLogQueryRequest struct {
	TaskID    *uint  `form:"taskId"`
	Status    string `form:"status"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"pageSize" binding:"min=1,max=100"`
	SortBy    string `form:"sortBy" binding:"oneof=id createdAt startTime endTime"`
	SortOrder string `form:"sortOrder" binding:"oneof=asc desc"`
}

// TaskLogStatistics 任务日志统计
type TaskLogStatistics struct {
	TotalLogs      int64 `json:"totalLogs"`
	RunningLogs    int64 `json:"runningLogs"`
	CompletedLogs  int64 `json:"completedLogs"`
	FailedLogs     int64 `json:"failedLogs"`
	TotalFiles     int64 `json:"totalFiles"`
	GeneratedFiles int64 `json:"generatedFiles"`
	SkippedFiles   int64 `json:"skippedFiles"`
	MetadataFiles  int64 `json:"metadataFiles"`
	SubtitleFiles  int64 `json:"subtitleFiles"`
}
