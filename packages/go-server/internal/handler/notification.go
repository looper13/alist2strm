package handler

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListNotificationQueueHandler 获取通知队列列表
func ListNotificationQueueHandler(c *gin.Context) {
	var req model.NotificationQueueQueryRequest
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

	notificationService := service.GetNotificationService()
	data, total, err := notificationService.ListQueue(&req)
	if err != nil {
		utils.Logger.Error("获取通知队列列表失败", zap.Error(err))
		utils.ResponseError(c, "获取通知队列列表失败")
		return
	}

	utils.ResponsePage(c, data, total, req.Page, req.PageSize)
}

// CreateNotificationQueueHandler 创建通知队列
func CreateNotificationQueueHandler(c *gin.Context) {
	var req model.NotificationQueueCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	notificationService := service.GetNotificationService()
	data, err := notificationService.AddToQueue(&req)
	if err != nil {
		utils.Logger.Error("创建通知队列失败", zap.Error(err))
		utils.ResponseError(c, "创建通知队列失败")
		return
	}

	utils.ResponseSuccess(c, data)
}

// GetNotificationQueueHandler 根据ID获取通知队列
func GetNotificationQueueHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	notificationService := service.GetNotificationService()
	data, err := notificationService.GetQueueByID(uint(id))
	if err != nil {
		utils.Logger.Error("获取通知队列失败", zap.Error(err))
		utils.ResponseError(c, "通知队列不存在")
		return
	}

	utils.ResponseSuccess(c, data)
}

// DeleteNotificationQueueHandler 删除通知队列
func DeleteNotificationQueueHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的ID")
		return
	}

	notificationService := service.GetNotificationService()
	err = notificationService.DeleteFromQueue(uint(id))
	if err != nil {
		utils.Logger.Error("删除通知队列失败", zap.Error(err))
		utils.ResponseError(c, "删除通知队列失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "删除成功", nil)
}

// ProcessPendingNotificationsHandler 处理待处理的通知
func ProcessPendingNotificationsHandler(c *gin.Context) {
	notificationService := service.GetNotificationService()
	processed, err := notificationService.ProcessPendingNotifications()
	if err != nil {
		utils.Logger.Error("处理待处理通知失败", zap.Error(err))
		utils.ResponseError(c, "处理通知失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "处理完成", gin.H{"processedCount": processed})
}

// RetryFailedNotificationsHandler 重试失败的通知
func RetryFailedNotificationsHandler(c *gin.Context) {
	notificationService := service.GetNotificationService()
	retried, err := notificationService.RetryFailedNotifications()
	if err != nil {
		utils.Logger.Error("重试失败通知失败", zap.Error(err))
		utils.ResponseError(c, "重试通知失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "重试完成", gin.H{"retriedCount": retried})
}

// GetNotificationStatisticsHandler 获取通知统计信息
func GetNotificationStatisticsHandler(c *gin.Context) {
	notificationService := service.GetNotificationService()
	stats, err := notificationService.GetStatistics()
	if err != nil {
		utils.Logger.Error("获取通知统计信息失败", zap.Error(err))
		utils.ResponseError(c, "获取统计信息失败")
		return
	}

	utils.ResponseSuccess(c, stats)
}

// GetEmbyConfigHandler 获取Emby配置
func GetEmbyConfigHandler(c *gin.Context) {
	embyService := service.GetEmbyNotificationService()
	config, err := embyService.GetConfig()
	if err != nil {
		utils.Logger.Error("获取Emby配置失败", zap.Error(err))
		utils.ResponseError(c, "获取Emby配置失败")
		return
	}

	utils.ResponseSuccess(c, config)
}

// UpdateEmbyConfigHandler 更新Emby配置
func UpdateEmbyConfigHandler(c *gin.Context) {
	var config service.EmbyConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	embyService := service.GetEmbyNotificationService()
	err := embyService.UpdateConfig(&config)
	if err != nil {
		utils.Logger.Error("更新Emby配置失败", zap.Error(err))
		utils.ResponseError(c, "更新Emby配置失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "更新成功", nil)
}

// TestEmbyConnectionHandler 测试Emby连接
func TestEmbyConnectionHandler(c *gin.Context) {
	embyService := service.GetEmbyNotificationService()
	err := embyService.TestConnection()
	if err != nil {
		utils.Logger.Error("Emby连接测试失败", zap.Error(err))
		utils.ResponseError(c, "连接测试失败: "+err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "连接测试成功", nil)
}

// GetTelegramConfigHandler 获取Telegram配置
func GetTelegramConfigHandler(c *gin.Context) {
	telegramService := service.GetTelegramNotificationService()
	config, err := telegramService.GetConfig()
	if err != nil {
		utils.Logger.Error("获取Telegram配置失败", zap.Error(err))
		utils.ResponseError(c, "获取Telegram配置失败")
		return
	}

	utils.ResponseSuccess(c, config)
}

// UpdateTelegramConfigHandler 更新Telegram配置
func UpdateTelegramConfigHandler(c *gin.Context) {
	var config service.TelegramConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	telegramService := service.GetTelegramNotificationService()
	err := telegramService.UpdateConfig(&config)
	if err != nil {
		utils.Logger.Error("更新Telegram配置失败", zap.Error(err))
		utils.ResponseError(c, "更新Telegram配置失败")
		return
	}

	utils.ResponseSuccessWithMessage(c, "更新成功", nil)
}

// TestTelegramConnectionHandler 测试Telegram连接
func TestTelegramConnectionHandler(c *gin.Context) {
	telegramService := service.GetTelegramNotificationService()
	err := telegramService.TestConnection()
	if err != nil {
		utils.Logger.Error("Telegram连接测试失败", zap.Error(err))
		utils.ResponseError(c, "连接测试失败: "+err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "连接测试成功", nil)
}
