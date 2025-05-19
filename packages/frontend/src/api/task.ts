import type { HttpResponse } from '~/types'
import { http } from './http'

export class TaskAPI {
  private baseUrl = '/tasks'

  /**
   * 创建任务
   */
  async create(data: Api.TaskCreateDto): Promise<HttpResponse<Api.Task>> {
    return http.post<Api.Task>(this.baseUrl, data)
  }

  /**
   * 更新任务
   */
  async update(id: number, data: Api.TaskUpdateDto): Promise<HttpResponse<Api.Task>> {
    return http.put(`${this.baseUrl}/${id}`, data)
  }

  /**
   * 删除任务
   */
  async delete(id: number): Promise<HttpResponse<void>> {
    return http.delete(`${this.baseUrl}/${id}`)
  }

  /**
   * 获取所有任务
   */
  async findAll(query: { name?: string }): Promise<HttpResponse<Api.Task[]>> {
    return http.get<Api.Task[]>(`${this.baseUrl}/all`, { params: query })
  }

  /**
   * 执行任务
   */
  async execute(id: number): Promise<HttpResponse<void>> {
    return http.post(`${this.baseUrl}/${id}/execute`)
  }

  /**
   * 获取任务日志
   */
  findLogs(query: Api.TaskLogQuery) {
    return http.get<Api.PaginationResponse<Api.TaskLog>>(`${this.baseUrl}/${query.taskId}/logs`, { params: query })
  }
}

export const taskAPI = new TaskAPI()
