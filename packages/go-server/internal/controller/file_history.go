package controller

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FileHistoryController struct {
	fileHistoryService *service.FileHistoryService
	logger             *zap.Logger
}

// NewFileHistoryController 创建文件历史控制器
func NewFileHistoryController() *FileHistoryController {
	return &FileHistoryController{
		fileHistoryService: service.GetFileHistoryService(),
		logger:             utils.Logger,
	}
}

// RegisterRoutes 注册路由
func (c *FileHistoryController) RegisterRoutes(router *gin.RouterGroup) {
	fileHistoryGroup := router.Group("/file-histories")
	{
		fileHistoryGroup.GET("", c.List)
		fileHistoryGroup.GET("/:id", c.GetByID)
		fileHistoryGroup.POST("", c.Create)
		fileHistoryGroup.PUT("/:id", c.Update)
		fileHistoryGroup.DELETE("/:id", c.Delete)
		fileHistoryGroup.DELETE("/batch", c.BatchDelete)
		fileHistoryGroup.DELETE("/clear", c.ClearAll)

		// 统计相关
		fileHistoryGroup.GET("/statistics", c.GetStatistics)
		fileHistoryGroup.GET("/validation-summary", c.GetValidationSummary)
		fileHistoryGroup.GET("/notification-summary", c.GetNotificationSummary)

		// 状态管理
		fileHistoryGroup.PUT("/:id/processed", c.MarkAsProcessed)
		fileHistoryGroup.PUT("/:id/validated", c.MarkAsValidated)
	}
}

// List 获取文件历史记录列表
func (c *FileHistoryController) List(ctx *gin.Context) {
	var req model.FileHistoryQueryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
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

	data, total, err := c.fileHistoryService.List(&req)
	if err != nil {
		c.logger.Error("获取文件历史记录列表失败", zap.Error(err))
		utils.ResponseError(ctx, "获取文件历史记录列表失败")
		return
	}

	utils.ResponsePage(ctx, data, total, req.Page, req.PageSize)
}

// GetByID 根据ID获取文件历史记录
func (c *FileHistoryController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	data, err := c.fileHistoryService.GetByID(uint(id))
	if err != nil {
		c.logger.Error("获取文件历史记录失败", zap.Error(err))
		utils.ResponseError(ctx, "文件历史记录不存在")
		return
	}

	utils.ResponseSuccess(ctx, data)
}

// Create 创建文件历史记录
func (c *FileHistoryController) Create(ctx *gin.Context) {
	var req model.FileHistoryCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	data, err := c.fileHistoryService.Create(&req)
	if err != nil {
		c.logger.Error("创建文件历史记录失败", zap.Error(err))
		utils.ResponseError(ctx, "创建文件历史记录失败")
		return
	}

	utils.ResponseSuccess(ctx, data)
}

// Update 更新文件历史记录
func (c *FileHistoryController) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	var req model.FileHistoryUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	err = c.fileHistoryService.Update(uint(id), &req)
	if err != nil {
		c.logger.Error("更新文件历史记录失败", zap.Error(err))
		utils.ResponseError(ctx, "更新文件历史记录失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "更新成功", nil)
}

// Delete 删除文件历史记录
func (c *FileHistoryController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	err = c.fileHistoryService.Delete(uint(id))
	if err != nil {
		c.logger.Error("删除文件历史记录失败", zap.Error(err))
		utils.ResponseError(ctx, "删除文件历史记录失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "删除成功", nil)
}

// BatchDelete 批量删除文件历史记录
func (c *FileHistoryController) BatchDelete(ctx *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	if err := c.fileHistoryService.BatchDelete(req.IDs); err != nil {
		c.logger.Error("批量删除文件历史记录失败", zap.Error(err))
		utils.ResponseError(ctx, "批量删除失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "批量删除成功", gin.H{"deletedCount": len(req.IDs)})
}

// ClearAll 清空所有文件历史记录
func (c *FileHistoryController) ClearAll(ctx *gin.Context) {
	err := c.fileHistoryService.ClearAll()
	if err != nil {
		c.logger.Error("清空文件历史记录失败", zap.Error(err))
		utils.ResponseError(ctx, "清空操作失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "清空成功", nil)
}

// GetStatistics 获取文件历史统计信息
func (c *FileHistoryController) GetStatistics(ctx *gin.Context) {
	stats, err := c.fileHistoryService.GetStatistics()
	if err != nil {
		c.logger.Error("获取文件历史统计信息失败", zap.Error(err))
		utils.ResponseError(ctx, "获取统计信息失败")
		return
	}

	utils.ResponseSuccess(ctx, stats)
}

// GetValidationSummary 获取验证摘要
func (c *FileHistoryController) GetValidationSummary(ctx *gin.Context) {
	summary, err := c.fileHistoryService.GetValidationSummary()
	if err != nil {
		c.logger.Error("获取验证摘要失败", zap.Error(err))
		utils.ResponseError(ctx, "获取验证摘要失败")
		return
	}

	utils.ResponseSuccess(ctx, summary)
}

// GetNotificationSummary 获取通知摘要
func (c *FileHistoryController) GetNotificationSummary(ctx *gin.Context) {
	summary, err := c.fileHistoryService.GetNotificationSummary()
	if err != nil {
		c.logger.Error("获取通知摘要失败", zap.Error(err))
		utils.ResponseError(ctx, "获取通知摘要失败")
		return
	}

	utils.ResponseSuccess(ctx, summary)
}

// MarkAsProcessed 标记文件为已处理
func (c *FileHistoryController) MarkAsProcessed(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	var req struct {
		Status  model.ProcessingStatus `json:"status" binding:"required"`
		Message string                 `json:"message"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	if err := c.fileHistoryService.MarkAsProcessed(uint(id), req.Status, req.Message); err != nil {
		c.logger.Error("标记文件处理状态失败", zap.Error(err))
		utils.ResponseError(ctx, "标记处理状态失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "标记成功", nil)
}

// MarkAsValidated 标记文件为已验证
func (c *FileHistoryController) MarkAsValidated(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	var req struct {
		IsValid bool   `json:"isValid" binding:"required"`
		Message string `json:"message"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	if err := c.fileHistoryService.MarkAsValidated(uint(id), req.IsValid, req.Message); err != nil {
		c.logger.Error("标记文件验证状态失败", zap.Error(err))
		utils.ResponseError(ctx, "标记验证状态失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "标记成功", nil)
}
