package repository

import (
	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/model/filehistory"
	fileHistoryRequest "github.com/MccRay-s/alist2strm/model/filehistory/request"
)

type FileHistoryRepository struct{}

// 包级别的全局实例
var FileHistory = &FileHistoryRepository{}

// GetMainFileList 获取主文件分页列表（只查询 isMainFile = true 的数据）
func (r *FileHistoryRepository) GetMainFileList(req *fileHistoryRequest.FileHistoryListReq) ([]*filehistory.FileHistory, int64, error) {
	db := database.DB
	var fileHistories []*filehistory.FileHistory
	var total int64

	// 构建查询
	query := db.Model(&filehistory.FileHistory{}).Where("is_main_file = ?", true)

	// 任务ID过滤
	if req.TaskID != nil {
		query = query.Where("task_id = ?", *req.TaskID)
	}

	// 关键字搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("file_name LIKE ? OR source_path LIKE ? OR target_file_path LIKE ?", keyword, keyword, keyword)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&fileHistories).Error; err != nil {
		return nil, 0, err
	}

	return fileHistories, total, nil
}

// GetByMainFileID 根据主文件ID查询关联的文件
func (r *FileHistoryRepository) GetByMainFileID(mainFileID uint) ([]*filehistory.FileHistory, error) {
	db := database.DB
	var relatedFiles []*filehistory.FileHistory

	// 查询关联文件（main_file_id = mainFileID）
	if err := db.Where("main_file_id = ?", mainFileID).Order("created_at ASC").Find(&relatedFiles).Error; err != nil {
		return nil, err
	}

	return relatedFiles, nil
}

// GetByID 根据ID获取文件历史记录
func (r *FileHistoryRepository) GetByID(id uint) (*filehistory.FileHistory, error) {
	db := database.DB
	var fileHistory filehistory.FileHistory

	if err := db.First(&fileHistory, id).Error; err != nil {
		return nil, err
	}

	return &fileHistory, nil
}
