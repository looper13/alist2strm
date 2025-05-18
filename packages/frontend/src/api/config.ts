import type { HttpResponse, PaginationQuery, PaginationResponse } from '~/types'
import { http } from './http'

export class ConfigAPI {
  private baseUrl = '/configs'

  /**
   * 创建配置
   */
  async create(data: Api.ConfigCreateDto): Promise<HttpResponse<Api.Config>> {
    return http.post(this.baseUrl, data)
  }

  /**
   * 更新配置
   */
  async update(id: number, data: Api.ConfigUpdateDto): Promise<HttpResponse<Api.Config>> {
    return http.put(`${this.baseUrl}/${id}`, data)
  }

  /**
   * 删除配置
   */
  async delete(id: number): Promise<HttpResponse<void>> {
    return http.delete(`${this.baseUrl}/${id}`)
  }

  /**
   * 分页查询配置
   */
  async findByPage(query: PaginationQuery & { keyword?: string }): Promise<HttpResponse<PaginationResponse<Api.Config>>> {
    return http.get(this.baseUrl, { params: query })
  }

  /**
   * 查询所有配置
   */
  async findAll(): Promise<HttpResponse<Api.Config[]>> {
    return http.get(`${this.baseUrl}/all`)
  }

  /**
   * 根据ID查询配置
   */
  async findById(id: number): Promise<HttpResponse<Api.Config>> {
    return http.get(`${this.baseUrl}/${id}`)
  }
}

export const configAPI = new ConfigAPI()
