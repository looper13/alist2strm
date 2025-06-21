import { http } from './http'

export class TaskLogAPI {
  private baseUrl = '/task-log'

  /**
   * 获取任务日志
   */
  findLogs(query: Api.Task.LogQuery) {
    return http.get<Api.Common.PaginationResponse<Api.Task.Log>>(`${this.baseUrl}`, { params: query })
  }

  /**
   * 获取文件处理统计数据
   * @param timeRange 时间范围：day-日, month-月, year-年
   */
  getFileProcessingStats(timeRange: string = 'day') {
    return http.get<Api.TaskLog.FileProcessingStats>(`${this.baseUrl}/stats/processing`, {
      params: { timeRange },
    })
  }
}

export const taskLogAPI = new TaskLogAPI()
