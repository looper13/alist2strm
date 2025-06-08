package model

import (
	"time"
)

// NotificationType 通知类型枚举
type NotificationType string

const (
	NotificationTypeEmby     NotificationType = "emby"
	NotificationTypeTelegram NotificationType = "telegram"
)

// NotificationEvent 通知事件枚举
type NotificationEvent string

const (
	NotificationEventTaskCompleted NotificationEvent = "task_completed"
	NotificationEventTaskFailed    NotificationEvent = "task_failed"
	NotificationEventFileInvalid   NotificationEvent = "file_invalid"
)

// NotificationQueueStatus 通知队列状态枚举
type NotificationQueueStatus string

const (
	NotificationQueueStatusPending    NotificationQueueStatus = "pending"
	NotificationQueueStatusProcessing NotificationQueueStatus = "processing"
	NotificationQueueStatusCompleted  NotificationQueueStatus = "completed"
	NotificationQueueStatusFailed     NotificationQueueStatus = "failed"
)

// NotificationQueue 通知队列模型
type NotificationQueue struct {
	ID           uint                    `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time               `json:"createdAt"`
	UpdatedAt    time.Time               `json:"updatedAt"`
	Type         NotificationType        `gorm:"not null;index" json:"type"`
	Event        NotificationEvent       `gorm:"not null;index" json:"event"`
	Payload      string                  `gorm:"type:text;not null" json:"payload"` // JSON格式
	Status       NotificationQueueStatus `gorm:"not null;default:pending;index" json:"status"`
	RetryCount   int                     `gorm:"not null;default:0" json:"retryCount"`
	MaxRetries   int                     `gorm:"not null;default:3" json:"maxRetries"`
	NextRetryAt  *time.Time              `gorm:"index" json:"nextRetryAt"`
	ProcessedAt  *time.Time              `json:"processedAt"`
	ErrorMessage string                  `gorm:"type:text" json:"errorMessage"`
	Priority     int                     `gorm:"not null;default:5" json:"priority"` // 1-10，数字越小优先级越高
}

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

// SystemLogLevel 系统日志级别枚举
type SystemLogLevel string

const (
	SystemLogLevelDebug SystemLogLevel = "debug"
	SystemLogLevelInfo  SystemLogLevel = "info"
	SystemLogLevelWarn  SystemLogLevel = "warn"
	SystemLogLevelError SystemLogLevel = "error"
)

// SystemLogModule 系统日志模块枚举
type SystemLogModule string

const (
	SystemLogModuleNotification SystemLogModule = "notification"
	SystemLogModuleValidation   SystemLogModule = "validation"
	SystemLogModuleFileService  SystemLogModule = "file_service"
	SystemLogModuleTask         SystemLogModule = "task"
	SystemLogModuleSystem       SystemLogModule = "system"
	SystemLogModuleAuth         SystemLogModule = "auth"
	SystemLogModuleAPI          SystemLogModule = "api"
)

// SystemLog 系统日志模型
type SystemLog struct {
	ID        uint            `gorm:"primarykey" json:"id"`
	CreatedAt time.Time       `json:"createdAt"`
	Level     SystemLogLevel  `gorm:"not null;index" json:"level"`
	Module    SystemLogModule `gorm:"not null;index" json:"module"`
	Operation string          `gorm:"size:100;not null;index" json:"operation"`
	Message   string          `gorm:"type:text;not null" json:"message"`
	Data      string          `gorm:"type:text" json:"data"` // JSON格式
	UserID    *uint           `gorm:"index" json:"userId"`
	IP        string          `gorm:"size:45" json:"ip"`
	UserAgent string          `gorm:"size:500" json:"userAgent"`
}

// NotificationQueueCreateRequest 创建通知队列请求
type NotificationQueueCreateRequest struct {
	Type       NotificationType  `json:"type" binding:"required"`
	Event      NotificationEvent `json:"event" binding:"required"`
	Payload    string            `json:"payload" binding:"required"`
	Priority   int               `json:"priority" binding:"min=1,max=10"`
	MaxRetries int               `json:"maxRetries" binding:"min=1,max=10"`
}

// ValidationTaskCreateRequest 创建验证任务请求
type ValidationTaskCreateRequest struct {
	Type   ValidationTaskType `json:"type" binding:"required"`
	Config string             `json:"config"`
}

// SystemLogCreateRequest 创建系统日志请求
type SystemLogCreateRequest struct {
	Level     SystemLogLevel  `json:"level" binding:"required"`
	Module    SystemLogModule `json:"module" binding:"required"`
	Operation string          `json:"operation" binding:"required"`
	Message   string          `json:"message" binding:"required"`
	Data      string          `json:"data"`
	UserID    *uint           `json:"userId"`
	IP        string          `json:"ip"`
	UserAgent string          `json:"userAgent"`
}

// NotificationQueueQueryRequest 查询通知队列请求
type NotificationQueueQueryRequest struct {
	Page      int                      `form:"page" binding:"min=1"`
	PageSize  int                      `form:"pageSize" binding:"min=1,max=100"`
	Type      *NotificationType        `form:"type"`
	Event     *NotificationEvent       `form:"event"`
	Status    *NotificationQueueStatus `form:"status"`
	StartDate string                   `form:"startDate"`
	EndDate   string                   `form:"endDate"`
	SortBy    string                   `form:"sortBy"`
	SortOrder string                   `form:"sortOrder"`
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

// SystemLogQueryRequest 查询系统日志请求
type SystemLogQueryRequest struct {
	Page      int              `form:"page" binding:"min=1"`
	PageSize  int              `form:"pageSize" binding:"min=1,max=100"`
	Level     *SystemLogLevel  `form:"level"`
	Module    *SystemLogModule `form:"module"`
	Operation string           `form:"operation"`
	UserID    *uint            `form:"userId"`
	StartDate string           `form:"startDate"`
	EndDate   string           `form:"endDate"`
	SortBy    string           `form:"sortBy"`
	SortOrder string           `form:"sortOrder"`
}

// NotificationStatistics 通知统计
type NotificationStatistics struct {
	TotalNotifications     int64   `json:"totalNotifications"`
	PendingNotifications   int64   `json:"pendingNotifications"`
	CompletedNotifications int64   `json:"completedNotifications"`
	FailedNotifications    int64   `json:"failedNotifications"`
	EmbyNotifications      int64   `json:"embyNotifications"`
	TelegramNotifications  int64   `json:"telegramNotifications"`
	AverageRetryCount      float64 `json:"averageRetryCount"`
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
