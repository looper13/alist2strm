package systemlog

import (
	"time"
)

// SystemLog 系统日志模型
type SystemLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt" gorm:"index"`
	Level     string    `json:"level" gorm:"not null;index"`
	Module    string    `json:"module" gorm:"not null;index"`
	Action    string    `json:"action" gorm:"not null"`
	Message   string    `json:"message" gorm:"type:text"`
	UserID    *uint     `json:"userId" gorm:"index"`
	IP        string    `json:"ip" gorm:"size:45"`
	UserAgent string    `json:"userAgent" gorm:"type:text"`
	Extra     string    `json:"extra" gorm:"type:text"`
}

// TableName 表名
func (SystemLog) TableName() string {
	return "system_logs"
}
