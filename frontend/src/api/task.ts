import { http } from './http'

export class TaskAPI {
  private baseUrl = '/task'

  /**
   * 创建任务
   */
  async create(data: Api.Task.Create) {
    return http.post(this.baseUrl, data)
  }

  /**
   * 更新任务
   */
  async update(id: number, data: Api.Task.Update) {
    return http.put(`${this.baseUrl}/${id}`, data)
  }

  /**
   * 删除任务
   */
  async delete(id: number) {
    return http.delete(`${this.baseUrl}/${id}`)
  }

  /**
   * 获取所有任务
   */
  async findAll(query: { name?: string }) {
    return http.get<Api.Task.Record[]>(`${this.baseUrl}/all`, { params: query })
  }

  /**
   * 执行任务
   */
  async execute(id: number) {
    return http.post(`${this.baseUrl}/${id}/execute?async=true`)
  }

  /**
   * 获取任务日志
   */
  findLogs(query: Api.Task.LogQuery) {
    return http.get<Api.Common.PaginationResponse<Api.Task.Log>>(`${this.baseUrl}/${query.taskId}/logs`, { params: query })
  }

  /**
   * 重置任务状态
   */
  async resetStatus(id: number) {
    return http.post(`${this.baseUrl}/${id}/reset-status`)
  }
}

export const taskAPI = new TaskAPI()
