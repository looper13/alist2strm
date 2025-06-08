package model

import (
	"time"
)

// ProcessingStatus 文件处理状态枚举
type ProcessingStatus string

const (
	ProcessingStatusSuccess ProcessingStatus = "success"
	ProcessingStatusFailed  ProcessingStatus = "failed"
	ProcessingStatusPending ProcessingStatus = "pending"
	ProcessingStatusSkipped ProcessingStatus = "skipped"
)

// NotificationStatus 通知状态枚举
type NotificationStatus int

const (
	NotificationStatusNotSent NotificationStatus = 0 // 未发送
	NotificationStatusSuccess NotificationStatus = 1 // 成功通知
	NotificationStatusFailed  NotificationStatus = 2 // 失败通知
	NotificationStatusInvalid NotificationStatus = 3 // 失效通知
)

// FileCategory 文件类别枚举
type FileCategory string

const (
	FileCategoryMain     FileCategory = "main"     // 主文件（视频文件）
	FileCategoryMetadata FileCategory = "metadata" // 刮削数据文件
	FileCategorySubtitle FileCategory = "subtitle" // 字幕文件
)

// FileHistory 文件历史记录模型
type FileHistory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TaskID    uint      `gorm:"not null;index" json:"taskId"`
	TaskLogID uint      `gorm:"not null;index" json:"taskLogId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 文件基本信息
	FileName       string `gorm:"not null" json:"fileName"`
	SourcePath     string `gorm:"not null" json:"sourcePath"`
	SourceURL      string `gorm:"size:512" json:"sourceUrl"`
	TargetFilePath string `gorm:"not null" json:"targetFilePath"`
	FileSize       int64  `gorm:"not null" json:"fileSize"`
	FileType       string `gorm:"not null" json:"fileType"`
	FileSuffix     string `gorm:"not null" json:"fileSuffix"`

	// 文件分类和关联
	FileCategory FileCategory `gorm:"not null;default:main" json:"fileCategory"`
	IsMainFile   bool         `gorm:"not null;default:1" json:"isMainFile"`
	MainFileID   *uint        `gorm:"index" json:"mainFileId"`

	// 处理状态管理
	ProcessingStatus  ProcessingStatus `gorm:"not null;default:success" json:"processingStatus"`
	ProcessingMessage string           `gorm:"type:text" json:"processingMessage"`
	RetryCount        int              `gorm:"not null;default:0" json:"retryCount"`
	LastProcessedAt   *time.Time       `json:"lastProcessedAt"`

	// 失效检测相关
	LastCheckedAt        *time.Time `json:"lastCheckedAt"`
	IsValid              *bool      `gorm:"default:1" json:"isValid"`
	ValidationMessage    string     `gorm:"type:text" json:"validationMessage"`
	ValidationRetryCount int        `gorm:"not null;default:0" json:"validationRetryCount"`
	NextCheckAt          *time.Time `json:"nextCheckAt"`

	// 通知相关
	NotificationStatus  NotificationStatus `gorm:"not null;default:0" json:"notificationStatus"`
	EmbyNotified        bool               `gorm:"not null;default:0" json:"embyNotified"`
	TelegramNotified    bool               `gorm:"not null;default:0" json:"telegramNotified"`
	NotificationSentAt  *time.Time         `json:"notificationSentAt"`
	NotificationMessage string             `gorm:"type:text" json:"notificationMessage"`

	// 扩展字段
	MediaInfo string `gorm:"type:text" json:"mediaInfo"` // JSON格式
	Hash      string `gorm:"size:64" json:"hash"`        // 文件哈希值
	Metadata  string `gorm:"type:text" json:"metadata"`  // JSON格式
	Tags      string `gorm:"size:500" json:"tags"`       // 逗号分隔
}

// FileHistoryCreateRequest 创建文件历史记录请求
type FileHistoryCreateRequest struct {
	TaskID         uint         `json:"taskId" binding:"required"`
	TaskLogID      uint         `json:"taskLogId" binding:"required"`
	FileName       string       `json:"fileName" binding:"required"`
	SourcePath     string       `json:"sourcePath" binding:"required"`
	SourceURL      string       `json:"sourceUrl"`
	TargetFilePath string       `json:"targetFilePath" binding:"required"`
	FileSize       int64        `json:"fileSize" binding:"required"`
	FileType       string       `json:"fileType" binding:"required"`
	FileSuffix     string       `json:"fileSuffix" binding:"required"`
	FileCategory   FileCategory `json:"fileCategory"`
	IsMainFile     bool         `json:"isMainFile"`
	MainFileID     *uint        `json:"mainFileId"`
	MediaInfo      string       `json:"mediaInfo"`
	Hash           string       `json:"hash"`
	Metadata       string       `json:"metadata"`
	Tags           string       `json:"tags"`
}

// FileHistoryUpdateRequest 更新文件历史记录请求
type FileHistoryUpdateRequest struct {
	FileName            *string             `json:"fileName"`
	SourcePath          *string             `json:"sourcePath"`
	SourceURL           *string             `json:"sourceUrl"`
	TargetFilePath      *string             `json:"targetFilePath"`
	FileSize            *int64              `json:"fileSize"`
	FileType            *string             `json:"fileType"`
	FileSuffix          *string             `json:"fileSuffix"`
	FileCategory        *FileCategory       `json:"fileCategory"`
	ProcessingStatus    *ProcessingStatus   `json:"processingStatus"`
	ProcessingMessage   *string             `json:"processingMessage"`
	IsValid             *bool               `json:"isValid"`
	ValidationMessage   *string             `json:"validationMessage"`
	NotificationStatus  *NotificationStatus `json:"notificationStatus"`
	EmbyNotified        *bool               `json:"embyNotified"`
	TelegramNotified    *bool               `json:"telegramNotified"`
	NotificationMessage *string             `json:"notificationMessage"`
	MediaInfo           *string             `json:"mediaInfo"`
	Hash                *string             `json:"hash"`
	Metadata            *string             `json:"metadata"`
	Tags                *string             `json:"tags"`
}

// FileHistoryQueryRequest 查询文件历史记录请求
type FileHistoryQueryRequest struct {
	Page               int                 `form:"page" binding:"min=1"`
	PageSize           int                 `form:"pageSize" binding:"min=1,max=100"`
	TaskID             *uint               `form:"taskId"`
	TaskLogID          *uint               `form:"taskLogId"`
	FileName           string              `form:"fileName"`
	FileCategory       *FileCategory       `form:"fileCategory"`
	IsMainFile         *bool               `form:"isMainFile"`
	ProcessingStatus   *ProcessingStatus   `form:"processingStatus"`
	IsValid            *bool               `form:"isValid"`
	NotificationStatus *NotificationStatus `form:"notificationStatus"`
	EmbyNotified       *bool               `form:"embyNotified"`
	TelegramNotified   *bool               `form:"telegramNotified"`
	StartDate          string              `form:"startDate"`
	EndDate            string              `form:"endDate"`
	SortBy             string              `form:"sortBy"`
	SortOrder          string              `form:"sortOrder"`
}

// FileHistoryResponse 文件历史记录响应
type FileHistoryResponse struct {
	FileHistory
	TaskName    string `json:"taskName,omitempty"`
	TaskLogInfo string `json:"taskLogInfo,omitempty"`
}

// FileHistoryStatistics 文件历史统计
type FileHistoryStatistics struct {
	TotalFiles      int64      `json:"totalFiles"`
	MainFiles       int64      `json:"mainFiles"`
	MetadataFiles   int64      `json:"metadataFiles"`
	SubtitleFiles   int64      `json:"subtitleFiles"`
	SuccessFiles    int64      `json:"successFiles"`
	FailedFiles     int64      `json:"failedFiles"`
	PendingFiles    int64      `json:"pendingFiles"`
	SkippedFiles    int64      `json:"skippedFiles"`
	ValidFiles      int64      `json:"validFiles"`
	InvalidFiles    int64      `json:"invalidFiles"`
	NotifiedFiles   int64      `json:"notifiedFiles"`
	UnnotifiedFiles int64      `json:"unnotifiedFiles"`
	TotalSize       int64      `json:"totalSize"`
	AverageFileSize int64      `json:"averageFileSize"`
	LastProcessedAt *time.Time `json:"lastProcessedAt"`
	LastCheckedAt   *time.Time `json:"lastCheckedAt"`
}

// ValidationSummary 失效检测摘要
type ValidationSummary struct {
	TotalFiles     int64           `json:"totalFiles"`
	ValidFiles     int64           `json:"validFiles"`
	InvalidFiles   int64           `json:"invalidFiles"`
	PendingFiles   int64           `json:"pendingFiles"`
	LastCheckTime  *time.Time      `json:"lastCheckTime"`
	NextCheckTime  *time.Time      `json:"nextCheckTime"`
	FailureReasons []FailureReason `json:"failureReasons"`
}

// FailureReason 失效原因统计
type FailureReason struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// NotificationSummary 通知摘要
type NotificationSummary struct {
	TotalFiles           int64      `json:"totalFiles"`
	NotifiedFiles        int64      `json:"notifiedFiles"`
	EmbyNotified         int64      `json:"embyNotified"`
	TelegramNotified     int64      `json:"telegramNotified"`
	FailedNotifications  int64      `json:"failedNotifications"`
	LastNotificationTime *time.Time `json:"lastNotificationTime"`
}
