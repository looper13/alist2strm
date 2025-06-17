import { http } from './http'

export class TaskLogAPI {
  private baseUrl = '/task-log'

  /**
   * 获取任务日志
   */
  findLogs(query: Api.Task.LogQuery) {
    return http.get<Api.Common.PaginationResponse<Api.Task.Log>>(`${this.baseUrl}`, { params: query })
  }
}

export const taskLogAPI = new TaskLogAPI()
