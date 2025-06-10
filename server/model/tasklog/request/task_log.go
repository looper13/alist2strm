package request

import "github.com/MccRay-s/alist2strm/model/common/request"

// TaskLogCreateReq 任务日志创建请求
type TaskLogCreateReq struct {
	TaskID        uint   `json:"taskId" binding:"required" validate:"required" example:"任务ID"`
	Status        string `json:"status" binding:"required" validate:"required,oneof=running completed failed cancelled" example:"running"`
	Message       string `json:"message" validate:"max=1000" example:"任务执行消息"`
	TotalFile     int    `json:"totalFile" validate:"min=0" example:"总文件数"`
	GeneratedFile int    `json:"generatedFile" validate:"min=0" example:"生成文件数"`
	SkipFile      int    `json:"skipFile" validate:"min=0" example:"跳过文件数"`
	OverwriteFile int    `json:"overwriteFile" validate:"min=0" example:"覆盖文件数"`
	MetadataCount int    `json:"metadataCount" validate:"min=0" example:"元数据文件数"`
	SubtitleCount int    `json:"subtitleCount" validate:"min=0" example:"字幕文件数"`
	FailedCount   int    `json:"failedCount" validate:"min=0" example:"失败文件数"`
}

// TaskLogUpdateReq 任务日志更新请求
type TaskLogUpdateReq struct {
	ID            uint   `json:"-"` // 通过路径参数传递，不参与JSON绑定和验证
	Status        string `json:"status,omitempty" validate:"omitempty,oneof=running completed failed cancelled" example:"completed"`
	Message       string `json:"message,omitempty" validate:"omitempty,max=1000" example:"任务执行完成"`
	Duration      *int64 `json:"duration,omitempty" validate:"omitempty,min=0" example:"执行耗时（秒）"`
	TotalFile     *int   `json:"totalFile,omitempty" validate:"omitempty,min=0" example:"总文件数"`
	GeneratedFile *int   `json:"generatedFile,omitempty" validate:"omitempty,min=0" example:"生成文件数"`
	SkipFile      *int   `json:"skipFile,omitempty" validate:"omitempty,min=0" example:"跳过文件数"`
	OverwriteFile *int   `json:"overwriteFile,omitempty" validate:"omitempty,min=0" example:"覆盖文件数"`
	MetadataCount *int   `json:"metadataCount,omitempty" validate:"omitempty,min=0" example:"元数据文件数"`
	SubtitleCount *int   `json:"subtitleCount,omitempty" validate:"omitempty,min=0" example:"字幕文件数"`
	FailedCount   *int   `json:"failedCount,omitempty" validate:"omitempty,min=0" example:"失败文件数"`
}

// TaskLogInfoReq 任务日志信息查询请求
type TaskLogInfoReq struct {
	request.GetById
}

// TaskLogListReq 任务日志列表查询请求
type TaskLogListReq struct {
	request.PageInfo
	TaskID uint `json:"taskId" form:"taskId" binding:"required" validate:"required" example:"任务ID"`
}
