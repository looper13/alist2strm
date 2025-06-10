package response

import "time"

// TaskInfo 任务信息响应
type TaskInfo struct {
	ID                 uint       `json:"id"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	Name               string     `json:"name"`
	MediaType          string     `json:"mediaType"`
	SourcePath         string     `json:"sourcePath"`
	TargetPath         string     `json:"targetPath"`
	FileSuffix         string     `json:"fileSuffix"`
	Overwrite          bool       `json:"overwrite"`
	Enabled            bool       `json:"enabled"`
	Cron               string     `json:"cron"`
	Running            bool       `json:"running"`
	LastRunAt          *time.Time `json:"lastRunAt"`
	DownloadMetadata   bool       `json:"downloadMetadata"`
	DownloadSubtitle   bool       `json:"downloadSubtitle"`
	MetadataExtensions string     `json:"metadataExtensions"`
	SubtitleExtensions string     `json:"subtitleExtensions"`
}

// TaskListResp 任务列表响应
type TaskListResp struct {
	List     []TaskInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// TaskExecuteResp 任务执行结果响应
type TaskExecuteResp struct {
	TaskID         uint   `json:"taskId"`         // 任务ID
	TaskName       string `json:"taskName"`       // 任务名称
	IsSync         bool   `json:"isSync"`         // 是否同步执行
	Status         string `json:"status"`         // 执行状态: running, completed, failed, skipped
	StartTime      string `json:"startTime"`      // 开始时间
	EndTime        string `json:"endTime"`        // 结束时间 (异步执行时为空)
	Duration       string `json:"duration"`       // 执行耗时 (异步执行时为0)
	TotalCount     int    `json:"totalCount"`     // 总文件数量
	SuccessCount   int    `json:"successCount"`   // 成功处理数量
	FailedCount    int    `json:"failedCount"`    // 失败数量
	SkippedCount   int    `json:"skippedCount"`   // 跳过数量
	OverwriteCount int    `json:"overwriteCount"` // 覆盖数量
	SubtitleCount  int    `json:"subtitleCount"`  // 字幕文件数量
	MetadataCount  int    `json:"metadataCount"`  // 元数据文件数量
	ErrorFiles     int    `json:"errorFiles"`     // 错误文件数量
	ProcessedBytes int64  `json:"processedBytes"` // 处理的字节数
	Message        string `json:"message"`        // 执行消息
	ErrorMessage   string `json:"errorMessage"`   // 错误信息
}

// TaskStatusResp 任务状态响应
type TaskStatusResp struct {
	TaskID        uint   `json:"taskId"`        // 任务ID
	TaskName      string `json:"taskName"`      // 任务名称
	Enabled       bool   `json:"enabled"`       // 是否启用
	Running       bool   `json:"running"`       // 是否正在运行
	LastRunAt     string `json:"lastRunAt"`     // 最后运行时间
	LastRunResult string `json:"lastRunResult"` // 最后运行结果
	NextRunTime   string `json:"nextRunTime"`   // 下次运行时间（定时任务）
	Status        string `json:"status"`        // 当前状态: idle, running, completed, failed
}
