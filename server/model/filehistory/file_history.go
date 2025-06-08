package filehistory

import (
	"time"
)

// FileHistory 文件历史模型
type FileHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TaskID    uint      `json:"taskId" gorm:"not null;index"`
	TaskLogID uint      `json:"taskLogId" gorm:"not null;index"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 文件基本信息
	FileName       string `json:"fileName" gorm:"not null;index"`
	SourcePath     string `json:"sourcePath" gorm:"not null;index"`
	SourceURL      string `json:"sourceUrl" gorm:"size:512"`
	TargetFilePath string `json:"targetFilePath" gorm:"not null"`
	FileSize       int64  `json:"fileSize" gorm:"not null"`
	FileType       string `json:"fileType" gorm:"not null"`
	FileSuffix     string `json:"fileSuffix" gorm:"not null"`

	// 文件分类和关联
	FileCategory string `json:"fileCategory" gorm:"not null;default:main;index"`
	IsMainFile   bool   `json:"isMainFile" gorm:"not null;default:true"`
	MainFileID   *uint  `json:"mainFileId" gorm:"index"`

	// 处理状态管理
	ProcessingStatus  string     `json:"processingStatus" gorm:"not null;default:success;index"`
	ProcessingMessage string     `json:"processingMessage" gorm:"type:text"`
	RetryCount        int        `json:"retryCount" gorm:"not null;default:0"`
	LastProcessedAt   *time.Time `json:"lastProcessedAt"`

	// 失效检测相关
	LastCheckedAt        *time.Time `json:"lastCheckedAt" gorm:"index"`
	IsValid              *bool      `json:"isValid" gorm:"default:true;index"`
	ValidationMessage    string     `json:"validationMessage" gorm:"type:text"`
	ValidationRetryCount int        `json:"validationRetryCount" gorm:"not null;default:0"`
	NextCheckAt          *time.Time `json:"nextCheckAt" gorm:"index"`

	// 通知相关
	NotificationStatus  int        `json:"notificationStatus" gorm:"not null;default:0;index"`
	EmbyNotified        bool       `json:"embyNotified" gorm:"not null;default:false"`
	TelegramNotified    bool       `json:"telegramNotified" gorm:"not null;default:false"`
	NotificationSentAt  *time.Time `json:"notificationSentAt"`
	NotificationMessage string     `json:"notificationMessage" gorm:"type:text"`

	// 扩展字段
	MediaInfo string `json:"mediaInfo" gorm:"type:text"`
	Hash      string `json:"hash" gorm:"size:64"`
	Metadata  string `json:"metadata" gorm:"type:text"`
	Tags      string `json:"tags" gorm:"size:500"`
}

// TableName 表名
func (FileHistory) TableName() string {
	return "file_histories"
}
