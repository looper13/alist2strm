package controller

import (
	"strconv"

	"github.com/MccRay-s/alist2strm/model/common/response"
	taskRequest "github.com/MccRay-s/alist2strm/model/task/request"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/MccRay-s/alist2strm/utils"
	"github.com/gin-gonic/gin"
)

// 包级别的任务控制器实例
var Task = &TaskController{}

type TaskController struct{}

// GetTaskStats 获取任务统计数据
func (tc *TaskController) GetTaskStats(c *gin.Context) {
	// 获取时间范围参数，默认为day
	timeRange := c.DefaultQuery("timeRange", "day")

	// 参数验证
	if timeRange != "day" && timeRange != "month" && timeRange != "year" {
		response.FailWithMessage("无效的时间范围参数，支持的值为：day、month、year", c)
		return
	}

	stats, err := service.Task.GetTaskStats(timeRange)
	if err != nil {
		utils.Error("获取任务统计数据失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取任务统计数据成功", "timeRange", timeRange, "request_id", c.GetString("request_id"))
	response.SuccessWithData(stats, c)
}

// Create 创建任务
func (tc *TaskController) Create(c *gin.Context) {
	var req taskRequest.TaskCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("创建任务参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	err := service.Task.Create(&req)
	if err != nil {
		utils.Error("创建任务失败", "name", req.Name, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("创建任务成功", "name", req.Name, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("创建成功", c)
}

// GetTaskInfo 获取任务信息
func (tc *TaskController) GetTaskInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("获取任务信息ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("任务ID参数错误", c)
		return
	}

	req := &taskRequest.TaskInfoReq{}
	req.ID = id

	taskInfo, err := service.Task.GetTaskInfo(req)
	if err != nil {
		utils.Error("获取任务信息失败", "task_id", id, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取任务信息成功", "task_id", id, "request_id", c.GetString("request_id"))
	response.SuccessWithData(taskInfo, c)
}

// UpdateTask 更新任务
func (tc *TaskController) UpdateTask(c *gin.Context) {
	// 从路径参数获取任务ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("更新任务ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("任务ID参数错误", c)
		return
	}

	var req taskRequest.TaskUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("更新任务参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 设置任务ID
	req.ID = uint(id)

	err = service.Task.UpdateTask(&req)
	if err != nil {
		utils.Error("更新任务失败", "task_id", req.ID, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("更新任务成功", "task_id", req.ID, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("更新成功", c)
}

// DeleteTask 删除任务
func (tc *TaskController) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("删除任务ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("任务ID参数错误", c)
		return
	}

	err = service.Task.DeleteTask(uint(id))
	if err != nil {
		utils.Error("删除任务失败", "task_id", id, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("删除任务成功", "task_id", id, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("删除成功", c)
}

// GetTaskList 获取任务列表（分页）
func (tc *TaskController) GetTaskList(c *gin.Context) {
	var req taskRequest.TaskListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error("获取任务列表参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	taskList, err := service.Task.GetTaskList(&req)
	if err != nil {
		utils.Error("获取任务列表失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取任务列表成功", "total", taskList.Total, "page", taskList.Page, "request_id", c.GetString("request_id"))
	response.SuccessWithData(taskList, c)
}

// GetAllTasks 获取所有任务（不分页）
func (tc *TaskController) GetAllTasks(c *gin.Context) {
	var req taskRequest.TaskAllReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error("获取所有任务参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	tasks, err := service.Task.GetAllTasks(&req)
	if err != nil {
		utils.Error("获取所有任务失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取所有任务成功", "total", len(tasks), "request_id", c.GetString("request_id"))
	response.SuccessWithData(tasks, c)
}

// ToggleTaskEnabled 切换任务启用状态
func (tc *TaskController) ToggleTaskEnabled(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("切换任务启用状态ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("任务ID参数错误", c)
		return
	}

	err = service.Task.ToggleTaskEnabled(uint(id))
	if err != nil {
		utils.Error("切换任务启用状态失败", "task_id", id, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("切换任务启用状态成功", "task_id", id, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("任务状态切换成功", c)
}

// ResetTaskStatus 重置任务运行状态
func (tc *TaskController) ResetTaskStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("重置任务状态ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("任务ID参数错误", c)
		return
	}

	err = service.Task.ResetTaskStatus(uint(id))
	if err != nil {
		utils.Error("重置任务状态失败", "task_id", id, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("重置任务状态成功", "task_id", id, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("重置任务状态成功", c)
}

// ExecuteTask 执行任务
func (tc *TaskController) ExecuteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("执行任务ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("任务ID参数错误", c)
		return
	}

	// 获取是否异步执行参数（兼容两种方式）
	var async bool

	// 先尝试从 JSON 请求体获取
	var req taskRequest.TaskExecuteReq
	if err := c.ShouldBindJSON(&req); err == nil {
		async = !req.Sync // JSON 中是 sync 字段，需要取反
	} else {
		// 如果 JSON 解析失败，尝试从查询参数获取
		async = c.DefaultQuery("async", "false") == "true"
	}

	var err2 error
	if async {
		// 异步执行
		err2 = service.Task.ExecuteStrmGenerationAsync(uint(id))
		if err2 == nil {
			utils.Info("异步执行任务已启动", "task_id", id, "request_id", c.GetString("request_id"))
			response.SuccessWithMessage("任务已启动", c)
		}
	} else {
		// 同步执行
		execResult, err2 := service.Task.ExecuteStrmGeneration(uint(id))
		if err2 != nil {
			utils.Error("同步执行任务失败", "task_id", id, "error", err2.Error(), "request_id", c.GetString("request_id"))

			// 即使发生错误，也返回执行结果（如果有）
			if execResult != nil {
				execResult.Status = "failed"
				execResult.ErrorMessage = err2.Error()
				response.FailWithDetailed(execResult, err2.Error(), c)
			} else {
				response.FailWithMessage(err2.Error(), c)
			}
		} else {
			utils.Info("同步执行任务成功", "task_id", id,
				"成功", execResult.SuccessCount,
				"失败", execResult.FailedCount,
				"跳过", execResult.SkippedCount,
				"耗时", execResult.Duration,
				"request_id", c.GetString("request_id"))
			response.SuccessWithData(execResult, c)
		}
	}

	// 只需要处理异步执行的错误情况，因为同步执行已经在上面处理了
	if async && err2 != nil {
		utils.Error("启动异步任务失败", "task_id", id, "error", err2.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err2.Error(), c)
		return
	}
}
