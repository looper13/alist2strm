package response

import "github.com/MccRay-s/alist2strm/model/filehistory"

// FileHistoryInfoResp 文件历史详情响应
type FileHistoryInfoResp struct {
	*filehistory.FileHistory
}

// FileHistoryListResp 文件历史分页列表响应
type FileHistoryListResp struct {
	List  []*filehistory.FileHistory `json:"list"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page"`
	Size  int                        `json:"size"`
}

// FileHistoryRelatedResp 根据主文件ID查询关联文件响应
type FileHistoryRelatedResp struct {
	MainFile     *filehistory.FileHistory   `json:"mainFile"`
	RelatedFiles []*filehistory.FileHistory `json:"relatedFiles"`
}
