package validationtask

import (
	"time"
)

// ValidationTask 验证任务模型
type ValidationTask struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"index"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	Type         string     `json:"type" gorm:"not null;index"`
	Status       string     `json:"status" gorm:"not null;default:pending;index"`
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	TotalFiles   int        `json:"totalFiles" gorm:"not null;default:0"`
	CheckedFiles int        `json:"checkedFiles" gorm:"not null;default:0"`
	InvalidFiles int        `json:"invalidFiles" gorm:"not null;default:0"`
	ErrorFiles   int        `json:"errorFiles" gorm:"not null;default:0"`
	Progress     int        `json:"progress" gorm:"not null;default:0"`
	Message      string     `json:"message" gorm:"type:text"`
	Config       string     `json:"config" gorm:"type:text"`
}

// TableName 表名
func (ValidationTask) TableName() string {
	return "validation_tasks"
}
