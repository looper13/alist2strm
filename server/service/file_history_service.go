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
func (s *FileHistoryService) GetMainFileList(req *fileHistoryRequest.FileHistoryListReq) (*fileHistoryResponse.FileHistoryListResp, error) {
	fileHistories, total, err := repository.FileHistory.GetMainFileList(req)
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

// GetRelatedFilesByMainID 根据主文件ID查询关联文件
func (s *FileHistoryService) GetRelatedFilesByMainID(mainFileID uint) (*fileHistoryResponse.FileHistoryRelatedResp, error) {
	// 先获取主文件信息
	mainFile, err := repository.FileHistory.GetByID(mainFileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("主文件记录不存在")
		}
		return nil, err
	}

	// 检查是否为主文件
	if !mainFile.IsMainFile {
		return nil, errors.New("指定的文件不是主文件")
	}

	// 获取关联文件
	relatedFiles, err := repository.FileHistory.GetByMainFileID(mainFileID)
	if err != nil {
		return nil, err
	}

	return &fileHistoryResponse.FileHistoryRelatedResp{
		MainFile:     mainFile,
		RelatedFiles: relatedFiles,
	}, nil
}
