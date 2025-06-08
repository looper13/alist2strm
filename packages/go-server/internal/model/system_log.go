package model

import (
	"time"
)

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
