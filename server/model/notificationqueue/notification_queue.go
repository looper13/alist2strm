package notificationqueue

import (
	"time"
)

// NotificationQueue 通知队列模型
type NotificationQueue struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"index"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	Type         string     `json:"type" gorm:"not null;index"`
	Event        string     `json:"event" gorm:"not null;index"`
	Payload      string     `json:"payload" gorm:"not null;type:text"`
	Status       string     `json:"status" gorm:"not null;default:pending;index"`
	RetryCount   int        `json:"retryCount" gorm:"not null;default:0"`
	MaxRetries   int        `json:"maxRetries" gorm:"not null;default:3"`
	NextRetryAt  *time.Time `json:"nextRetryAt" gorm:"index"`
	ProcessedAt  *time.Time `json:"processedAt"`
	ErrorMessage string     `json:"errorMessage" gorm:"type:text"`
	Priority     int        `json:"priority" gorm:"not null;default:5;index"`
}

// TableName 表名
func (NotificationQueue) TableName() string {
	return "notification_queue"
}
