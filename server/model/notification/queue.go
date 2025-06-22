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
	TaskID             uint   `json:"taskId"`
	TaskName           string `json:"taskName"`
	Status             string `json:"status"`
	Duration           int64  `json:"duration"`
	TotalFile          int    `json:"totalFile"`          // 总文件数，与 TaskLog 保持一致
	GeneratedFile      int    `json:"generatedFile"`      // 生成的文件数，与 TaskLog 保持一致
	SkipFile           int    `json:"skipFile"`           // 跳过的文件数，与 TaskLog 保持一致
	OverwriteFile      int    `json:"overwriteFile"`      // 覆盖的文件数，与 TaskLog 保持一致
	MetadataCount      int    `json:"metadataCount"`      // 元数据文件数，与 TaskLog 保持一致
	SubtitleCount      int    `json:"subtitleCount"`      // 字幕文件数，与 TaskLog 保持一致
	MetadataDownloaded int    `json:"metadataDownloaded"` // 已下载的元数据文件数，与 TaskLog 保持一致
	SubtitleDownloaded int    `json:"subtitleDownloaded"` // 已下载的字幕文件数，与 TaskLog 保持一致
	FailedCount        int    `json:"failedCount"`        // 处理失败的文件数，与 TaskLog 保持一致

	// 以下字段是为了通知显示更详细信息而保留的额外字段
	MetadataSkipped int       `json:"metadataSkipped"` // 跳过的元数据文件数
	SubtitleSkipped int       `json:"subtitleSkipped"` // 跳过的字幕文件数
	OtherSkipped    int       `json:"otherSkipped"`    // 跳过的其他文件数
	ErrorMessage    string    `json:"errorMessage,omitempty"`
	EventTime       time.Time `json:"eventTime"`  // 事件发生时间
	SourcePath      string    `json:"sourcePath"` // 任务源路径
	TargetPath      string    `json:"targetPath"` // 任务目标路径
}

// GetTaskName 获取任务名称
func (d *TaskNotificationData) GetTaskName() string {
	return d.TaskName
}
