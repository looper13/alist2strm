package tasklog

import (
	"time"
)

// TaskLog 任务日志模型
type TaskLog struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	TaskID        uint       `json:"taskId" gorm:"not null;index"`
	Status        string     `json:"status" gorm:"not null"`
	Message       string     `json:"message" gorm:"type:text"`
	StartTime     time.Time  `json:"startTime" gorm:"not null"`
	EndTime       *time.Time `json:"endTime"`
	TotalFile     int        `json:"totalFile" gorm:"not null;default:0"`
	GeneratedFile int        `json:"generatedFile" gorm:"not null;default:0"`
	SkipFile      int        `json:"skipFile" gorm:"not null;default:0"`
	MetadataCount int        `json:"metadataCount" gorm:"not null;default:0"`
	SubtitleCount int        `json:"subtitleCount" gorm:"not null;default:0"`
}

// TableName 表名
func (TaskLog) TableName() string {
	return "task_logs"
}
