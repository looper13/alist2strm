package response

import (
	"time"

	"github.com/MccRay-s/alist2strm/model/tasklog"
)

// TaskLogInfoResp 任务日志信息响应
type TaskLogInfoResp struct {
	tasklog.TaskLog
}

// TaskLogListResp 任务日志列表响应
type TaskLogListResp struct {
	List     []tasklog.TaskLog `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// TaskLogCreateResp 任务日志创建响应
type TaskLogCreateResp struct {
	ID        uint      `json:"id"`
	TaskID    uint      `json:"taskId"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"startTime"`
	Message   string    `json:"message"`
}
