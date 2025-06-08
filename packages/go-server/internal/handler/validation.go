package handler

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListValidationTasksHandler 获取验证任务列表
func ListValidationTasksHandler(c *gin.Context) {
	var req model.ValidationTaskQueryRequest
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

	validationService := service.GetValidationService()
	data, total, err := validationService.ListTasks(&req)
	if err != nil {
		utils.Logger.Error("获取验证任务列表失败", zap.Error(err))
		utils.ResponseError(c, "获取验证任务列表失败")
		return
	}

	utils.ResponsePage(c, data, total, req.Page, req.PageSize)
}

// CreateValidationTaskHandler 创建验证任务
func CreateValidationTaskHandler(c *gin.Context) {
	var req model.ValidationTaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	validationService := service.GetValidationService()
	data, err := validationService.CreateTask(&req)
	if err != nil {
		utils.Logger.Error("创建验证任务失败", zap.Error(err))
		utils.ResponseError(c, "创建验证任务失败")
		return
	}

	utils.ResponseSuccess(c, data)
}

// GetValidationTaskHandler 根据ID获取验证任务
func GetValidationTaskHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	validationService := service.GetValidationService()
	data, err := validationService.GetTaskByID(uint(id))
	if err != nil {
		utils.Logger.Error("获取验证任务失败", zap.Error(err))
		utils.ResponseError(c, "验证任务不存在")
		return
	}

	utils.ResponseSuccess(c, data)
}

// StartValidationTaskHandler 启动验证任务
func StartValidationTaskHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	validationService := service.GetValidationService()
	err = validationService.StartTask(uint(id))
	if err != nil {
		utils.Logger.Error("启动验证任务失败", zap.Error(err))
		utils.ResponseError(c, "启动验证任务失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "任务启动成功", nil)
}

// CancelValidationTaskHandler 取消验证任务
func CancelValidationTaskHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	validationService := service.GetValidationService()
	err = validationService.CancelTask(uint(id))
	if err != nil {
		utils.Logger.Error("取消验证任务失败", zap.Error(err))
		utils.ResponseError(c, "取消验证任务失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "任务取消成功", nil)
}

// DeleteValidationTaskHandler 删除验证任务
func DeleteValidationTaskHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	validationService := service.GetValidationService()
	err = validationService.DeleteTask(uint(id))
	if err != nil {
		utils.Logger.Error("删除验证任务失败", zap.Error(err))
		utils.ResponseError(c, "删除验证任务失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "删除成功", nil)
}

// GetValidationStatisticsHandler 获取验证统计信息
func GetValidationStatisticsHandler(c *gin.Context) {
	validationService := service.GetValidationService()
	stats, err := validationService.GetStatistics()
	if err != nil {
		utils.Logger.Error("获取验证统计信息失败", zap.Error(err))
		utils.ResponseError(c, "获取统计信息失败")
		return
	}

	utils.ResponseSuccess(c, stats)
}

// ValidateFileHandler 验证单个文件
func ValidateFileHandler(c *gin.Context) {
	var req struct {
		FilePath string `json:"filePath" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	validationService := service.GetValidationService()
	isValid, message, err := validationService.ValidateFile(req.FilePath)
	if err != nil {
		utils.Logger.Error("验证文件失败", zap.Error(err))
		utils.ResponseError(c, "验证文件失败")
		return
	}

	utils.ResponseSuccess(c, gin.H{
		"isValid": isValid,
		"message": message,
	})
}

// CleanupInvalidFilesHandler 清理无效文件
func CleanupInvalidFilesHandler(c *gin.Context) {
	validationService := service.GetValidationService()
	cleaned, err := validationService.CleanupInvalidFiles()
	if err != nil {
		utils.Logger.Error("清理无效文件失败", zap.Error(err))
		utils.ResponseError(c, "清理无效文件失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "清理完成", gin.H{"cleanedCount": cleaned})
}
