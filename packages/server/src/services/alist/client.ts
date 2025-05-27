import axios, { AxiosInstance } from 'axios'
import type { AlistConfig } from '@/types/alist.js'
import { logger } from '@/utils/logger.js'

export class AlistClient {
  private client: AxiosInstance
  private config: AlistConfig

  constructor(config: AlistConfig) {
    this.config = config
    this.client = this.createClient(config)
  }

  private createClient(config: AlistConfig): AxiosInstance {
    return axios.create({
      baseURL: config.host,
      headers: {
        Authorization: config.token,
      },
    })
  }

  updateConfig(config: AlistConfig): void {
    this.config = config
    this.client = this.createClient(config)
  }

  async listFiles(path: string, page: number, perPage: number): Promise<App.AList.AlistListResponse<App.AList.AlistFile[]>> {
    try {
      const response = await this.client.post<App.AList.AlistListResponse<App.AList.AlistFile[]>>('/api/fs/list', {
        path,
        password: '',
        page,
        per_page: perPage,
        refresh: false,
      })

      if (response.data.code !== 200) {
        throw new Error(response.data.message || 'AList API 返回非200状态码')
      }

      return response.data
    }
    catch (error) {
      logger.error.error('获取文件列表失败', {
        path,
        page,
        error: error instanceof Error ? error.message : String(error),
      })
      throw error
    }
  }

  async getFileInfo(path: string): Promise<App.AList.AlistGetResponse<App.AList.AlistFile>> {
    try {
      const response = await this.client.post<App.AList.AlistGetResponse<App.AList.AlistFile>>('/api/fs/get', {
        path,
        password: '',
      })

      if (response.data.code !== 200) {
        throw new Error(response.data.message || 'AList API 返回非200状态码')
      }

      return response.data
    }
    catch (error) {
      logger.error.error('获取文件信息失败', {
        path,
        error: error instanceof Error ? error.message : String(error),
      })
      throw error
    }
  }
} 