package notification

import (
	"time"

	"gorm.io/gorm"
)

// Status 通知状态
type Status string

const (
	// StatusPending 待处理
	StatusPending Status = "pending"
	// StatusProcessing 处理中
	StatusProcessing Status = "processing"
	// StatusSent 已发送
	StatusSent Status = "sent"
	// StatusFailed 发送失败
	StatusFailed Status = "failed"
)

// Queue 通知队列模型
type Queue struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	ChannelType   string         `json:"channelType" gorm:"not null;index"`
	TemplateType  string         `json:"templateType" gorm:"not null"`
	Status        Status         `json:"status" gorm:"not null;index"`
	Payload       string         `json:"payload" gorm:"type:text;not null"` // JSON格式的通知数据
	ErrorMessage  string         `json:"errorMessage" gorm:"type:text"`     // 错误信息
	RetryCount    int            `json:"retryCount" gorm:"default:0"`       // 重试次数
	NextRetryTime *time.Time     `json:"nextRetryTime"`                     // 下次重试时间
}

// TableName 表名
func (Queue) TableName() string {
	return "notification_queue"
}

// NotificationData 通知数据接口
type NotificationData interface {
	GetTaskName() string
}

// TaskNotificationData 任务通知数据
type TaskNotificationData struct {
	TaskID         uint      `json:"taskId"`
	TaskName       string    `json:"taskName"`
	Status         string    `json:"status"`
	Duration       int64     `json:"duration"`
	TotalFiles     int       `json:"totalFiles"`
	GeneratedFiles int       `json:"generatedFiles"`
	SkippedFiles   int       `json:"skippedFiles"`
	MetadataFiles  int       `json:"metadataFiles"`
	SubtitleFiles  int       `json:"subtitleFiles"`
	ErrorMessage   string    `json:"errorMessage,omitempty"`
	EventTime      time.Time `json:"eventTime"`         // 事件发生时间
	SourcePath     string    `json:"sourcePath"`        // 任务源路径
	TargetPath     string    `json:"targetPath"`        // 任务目标路径
}

// GetTaskName 获取任务名称
func (d *TaskNotificationData) GetTaskName() string {
	return d.TaskName
}
