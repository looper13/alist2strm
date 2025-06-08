package controller

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ValidationController struct {
	validationService *service.ValidationService
	logger            *zap.Logger
}

// NewValidationController 创建验证控制器
func NewValidationController() *ValidationController {
	return &ValidationController{
		validationService: service.GetValidationService(),
		logger:            utils.Logger,
	}
}

// RegisterRoutes 注册路由
func (c *ValidationController) RegisterRoutes(router *gin.RouterGroup) {
	validationGroup := router.Group("/validation")
	{
		// 验证任务管理
		validationGroup.GET("/tasks", c.ListValidationTasks)
		validationGroup.POST("/tasks", c.CreateValidationTask)
		validationGroup.GET("/tasks/:id", c.GetValidationTask)
		validationGroup.POST("/tasks/:id/start", c.StartValidationTask)
		validationGroup.POST("/tasks/:id/cancel", c.CancelValidationTask)
		validationGroup.DELETE("/tasks/:id", c.DeleteValidationTask)

		// 验证统计和操作
		validationGroup.GET("/statistics", c.GetValidationStatistics)
		validationGroup.POST("/validate-file", c.ValidateFile)
		validationGroup.POST("/cleanup-invalid", c.CleanupInvalidFiles)
	}
}

// ListValidationTasks 获取验证任务列表
func (c *ValidationController) ListValidationTasks(ctx *gin.Context) {
	var req model.ValidationTaskQueryRequest
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

	data, total, err := c.validationService.ListTasks(&req)
	if err != nil {
		c.logger.Error("获取验证任务列表失败", zap.Error(err))
		utils.ResponseError(ctx, "获取验证任务列表失败")
		return
	}

	utils.ResponsePage(ctx, data, total, req.Page, req.PageSize)
}

// CreateValidationTask 创建验证任务
func (c *ValidationController) CreateValidationTask(ctx *gin.Context) {
	var req model.ValidationTaskCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	data, err := c.validationService.CreateTask(&req)
	if err != nil {
		c.logger.Error("创建验证任务失败", zap.Error(err))
		utils.ResponseError(ctx, "创建验证任务失败")
		return
	}

	utils.ResponseSuccess(ctx, data)
}

// GetValidationTask 根据ID获取验证任务
func (c *ValidationController) GetValidationTask(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	data, err := c.validationService.GetTaskByID(uint(id))
	if err != nil {
		c.logger.Error("获取验证任务失败", zap.Error(err))
		utils.ResponseError(ctx, "验证任务不存在")
		return
	}

	utils.ResponseSuccess(ctx, data)
}

// StartValidationTask 启动验证任务
func (c *ValidationController) StartValidationTask(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	err = c.validationService.StartTask(uint(id))
	if err != nil {
		c.logger.Error("启动验证任务失败", zap.Error(err))
		utils.ResponseError(ctx, "启动验证任务失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "任务启动成功", nil)
}

// CancelValidationTask 取消验证任务
func (c *ValidationController) CancelValidationTask(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	err = c.validationService.CancelTask(uint(id))
	if err != nil {
		c.logger.Error("取消验证任务失败", zap.Error(err))
		utils.ResponseError(ctx, "取消验证任务失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "任务取消成功", nil)
}

// DeleteValidationTask 删除验证任务
func (c *ValidationController) DeleteValidationTask(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	err = c.validationService.DeleteTask(uint(id))
	if err != nil {
		c.logger.Error("删除验证任务失败", zap.Error(err))
		utils.ResponseError(ctx, "删除验证任务失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "删除成功", nil)
}

// GetValidationStatistics 获取验证统计信息
func (c *ValidationController) GetValidationStatistics(ctx *gin.Context) {
	stats, err := c.validationService.GetStatistics()
	if err != nil {
		c.logger.Error("获取验证统计信息失败", zap.Error(err))
		utils.ResponseError(ctx, "获取统计信息失败")
		return
	}

	utils.ResponseSuccess(ctx, stats)
}

// ValidateFile 验证单个文件
func (c *ValidationController) ValidateFile(ctx *gin.Context) {
	var req struct {
		FilePath string `json:"filePath" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	isValid, message, err := c.validationService.ValidateFile(req.FilePath)
	if err != nil {
		c.logger.Error("验证文件失败", zap.Error(err))
		utils.ResponseError(ctx, "验证文件失败")
		return
	}

	utils.ResponseSuccess(ctx, gin.H{
		"isValid": isValid,
		"message": message,
	})
}

// CleanupInvalidFiles 清理无效文件
func (c *ValidationController) CleanupInvalidFiles(ctx *gin.Context) {
	cleaned, err := c.validationService.CleanupInvalidFiles()
	if err != nil {
		c.logger.Error("清理无效文件失败", zap.Error(err))
		utils.ResponseError(ctx, "清理无效文件失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "清理完成", gin.H{"cleanedCount": cleaned})
}
