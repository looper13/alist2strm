package handler

import (
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateTask 创建任务
func CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("解析创建任务请求失败", zap.Error(err))
		utils.ResponseError(c, err.Error())
		return
	}

	task, err := service.GetTaskService().CreateTask(&req)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, task)
}

// UpdateTask 更新任务
func UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的任务ID")
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("解析更新任务请求失败", zap.Error(err))
		utils.ResponseError(c, err.Error())
		return
	}

	task, err := service.GetTaskService().UpdateTask(uint(id), &req)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, task)
}

// DeleteTask 删除任务
func DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的任务ID")
		return
	}

	if err := service.GetTaskService().DeleteTask(uint(id)); err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "任务已删除", nil)
}

// GetTask 获取任务详情
func GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的任务ID")
		return
	}

	task, err := service.GetTaskService().GetTaskByID(uint(id))
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, task)
}

// ListTasks 获取任务列表
func ListTasks(c *gin.Context) {
	name := c.Query("name")

	tasks, err := service.GetTaskService().ListTasks(name)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, tasks)
}

// SetTaskStatus 设置任务状态
func SetTaskStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的任务ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	if err := service.GetTaskService().SetTaskStatus(uint(id), req.Enabled); err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "任务状态已更新", nil)
}

// ResetTaskStatus 重置任务运行状态
func ResetTaskStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的任务ID")
		return
	}

	if err := service.GetTaskService().ResetRunningStatus(uint(id)); err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "任务运行状态已重置", nil)
}
