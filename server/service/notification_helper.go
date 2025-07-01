package service

import (
	"fmt"

	"github.com/MccRay-s/alist2strm/model/task"
)

// SendTaskNotification 发送任务通知
func (s *StrmGeneratorService) sendNotification(taskInfo *task.Task, taskLogID uint, status string, duration int64, stats map[string]interface{}) error {
	// 使用通知服务发送通知
	notificationService := GetNotificationService()
	if notificationService == nil {
		s.logger.Warn("通知服务不可用，无法发送通知")
		return fmt.Errorf("通知服务不可用")
	}

	return notificationService.SendTaskNotification(taskInfo, taskLogID, status, duration, stats)
}
