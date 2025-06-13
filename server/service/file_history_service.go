package service

import (
	"errors"

	fileHistoryRequest "github.com/MccRay-s/alist2strm/model/filehistory/request"
	fileHistoryResponse "github.com/MccRay-s/alist2strm/model/filehistory/response"
	"github.com/MccRay-s/alist2strm/repository"
	"gorm.io/gorm"
)

type FileHistoryService struct{}

// 包级别的全局实例
var FileHistoryServiceApp = &FileHistoryService{}

// GetMainFileList 获取主文件分页列表
func (s *FileHistoryService) GetFileList(req *fileHistoryRequest.FileHistoryListReq) (*fileHistoryResponse.FileHistoryListResp, error) {
	fileHistories, total, err := repository.FileHistory.GetFileList(req)
	if err != nil {
		return nil, err
	}

	return &fileHistoryResponse.FileHistoryListResp{
		List:  fileHistories,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// GetFileHistoryInfo 获取文件历史详情
func (s *FileHistoryService) GetFileHistoryInfo(id uint) (*fileHistoryResponse.FileHistoryInfoResp, error) {
	fileHistory, err := repository.FileHistory.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件历史记录不存在")
		}
		return nil, err
	}

	return &fileHistoryResponse.FileHistoryInfoResp{
		FileHistory: fileHistory,
	}, nil
}
