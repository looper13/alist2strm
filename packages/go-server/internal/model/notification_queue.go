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

// NotificationQueueCreateRequest 创建通知队列请求
type NotificationQueueCreateRequest struct {
	Type       NotificationType  `json:"type" binding:"required"`
	Event      NotificationEvent `json:"event" binding:"required"`
	Payload    string            `json:"payload" binding:"required"`
	Priority   int               `json:"priority" binding:"min=1,max=10"`
	MaxRetries int               `json:"maxRetries" binding:"min=1,max=10"`
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
