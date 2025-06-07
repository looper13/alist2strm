package model

// AlistConfig Alist配置结构
type AlistConfig struct {
	Token            string `json:"token"`            // Alist的认证Token
	Host             string `json:"host"`             // Alist服务器地址
	ReqInterval      int    `json:"reqInterval"`      // 请求间隔(毫秒)
	ReqRetryCount    int    `json:"reqRetryCount"`    // 请求重试次数
	ReqRetryInterval int    `json:"reqRetryInterval"` // 重试间隔(毫秒)
}

// FileInfo Alist文件信息结构
type FileInfo struct {
	Name      string `json:"name"`      // 文件名
	Path      string `json:"path"`      // 完整路径
	Size      int64  `json:"size"`      // 文件大小
	Type      int    `json:"type"`      // 类型：1目录，2文件
	Modified  string `json:"modified"`  // 修改时间
	Sign      string `json:"sign"`      // 签名（用于下载）
	Thumbnail string `json:"thumbnail"` // 缩略图
}

// ListResponse Alist目录列表响应
type ListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Content []FileInfo `json:"content"`
		Total   int        `json:"total"`
	} `json:"data"`
}

// FsGetResponse Alist文件信息响应
type FsGetResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    FileInfo `json:"data"`
}
