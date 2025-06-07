package model

import (
	"time"
)

type Task struct {
	ID                 uint       `gorm:"primarykey" json:"id"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	Name               string     `gorm:"not null" json:"name"`
	MediaType          string     `gorm:"not null;default:movie" json:"mediaType"`
	SourcePath         string     `gorm:"not null" json:"sourcePath"`
	TargetPath         string     `gorm:"not null" json:"targetPath"`
	FileSuffix         string     `gorm:"not null" json:"fileSuffix"`
	Overwrite          bool       `gorm:"not null;default:0" json:"overwrite"`
	Enabled            bool       `gorm:"not null;default:1" json:"enabled"`
	Cron               string     `json:"cron"`
	Running            bool       `gorm:"not null;default:0" json:"running"`
	LastRunAt          *time.Time `json:"lastRunAt"`
	DownloadMetadata   bool       `gorm:"not null;default:0" json:"downloadMetadata"`
	DownloadSubtitle   bool       `gorm:"not null;default:0" json:"downloadSubtitle"`
	MetadataExtensions string     `gorm:"default:.nfo,.jpg,.png" json:"metadataExtensions"`
	SubtitleExtensions string     `gorm:"default:.srt,.ass,.ssa" json:"subtitleExtensions"`
}
