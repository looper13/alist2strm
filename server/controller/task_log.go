package controller

import (
	"strconv"

	"github.com/MccRay-s/alist2strm/model/common/response"
	taskLogRequest "github.com/MccRay-s/alist2strm/model/tasklog/request"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/gin-gonic/gin"
)

type TaskLogController struct{}

// 包级别的全局实例
var TaskLogControllerInstance = &TaskLogController{}

// GetTaskLogInfo 获取任务日志信息
func (c *TaskLogController) GetTaskLogInfo(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID参数错误", ctx)
		return
	}

	req := &taskLogRequest.TaskLogInfoReq{
		GetById: struct {
			ID int `json:"id" form:"id"`
		}{ID: int(id)},
	}

	resp, err := service.TaskLogServiceInstance.GetTaskLogInfo(req)
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}

	response.SuccessWithData(resp, ctx)
}

// GetTaskLogList 获取任务日志列表
func (c *TaskLogController) GetTaskLogList(ctx *gin.Context) {
	var req taskLogRequest.TaskLogListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), ctx)
		return
	}

	resp, err := service.TaskLogServiceInstance.GetTaskLogList(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}

	response.SuccessWithData(resp, ctx)
}

// GetFileProcessingStats 获取文件处理统计数据
func (c *TaskLogController) GetFileProcessingStats(ctx *gin.Context) {
	// 获取时间范围参数，默认为day
	timeRange := ctx.DefaultQuery("timeRange", "day")

	// 参数验证
	if timeRange != "day" && timeRange != "month" && timeRange != "year" {
		response.FailWithMessage("无效的时间范围参数，支持的值为：day、month、year", ctx)
		return
	}

	stats, err := service.TaskLogServiceInstance.GetFileProcessingStats(timeRange)
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}

	response.SuccessWithData(stats, ctx)
}
