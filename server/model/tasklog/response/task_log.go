package response

import (
	"time"

	"github.com/MccRay-s/alist2strm/model/tasklog"
)

// TaskLogInfoResp 任务日志信息响应
type TaskLogInfoResp struct {
	tasklog.TaskLog
}

// TaskLogListResp 任务日志列表响应
type TaskLogListResp struct {
	List     []tasklog.TaskLog `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// TaskLogCreateResp 任务日志创建响应
type TaskLogCreateResp struct {
	ID        uint      `json:"id"`
	TaskID    uint      `json:"taskId"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"startTime"`
	Message   string    `json:"message"`
}

// FileProcessingStatsResp 文件处理统计数据响应
type FileProcessingStatsResp struct {
	TotalFiles         int64 `json:"totalFiles"`         // 扫描的文件总数
	ProcessedFiles     int64 `json:"processedFiles"`     // 已处理的文件数
	SkippedFiles       int64 `json:"skippedFiles"`       // 跳过处理的文件数
	StrmGenerated      int64 `json:"strmGenerated"`      // 生成的STRM文件数
	MetadataDownloaded int64 `json:"metadataDownloaded"` // 下载的元数据文件数
	SubtitleDownloaded int64 `json:"subtitleDownloaded"` // 下载的字幕文件数
}
