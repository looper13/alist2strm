package controller

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TaskController struct {
	taskService *service.TaskService
	logger      *zap.Logger
}

// NewTaskController 创建任务控制器
func NewTaskController() *TaskController {
	return &TaskController{
		taskService: service.GetTaskService(),
		logger:      utils.Logger,
	}
}

// RegisterRoutes 注册路由
func (c *TaskController) RegisterRoutes(router *gin.RouterGroup) {
	taskGroup := router.Group("/tasks")
	{
		taskGroup.POST("", c.CreateTask)
		taskGroup.PUT("/:id", c.UpdateTask)
		taskGroup.DELETE("/:id", c.DeleteTask)
		taskGroup.GET("/:id", c.GetTask)
		taskGroup.GET("/all", c.ListTasks)
		taskGroup.PUT("/:id/status", c.SetTaskStatus)
		taskGroup.PUT("/:id/reset", c.ResetTaskStatus)
		taskGroup.POST("/:id/execute", c.ExecuteTask)
	}
}

// CreateTask 创建任务
func (c *TaskController) CreateTask(ctx *gin.Context) {
	var req service.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析创建任务请求失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	task, err := c.taskService.CreateTask(&req)
	if err != nil {
		c.logger.Error("创建任务失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("创建任务成功", zap.String("name", req.Name))
	utils.ResponseSuccess(ctx, task)
}

// UpdateTask 更新任务
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的任务ID")
		return
	}

	var req service.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析更新任务请求失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	task, err := c.taskService.UpdateTask(uint(id), &req)
	if err != nil {
		c.logger.Error("更新任务失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("更新任务成功", zap.Uint("id", uint(id)))
	utils.ResponseSuccess(ctx, task)
}

// DeleteTask 删除任务
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的任务ID")
		return
	}

	if err := c.taskService.DeleteTask(uint(id)); err != nil {
		c.logger.Error("删除任务失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("删除任务成功", zap.Uint("id", uint(id)))
	utils.ResponseSuccessWithMessage(ctx, "任务已删除", nil)
}

// GetTask 获取任务详情
func (c *TaskController) GetTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的任务ID")
		return
	}

	task, err := c.taskService.GetTaskByID(uint(id))
	if err != nil {
		c.logger.Error("获取任务失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	utils.ResponseSuccess(ctx, task)
}

// ListTasks 获取任务列表
func (c *TaskController) ListTasks(ctx *gin.Context) {
	name := ctx.Query("name")

	tasks, err := c.taskService.ListTasks(name)
	if err != nil {
		c.logger.Error("获取任务列表失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	utils.ResponseSuccess(ctx, tasks)
}

// SetTaskStatus 设置任务状态
func (c *TaskController) SetTaskStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的任务ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析任务状态请求失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	if err := c.taskService.SetTaskStatus(uint(id), req.Enabled); err != nil {
		c.logger.Error("设置任务状态失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("任务状态更新成功", zap.Uint("id", uint(id)), zap.Bool("enabled", req.Enabled))
	utils.ResponseSuccessWithMessage(ctx, "任务状态已更新", nil)
}

// ResetTaskStatus 重置任务运行状态
func (c *TaskController) ResetTaskStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的任务ID")
		return
	}

	if err := c.taskService.ResetRunningStatus(uint(id)); err != nil {
		c.logger.Error("重置任务状态失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("任务运行状态重置成功", zap.Uint("id", uint(id)))
	utils.ResponseSuccessWithMessage(ctx, "任务运行状态已重置", nil)
}

// ExecuteTask 执行指定任务
func (c *TaskController) ExecuteTask(ctx *gin.Context) {
	taskID := ctx.Param("id")
	if taskID == "" {
		c.logger.Error("任务ID不能为空")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 将taskID转换为uint
	taskIDUint, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		c.logger.Error("任务ID格式错误", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "任务ID格式错误"})
		return
	}

	// 获取任务信息
	task := &model.Task{}
	if err := model.DB.First(task, taskID).Error; err != nil {
		c.logger.Error("获取任务信息失败", zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 获取TaskLogService实例
	taskLogService := service.GetTaskLogService(model.DB, c.logger)

	// 创建任务日志记录
	taskLog, err := taskLogService.StartTask(uint(taskIDUint), "开始执行任务")
	if err != nil {
		c.logger.Error("创建任务日志失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务日志失败"})
		return
	}

	// 获取 Alist 配置
	alistConfig := &model.Config{}
	if err := model.DB.Where("code = ?", "ALIST").First(alistConfig).Error; err != nil {
		c.logger.Error("获取Alist配置失败", zap.Error(err))
		// 标记任务失败
		taskLogService.MarkAsFailed(taskLog.ID, "获取Alist配置失败: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "未找到Alist配置"})
		return
	}

	// 解析 Alist 配置
	var alistConfigValue service.AlistConfig
	if err := json.Unmarshal([]byte(alistConfig.Value), &alistConfigValue); err != nil {
		c.logger.Error("解析Alist配置失败", zap.Error(err))
		// 标记任务失败
		taskLogService.MarkAsFailed(taskLog.ID, "解析Alist配置失败: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Alist配置格式错误"})
		return
	}

	// 创建文件服务实例
	fileService := service.NewFileService(c.logger)

	// 设置任务配置
	taskConfig := &service.TaskConfig{
		SourcePath: task.SourcePath,
		TargetPath: task.TargetPath,
		MediaType:  service.MediaType(task.MediaType),
		Overwrite:  task.Overwrite,
		FileSuffix: strings.Split(task.FileSuffix, ","),
	}

	// 处理元数据配置
	if task.DownloadMetadata {
		taskConfig.DownloadMetadata = true
		taskConfig.MetadataExtensions = strings.Split(task.MetadataExtensions, ",")
	}

	// 处理字幕配置
	if task.DownloadSubtitle {
		taskConfig.DownloadSubtitle = true
		taskConfig.SubtitleExtensions = strings.Split(task.SubtitleExtensions, ",")
	}

	// 设置任务配置
	fileService.SetTaskConfig(taskConfig)

	// 设置 Alist 客户端配置
	alistClient := service.GetAlistClient(c.logger)
	alistClient.UpdateConfig(&alistConfigValue)

	// 执行任务 - 使用递归方法处理目录树
	if err := fileService.ProcessDirectoryRecursive(task.SourcePath); err != nil {
		c.logger.Error("执行任务失败",
			zap.String("taskId", taskID),
			zap.Error(err))
		// 标记任务失败
		taskLogService.MarkAsFailed(taskLog.ID, "执行任务失败: "+err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "执行任务失败: " + err.Error()})
		return
	}

	// 获取文件处理统计信息
	stats := fileService.GetStats()

	// 标记任务完成
	if err := taskLogService.MarkAsCompleted(
		taskLog.ID,
		stats.TotalFiles,
		stats.GeneratedFiles,
		stats.SkippedFiles,
		stats.MetadataFiles,
		stats.SubtitleFiles,
	); err != nil {
		c.logger.Error("更新任务日志失败", zap.Error(err))
		// 虽然任务执行成功，但日志更新失败，不影响返回结果
	}

	c.logger.Info("任务执行成功",
		zap.String("taskId", taskID),
		zap.Uint("taskLogId", taskLog.ID),
		zap.Int("totalFiles", stats.TotalFiles),
		zap.Int("generatedFiles", stats.GeneratedFiles))

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "任务执行成功",
		"taskId":    taskID,
		"taskLogId": taskLog.ID,
		"statistics": gin.H{
			"totalFiles":     stats.TotalFiles,
			"generatedFiles": stats.GeneratedFiles,
			"skippedFiles":   stats.SkippedFiles,
			"metadataFiles":  stats.MetadataFiles,
			"subtitleFiles":  stats.SubtitleFiles,
		},
	})
}
