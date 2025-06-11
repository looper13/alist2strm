package request

// FileHistoryListReq 文件历史分页查询请求
type FileHistoryListReq struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
	TaskID   *uint  `json:"taskId" form:"taskId"`
	Keyword  string `json:"keyword" form:"keyword"` // 可搜索文件名、源路径、目标路径
}

// FileHistoryByMainFileReq 根据主文件ID查询关联文件请求
type FileHistoryByMainFileReq struct {
	MainFileID uint `json:"mainFileId" form:"mainFileId" binding:"required"`
}
