package controller

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NotificationController struct {
	notificationService         *service.NotificationService
	embyNotificationService     *service.EmbyNotificationService
	telegramNotificationService *service.TelegramNotificationService
	logger                      *zap.Logger
}

// NewNotificationController 创建通知控制器
func NewNotificationController() *NotificationController {
	return &NotificationController{
		notificationService:         service.GetNotificationService(),
		embyNotificationService:     service.GetEmbyNotificationService(),
		telegramNotificationService: service.GetTelegramNotificationService(),
		logger:                      utils.Logger,
	}
}

// RegisterRoutes 注册路由
func (c *NotificationController) RegisterRoutes(router *gin.RouterGroup) {
	notificationGroup := router.Group("/notifications")
	{
		// 通知队列管理
		notificationGroup.GET("", c.ListNotificationQueue)
		notificationGroup.POST("", c.CreateNotificationQueue)
		notificationGroup.GET("/:id", c.GetNotificationQueue)
		notificationGroup.DELETE("/:id", c.DeleteNotificationQueue)

		// 通知处理
		notificationGroup.POST("/process-pending", c.ProcessPendingNotifications)
		notificationGroup.POST("/retry-failed", c.RetryFailedNotifications)
		notificationGroup.GET("/statistics", c.GetNotificationStatistics)

		// Emby 通知相关
		embyGroup := notificationGroup.Group("/emby")
		{
			embyGroup.GET("/config", c.GetEmbyConfig)
			embyGroup.PUT("/config", c.UpdateEmbyConfig)
			embyGroup.POST("/test", c.TestEmbyConnection)
		}

		// Telegram 通知相关
		telegramGroup := notificationGroup.Group("/telegram")
		{
			telegramGroup.GET("/config", c.GetTelegramConfig)
			telegramGroup.PUT("/config", c.UpdateTelegramConfig)
			telegramGroup.POST("/test", c.TestTelegramConnection)
		}
	}
}

// ListNotificationQueue 获取通知队列列表
func (c *NotificationController) ListNotificationQueue(ctx *gin.Context) {
	var req model.NotificationQueueQueryRequest
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

	data, total, err := c.notificationService.ListQueue(&req)
	if err != nil {
		c.logger.Error("获取通知队列列表失败", zap.Error(err))
		utils.ResponseError(ctx, "获取通知队列列表失败")
		return
	}

	utils.ResponsePage(ctx, data, total, req.Page, req.PageSize)
}

// CreateNotificationQueue 创建通知队列
func (c *NotificationController) CreateNotificationQueue(ctx *gin.Context) {
	var req model.NotificationQueueCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	data, err := c.notificationService.AddToQueue(&req)
	if err != nil {
		c.logger.Error("创建通知队列失败", zap.Error(err))
		utils.ResponseError(ctx, "创建通知队列失败")
		return
	}

	utils.ResponseSuccess(ctx, data)
}

// GetNotificationQueue 根据ID获取通知队列
func (c *NotificationController) GetNotificationQueue(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	data, err := c.notificationService.GetQueueByID(uint(id))
	if err != nil {
		c.logger.Error("获取通知队列失败", zap.Error(err))
		utils.ResponseError(ctx, "通知队列不存在")
		return
	}

	utils.ResponseSuccess(ctx, data)
}

// DeleteNotificationQueue 删除通知队列
func (c *NotificationController) DeleteNotificationQueue(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的ID")
		return
	}

	err = c.notificationService.DeleteFromQueue(uint(id))
	if err != nil {
		c.logger.Error("删除通知队列失败", zap.Error(err))
		utils.ResponseError(ctx, "删除通知队列失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "删除成功", nil)
}

// ProcessPendingNotifications 处理待处理的通知
func (c *NotificationController) ProcessPendingNotifications(ctx *gin.Context) {
	processed, err := c.notificationService.ProcessPendingNotifications()
	if err != nil {
		c.logger.Error("处理待处理通知失败", zap.Error(err))
		utils.ResponseError(ctx, "处理通知失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "处理完成", gin.H{"processedCount": processed})
}

// RetryFailedNotifications 重试失败的通知
func (c *NotificationController) RetryFailedNotifications(ctx *gin.Context) {
	retried, err := c.notificationService.RetryFailedNotifications()
	if err != nil {
		c.logger.Error("重试失败通知失败", zap.Error(err))
		utils.ResponseError(ctx, "重试通知失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "重试完成", gin.H{"retriedCount": retried})
}

// GetNotificationStatistics 获取通知统计信息
func (c *NotificationController) GetNotificationStatistics(ctx *gin.Context) {
	stats, err := c.notificationService.GetStatistics()
	if err != nil {
		c.logger.Error("获取通知统计信息失败", zap.Error(err))
		utils.ResponseError(ctx, "获取统计信息失败")
		return
	}

	utils.ResponseSuccess(ctx, stats)
}

// GetEmbyConfig 获取Emby配置
func (c *NotificationController) GetEmbyConfig(ctx *gin.Context) {
	config, err := c.embyNotificationService.GetConfig()
	if err != nil {
		c.logger.Error("获取Emby配置失败", zap.Error(err))
		utils.ResponseError(ctx, "获取Emby配置失败")
		return
	}

	utils.ResponseSuccess(ctx, config)
}

// UpdateEmbyConfig 更新Emby配置
func (c *NotificationController) UpdateEmbyConfig(ctx *gin.Context) {
	var config service.EmbyConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	err := c.embyNotificationService.UpdateConfig(&config)
	if err != nil {
		c.logger.Error("更新Emby配置失败", zap.Error(err))
		utils.ResponseError(ctx, "更新Emby配置失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "更新成功", nil)
}

// TestEmbyConnection 测试Emby连接
func (c *NotificationController) TestEmbyConnection(ctx *gin.Context) {
	err := c.embyNotificationService.TestConnection()
	if err != nil {
		c.logger.Error("Emby连接测试失败", zap.Error(err))
		utils.ResponseError(ctx, "连接测试失败: "+err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "连接测试成功", nil)
}

// GetTelegramConfig 获取Telegram配置
func (c *NotificationController) GetTelegramConfig(ctx *gin.Context) {
	config, err := c.telegramNotificationService.GetConfig()
	if err != nil {
		c.logger.Error("获取Telegram配置失败", zap.Error(err))
		utils.ResponseError(ctx, "获取Telegram配置失败")
		return
	}

	utils.ResponseSuccess(ctx, config)
}

// UpdateTelegramConfig 更新Telegram配置
func (c *NotificationController) UpdateTelegramConfig(ctx *gin.Context) {
	var config service.TelegramConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		utils.ResponseError(ctx, "参数错误")
		return
	}

	err := c.telegramNotificationService.UpdateConfig(&config)
	if err != nil {
		c.logger.Error("更新Telegram配置失败", zap.Error(err))
		utils.ResponseError(ctx, "更新Telegram配置失败")
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "更新成功", nil)
}

// TestTelegramConnection 测试Telegram连接
func (c *NotificationController) TestTelegramConnection(ctx *gin.Context) {
	err := c.telegramNotificationService.TestConnection()
	if err != nil {
		c.logger.Error("Telegram连接测试失败", zap.Error(err))
		utils.ResponseError(ctx, "连接测试失败: "+err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(ctx, "连接测试成功", nil)
}
