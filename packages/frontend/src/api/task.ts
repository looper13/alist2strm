import type { HttpResponse, PaginationQuery, PaginationResponse } from '~/types'
import { http } from './http'

export class TaskAPI {
  private baseUrl = '/tasks'

  async create(data: Api.TaskCreateDto): Promise<HttpResponse<Api.Task>> {
    return http.post(this.baseUrl, data)
  }

  async update(id: number, data: Api.TaskUpdateDto): Promise<HttpResponse<Api.Task>> {
    return http.put(`${this.baseUrl}/${id}`, data)
  }

  async delete(id: number): Promise<HttpResponse<void>> {
    return http.delete(`${this.baseUrl}/${id}`)
  }

  async findByPage(query: PaginationQuery & { keyword?: string }): Promise<HttpResponse<PaginationResponse<Api.Task>>> {
    return http.get(this.baseUrl, { params: query })
  }

  async findAll(query: { name?: string }): Promise<HttpResponse<Api.Task[]>> {
    return http.get(`${this.baseUrl}/all`, { params: query })
  }

  async execute(id: number): Promise<HttpResponse<void>> {
    return http.post(`${this.baseUrl}/${id}/execute`)
  }

  /**
   * 获取任务日志
   */
  findLogs(taskId: number) {
    return http.get<HttpResponse<Api.TaskLog[]>>(`${this.baseUrl}/${taskId}/logs`)
  }
}

export const taskAPI = new TaskAPI()
