package repository

import (
	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/model/filehistory"
	fileHistoryRequest "github.com/MccRay-s/alist2strm/model/filehistory/request"
)

type FileHistoryRepository struct{}

// 包级别的全局实例
var FileHistory = &FileHistoryRepository{}

// 获取文件分页列表
func (r *FileHistoryRepository) GetFileList(req *fileHistoryRequest.FileHistoryListReq) ([]*filehistory.FileHistory, int64, error) {
	db := database.DB
	var fileHistories []*filehistory.FileHistory
	var total int64

	// 构建查询
	query := db.Model(&filehistory.FileHistory{}).Where("is_main_file = ?", true)

	// 任务ID过滤
	if req.TaskID != nil {
		query = query.Where("task_id = ?", *req.TaskID)
	}
	// 任务日志ID过滤
	if req.TaskLogID != nil {
		query = query.Where("task_log_id = ?", *req.TaskLogID)
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

// GetByID 根据ID获取文件历史记录
func (r *FileHistoryRepository) GetByID(id uint) (*filehistory.FileHistory, error) {
	db := database.DB
	var fileHistory filehistory.FileHistory

	if err := db.First(&fileHistory, id).Error; err != nil {
		return nil, err
	}

	return &fileHistory, nil
}
