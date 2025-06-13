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
	FileName       string     `json:"fileName" gorm:"not null;index"`
	SourcePath     string     `json:"sourcePath" gorm:"not null;index"`
	TargetFilePath string     `json:"targetFilePath" gorm:"not null"`
	FileSize       int64      `json:"fileSize" gorm:"not null"`
	FileType       string     `json:"fileType" gorm:"not null"`
	FileSuffix     string     `json:"fileSuffix" gorm:"not null"`
	ModifiedAt     *time.Time `json:"modifiedAt" gorm:"index"`
	Hash           string     `json:"hash" gorm:"size:64"`
}

// TableName 表名
func (FileHistory) TableName() string {
	return "file_histories"
}
