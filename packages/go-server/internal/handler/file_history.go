package handler

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListFileHistoriesHandler 获取文件历史记录列表
func ListFileHistoriesHandler(c *gin.Context) {
	var req model.FileHistoryQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	fileHistoryService := service.GetFileHistoryService()
	data, total, err := fileHistoryService.List(&req)
	if err != nil {
		utils.Logger.Error("获取文件历史记录列表失败", zap.Error(err))
		utils.ResponseError(c, "获取文件历史记录列表失败")
		return
	}

	utils.ResponsePage(c, data, total, req.Page, req.PageSize)
}

// GetFileHistoryHandler 根据ID获取文件历史记录
func GetFileHistoryHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	fileHistoryService := service.GetFileHistoryService()
	data, err := fileHistoryService.GetByID(uint(id))
	if err != nil {
		utils.Logger.Error("获取文件历史记录失败", zap.Error(err))
		utils.ResponseError(c, "文件历史记录不存在")
		return
	}

	utils.ResponseSuccess(c, data)
}

// CreateFileHistoryHandler 创建文件历史记录
func CreateFileHistoryHandler(c *gin.Context) {
	var req model.FileHistoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	fileHistoryService := service.GetFileHistoryService()
	data, err := fileHistoryService.Create(&req)
	if err != nil {
		utils.Logger.Error("创建文件历史记录失败", zap.Error(err))
		utils.ResponseError(c, "创建文件历史记录失败")
		return
	}

	utils.ResponseSuccess(c, data)
}

// UpdateFileHistoryHandler 更新文件历史记录
func UpdateFileHistoryHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	var req model.FileHistoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	fileHistoryService := service.GetFileHistoryService()
	err = fileHistoryService.Update(uint(id), &req)
	if err != nil {
		utils.Logger.Error("更新文件历史记录失败", zap.Error(err))
		utils.ResponseError(c, "更新文件历史记录失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "更新成功", nil)
}

// DeleteFileHistoryHandler 删除文件历史记录
func DeleteFileHistoryHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	fileHistoryService := service.GetFileHistoryService()
	err = fileHistoryService.Delete(uint(id))
	if err != nil {
		utils.Logger.Error("删除文件历史记录失败", zap.Error(err))
		utils.ResponseError(c, "删除文件历史记录失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "删除成功", nil)
}

// BatchDeleteFileHistoriesHandler 批量删除文件历史记录
func BatchDeleteFileHistoriesHandler(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	fileHistoryService := service.GetFileHistoryService()
	if err := fileHistoryService.BatchDelete(req.IDs); err != nil {
		utils.Logger.Error("批量删除文件历史记录失败", zap.Error(err))
		utils.ResponseError(c, "批量删除失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "批量删除成功", gin.H{"deletedCount": len(req.IDs)})
}

// ClearAllFileHistoriesHandler 清空所有文件历史记录
func ClearAllFileHistoriesHandler(c *gin.Context) {
	fileHistoryService := service.GetFileHistoryService()
	err := fileHistoryService.ClearAll()
	if err != nil {
		utils.Logger.Error("清空文件历史记录失败", zap.Error(err))
		utils.ResponseError(c, "清空操作失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "清空成功", nil)
}

// GetFileHistoryStatisticsHandler 获取文件历史统计信息
func GetFileHistoryStatisticsHandler(c *gin.Context) {
	fileHistoryService := service.GetFileHistoryService()
	stats, err := fileHistoryService.GetStatistics()
	if err != nil {
		utils.Logger.Error("获取文件历史统计信息失败", zap.Error(err))
		utils.ResponseError(c, "获取统计信息失败")
		return
	}

	utils.ResponseSuccess(c, stats)
}

// GetFileHistoryValidationSummaryHandler 获取验证摘要
func GetFileHistoryValidationSummaryHandler(c *gin.Context) {
	fileHistoryService := service.GetFileHistoryService()
	summary, err := fileHistoryService.GetValidationSummary()
	if err != nil {
		utils.Logger.Error("获取验证摘要失败", zap.Error(err))
		utils.ResponseError(c, "获取验证摘要失败")
		return
	}

	utils.ResponseSuccess(c, summary)
}

// GetFileHistoryNotificationSummaryHandler 获取通知摘要
func GetFileHistoryNotificationSummaryHandler(c *gin.Context) {
	fileHistoryService := service.GetFileHistoryService()
	summary, err := fileHistoryService.GetNotificationSummary()
	if err != nil {
		utils.Logger.Error("获取通知摘要失败", zap.Error(err))
		utils.ResponseError(c, "获取通知摘要失败")
		return
	}

	utils.ResponseSuccess(c, summary)
}

// MarkFileHistoryAsProcessedHandler 标记文件为已处理
func MarkFileHistoryAsProcessedHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	var req struct {
		Status  model.ProcessingStatus `json:"status" binding:"required"`
		Message string                 `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	fileHistoryService := service.GetFileHistoryService()
	if err := fileHistoryService.MarkAsProcessed(uint(id), req.Status, req.Message); err != nil {
		utils.Logger.Error("标记文件处理状态失败", zap.Error(err))
		utils.ResponseError(c, "标记处理状态失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "标记成功", nil)
}

// MarkFileHistoryAsValidatedHandler 标记文件为已验证
func MarkFileHistoryAsValidatedHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	var req struct {
		IsValid bool   `json:"isValid" binding:"required"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	fileHistoryService := service.GetFileHistoryService()
	if err := fileHistoryService.MarkAsValidated(uint(id), req.IsValid, req.Message); err != nil {
		utils.Logger.Error("标记文件验证状态失败", zap.Error(err))
		utils.ResponseError(c, "标记验证状态失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "标记成功", nil)
}
