package handler

import (
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ExecuteTaskHandler 执行指定任务
func ExecuteTaskHandler(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.Error("任务ID不能为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
		return
	}

	// 获取任务信息
	task := &model.Task{}
	if err := model.DB.First(task, taskID).Error; err != nil {
		utils.Error("获取任务信息失败", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 获取 Alist 配置
	alistConfig := &model.Config{}
	if err := model.DB.Where("code = ?", "ALIST").First(alistConfig).Error; err != nil {
		utils.Error("获取Alist配置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到Alist配置"})
		return
	}

	// 解析 Alist 配置
	var alistConfigValue service.AlistConfig
	if err := json.Unmarshal([]byte(alistConfig.Value), &alistConfigValue); err != nil {
		utils.Error("解析Alist配置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Alist配置格式错误"})
		return
	}

	// 创建文件服务实例
	fileService := service.NewFileService(utils.Logger)

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
	alistClient := service.GetAlistClient(utils.Logger)
	alistClient.UpdateConfig(&alistConfigValue)

	// 执行任务
	if err := fileService.ProcessDirectory(task.SourcePath); err != nil {
		utils.Error("执行任务失败",
			zap.String("taskId", taskID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "执行任务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务执行成功",
		"taskId":  taskID,
	})
}
