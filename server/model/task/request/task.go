package request

import "github.com/MccRay-s/alist2strm/model/common/request"

// TaskCreateReq 任务创建请求
type TaskCreateReq struct {
	Name               string `json:"name" binding:"required" validate:"required,min=1,max=100" example:"任务名称"`
	MediaType          string `json:"mediaType" binding:"required" validate:"required,oneof=movie tv" example:"movie"`
	ConfigType         string `json:"configType" binding:"required" validate:"required,oneof=alist clouddrive local" example:"alist"`
	SourcePath         string `json:"sourcePath" binding:"required" validate:"required" example:"源路径"`
	TargetPath         string `json:"targetPath" binding:"required" validate:"required" example:"目标路径"`
	FileSuffix         string `json:"fileSuffix" binding:"required" validate:"required" example:"文件后缀"`
	Overwrite          bool   `json:"overwrite" example:"是否覆盖"`
	Enabled            bool   `json:"enabled" example:"是否启用"`
	Cron               string `json:"cron" validate:"omitempty" example:"定时任务表达式"`
	DownloadMetadata   bool   `json:"downloadMetadata" example:"是否下载刮削数据"`
	DownloadSubtitle   bool   `json:"downloadSubtitle" example:"是否下载字幕"`
	MetadataExtensions string `json:"metadataExtensions" example:"刮削数据文件扩展名"`
	SubtitleExtensions string `json:"subtitleExtensions" example:"字幕文件扩展名"`
}

// TaskUpdateReq 任务更新请求
type TaskUpdateReq struct {
	ID                 uint   `json:"-"` // 通过路径参数传递，不参与JSON绑定和验证
	Name               string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"任务名称"`
	MediaType          string `json:"mediaType,omitempty" validate:"omitempty,oneof=movie tv" example:"movie"`
	ConfigType         string `json:"configType" validate:"required,oneof=alist clouddrive local" example:"alist"`
	SourcePath         string `json:"sourcePath,omitempty" example:"源路径"`
	TargetPath         string `json:"targetPath,omitempty" example:"目标路径"`
	FileSuffix         string `json:"fileSuffix,omitempty" example:"文件后缀"`
	Overwrite          *bool  `json:"overwrite,omitempty" example:"是否覆盖"`
	Enabled            *bool  `json:"enabled,omitempty" example:"是否启用"`
	Cron               string `json:"cron,omitempty" example:"定时任务表达式"`
	DownloadMetadata   *bool  `json:"downloadMetadata,omitempty" example:"是否下载刮削数据"`
	DownloadSubtitle   *bool  `json:"downloadSubtitle,omitempty" example:"是否下载字幕"`
	MetadataExtensions string `json:"metadataExtensions,omitempty" example:"刮削数据文件扩展名"`
	SubtitleExtensions string `json:"subtitleExtensions,omitempty" example:"字幕文件扩展名"`
}

// TaskInfoReq 任务信息查询请求
type TaskInfoReq struct {
	request.GetById
}

// TaskListReq 任务列表查询请求
type TaskListReq struct {
	request.PageInfo
	Name       string `json:"name" form:"name" example:"任务名称筛选"`
	MediaType  string `json:"mediaType" form:"mediaType" example:"媒体类型筛选"`
	ConfigType string `json:"configType" form:"configType" example:"配置类型筛选"`
	Enabled    *bool  `json:"enabled" form:"enabled" example:"启用状态筛选"`
	Running    *bool  `json:"running" form:"running" example:"运行状态筛选"`
}

// TaskAllReq 所有任务查询请求（不分页）
type TaskAllReq struct {
	Name      string `json:"name" form:"name" example:"任务名称筛选"`
	MediaType string `json:"mediaType" form:"mediaType" example:"媒体类型筛选"`
	Enabled   *bool  `json:"enabled" form:"enabled" example:"启用状态筛选"`
	Running   *bool  `json:"running" form:"running" example:"运行状态筛选"`
}

// TaskExecuteReq 任务执行请求
type TaskExecuteReq struct {
	Sync bool `json:"sync" example:"是否同步执行"` // true: 同步执行，false: 异步执行
}

// TaskStatusReq 任务状态查询请求
type TaskStatusReq struct {
	request.GetById
}
