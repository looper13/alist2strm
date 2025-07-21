package task

import (
	"time"
)

// TaskStats 任务统计数据结构
type TaskStats struct {
	Total           int64 // 任务总数
	Enabled         int64 // 已启用任务数
	Disabled        int64 // 已禁用任务数
	TotalExecutions int64 // 总执行次数
	SuccessCount    int64 // 成功执行次数
	FailedCount     int64 // 失败执行次数
}

// Task 任务模型
type Task struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	Name               string     `json:"name" gorm:"type:VARCHAR(255);not null" validate:"required"`
	MediaType          string     `json:"mediaType" gorm:"type:VARCHAR(50);not null;default:movie"`  // 媒体类型：movie/tv
	ConfigType         string     `json:"configType" gorm:"type:VARCHAR(10);not null;default:alist"` // 配置类型：alist/cloudrive/local
	SourcePath         string     `json:"sourcePath" gorm:"type:VARCHAR(255);not null" validate:"required"`
	TargetPath         string     `json:"targetPath" gorm:"type:VARCHAR(255);not null" validate:"required"`
	FileSuffix         string     `json:"fileSuffix" gorm:"type:VARCHAR(255);not null" validate:"required"`
	Overwrite          bool       `json:"overwrite" gorm:"type:TINYINT(1);not null;default:0"`
	Enabled            bool       `json:"enabled" gorm:"type:TINYINT(1);not null;default:1"`
	Cron               string     `json:"cron" gorm:"type:VARCHAR(255)"`
	Running            bool       `json:"running" gorm:"type:TINYINT(1);not null;default:0"`
	LastRunAt          *time.Time `json:"lastRunAt"`
	DownloadMetadata   bool       `json:"downloadMetadata" gorm:"type:TINYINT(1);not null;default:0"`      // 是否下载刮削数据
	DownloadSubtitle   bool       `json:"downloadSubtitle" gorm:"type:TINYINT(1);not null;default:0"`      // 是否下载字幕
	MetadataExtensions string     `json:"metadataExtensions" gorm:"type:VARCHAR(255);default:nfo,jpg,png"` // 刮削数据文件扩展名
	SubtitleExtensions string     `json:"subtitleExtensions" gorm:"type:VARCHAR(255);default:srt,ass,ssa"` // 字幕文件扩展名
}

// TableName 表名
func (Task) TableName() string {
	return "tasks"
}
