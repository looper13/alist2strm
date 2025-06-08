package task

import (
	"time"
)

// Task 任务模型
type Task struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	Name               string     `json:"name" gorm:"not null" validate:"required"`
	MediaType          string     `json:"mediaType" gorm:"not null;default:movie"`
	SourcePath         string     `json:"sourcePath" gorm:"not null" validate:"required"`
	TargetPath         string     `json:"targetPath" gorm:"not null" validate:"required"`
	FileSuffix         string     `json:"fileSuffix" gorm:"not null" validate:"required"`
	Overwrite          bool       `json:"overwrite" gorm:"not null;default:false"`
	Enabled            bool       `json:"enabled" gorm:"not null;default:true"`
	Cron               string     `json:"cron"`
	Running            bool       `json:"running" gorm:"not null;default:false"`
	LastRunAt          *time.Time `json:"lastRunAt"`
	DownloadMetadata   bool       `json:"downloadMetadata" gorm:"not null;default:false"`
	DownloadSubtitle   bool       `json:"downloadSubtitle" gorm:"not null;default:false"`
	MetadataExtensions string     `json:"metadataExtensions" gorm:"default:.nfo,.jpg,.png"`
	SubtitleExtensions string     `json:"subtitleExtensions" gorm:"default:.srt,.ass,.ssa"`
}

// TableName 表名
func (Task) TableName() string {
	return "tasks"
}
