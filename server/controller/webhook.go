package controller

import (
	"log"
	"net/http"
	"strings"

	taskRequest "github.com/MccRay-s/alist2strm/model/task/request"
	"github.com/MccRay-s/alist2strm/model/webhook"
	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/gin-gonic/gin"
)

// 包级别的Webhook控制器实例
var Webhook = &WebhookController{}

type WebhookController struct{}

// FileNotifyHandler 处理来自 file_system_watcher 的 webhook
func (*WebhookController) FileNotifyHandler(c *gin.Context) {
	var payload webhook.FileWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("[FileNotify] 解析 JSON 失败: %v", err)
		// 返回一个包含错误信息的 JSON 响应
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体", "details": err.Error()})
		return
	}

	list, err := service.Task.GetTaskList(&taskRequest.TaskListReq{ConfigType: "clouddrive"})
	if err != nil {
		log.Printf("获取clouddrive列表失败: %v", err)
		return
	}
	tasks := list.List

	for _, event := range payload.Data {
		sourceFile := event.SourceFile
		log.Printf("  -> Action: %s, IsDir: %v, Source: %s, Destination: %s",
			event.Action, event.IsDir, sourceFile, event.DestinationFile)

		// 异步处理，避免阻塞 webhook 响应
		go func(e webhook.FileChangeEvent) {
			switch e.Action {
			case "create":
				for _, taskInfo := range tasks {
					// 检查创建的文件/目录是否在任务的源路径下
					if strings.HasPrefix(e.SourceFile, taskInfo.SourcePath) {
						log.Printf("  -> File '%s' matches task '%s'. Triggering processing.", e.SourceFile, taskInfo.Name)

						// 从数据库获取完整的任务对象
						fullTask, err := repository.Task.GetByID(taskInfo.ID)
						if err != nil {
							log.Printf("  -> Error fetching full task details for task ID %d: %v", taskInfo.ID, err)
							continue // 继续处理下一个匹配的任务
						}

						err = service.GetStrmGeneratorService().ProcessFileChangeEvent(fullTask, &e)
						if err != nil {
							log.Printf("  -> Error processing event for '%s': %v", e.SourceFile, err)
						}
					}
				}
			case "delete":
				for _, taskInfo := range tasks {
					if strings.HasPrefix(e.SourceFile, taskInfo.SourcePath) {
						log.Printf("  -> File '%s' matches task '%s'. Triggering delete.", e.SourceFile, taskInfo.Name)
						fullTask, err := repository.Task.GetByID(taskInfo.ID)
						if err != nil {
							log.Printf("  -> Error fetching full task details for task ID %d: %v", taskInfo.ID, err)
							continue
						}
						if err := service.GetStrmGeneratorService().ProcessFileDeleteEvent(fullTask, e.SourceFile); err != nil {
							log.Printf("  -> Error processing delete event for '%s': %v", e.SourceFile, err)
						}
					}
				}
			case "rename", "move":
				for _, taskInfo := range tasks {
					// For rename/move, either the source or destination could be in the task path.
					if strings.HasPrefix(e.SourceFile, taskInfo.SourcePath) || strings.HasPrefix(e.DestinationFile, taskInfo.SourcePath) {
						log.Printf("  -> File move/rename '%s'->'%s' matches task '%s'. Triggering processing.", e.SourceFile, e.DestinationFile, taskInfo.Name)
						fullTask, err := repository.Task.GetByID(taskInfo.ID)
						if err != nil {
							log.Printf("  -> Error fetching full task details for task ID %d: %v", taskInfo.ID, err)
							continue
						}
						if err := service.GetStrmGeneratorService().ProcessFileRenameEvent(fullTask, &e); err != nil {
							log.Printf("  -> Error processing rename/move event for '%s': %v", e.SourceFile, err)
						}
					}
				}
			}
		}(event)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File notification received and is being processed.",
	})
}

// MountNotifyHandler 处理来自 mount_point_watcher 的 webhook
func (*WebhookController) MountNotifyHandler(c *gin.Context) {
	// 1. 从 URL 查询参数中获取信息
	deviceName := c.Query("device_name")
	userName := c.Query("user_name")
	eventType := c.Query("type")
	log.Printf("[MountNotify] 收到来自设备 '%s' (用户: %s) 的请求, 类型: %s", deviceName, userName, eventType)

	// 2. 解析 JSON 请求体
	var payload webhook.MountWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("[MountNotify] 解析 JSON 失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体", "details": err.Error()})
		return
	}

	// 3. 处理解析后的数据
	log.Printf("[MountNotify] 成功解析负载 (版本: %s, 事件: %s/%s)",
		payload.Version, payload.EventCategory, payload.EventName)

	for _, event := range payload.Data {
		log.Printf("  -> Action: %s, MountPoint: %s, Status: %v, Reason: '%s'",
			event.Action, event.MountPoint, event.Status, event.Reason)
	}

	// 4. 发送成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Mount notification received.",
	})
}
