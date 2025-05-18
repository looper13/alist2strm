import config from '@/config'
import { createHttpClient, HttpClient } from '@/utils/http'
import type { AlistFile, AlistMediaFile, AlistDir } from '@/types'
import { Config } from '@/models/config'

class AlistService {
  private http: HttpClient | undefined
  private initialized = false

  constructor() {
    this._initializeHttp()
  }

  private async _initializeHttp() {
    if (this.initialized)
      return

    try {
      const [hostConfig, tokenConfig] = await Promise.all([
        Config.findOne({ where: { code: 'ALIST_HOST' } }),
        Config.findOne({ where: { code: 'ALIST_TOKEN' } }),
      ])

      if (!hostConfig?.value || !tokenConfig?.value)
        throw new Error('AList 配置未找到，请先配置')

      this.http = createHttpClient(hostConfig.value, {
        headers: {
          Authorization: tokenConfig.value,
        },
      })

      this.initialized = true
    }
    catch (error) {
      console.error('AList 初始化失败:', error)
      throw error
    }
  }

  // 获取文件列表
  async listFiles(path: string): Promise<AlistFile[]> {
    const response = await this.http!.post<{ content: AlistFile[] }>('/api/fs/list', {
      path,
      password: '',
      page: 1,
      per_page: 0,
      refresh: false,
    })
    return response.data?.content || []
  }

  // 获取目录列表
  async listDirs(path: string): Promise<AlistDir[]> {
    const response = await this.http!.post<{ data: AlistDir[] }>('/api/fs/dirs', {
      path,
      password: '',
      force_root: false,
    })
    return response.data?.data || []
  }

  // 获取文件信息
  async getFileInfo(path: string): Promise<AlistFile> {
    const response = await this.http!.post<AlistFile>('/api/fs/get', {
      path,
      password: '',
    })
    if (!response.data) {
      throw new Error('Failed to get file info')
    }
    return response.data
  }

  // 获取媒体文件信息
  async getMediaInfo(path: string): Promise<AlistMediaFile> {
    const response = await this.http!.post<AlistMediaFile>('/api/fs/get', {
      path,
      password: '',
    })
    if (!response.data) {
      throw new Error('Failed to get media info')
    }
    return response.data
  }

  // 获取下载链接
  async getDownloadUrl(path: string): Promise<string> {
    interface LinkResponse {
      code: number
      message: string
      data: {
        raw_url: string
      }
    }
    
    const response = await this.http!.post<LinkResponse>('/api/fs/link', {
      path,
      password: '',
    })
    if (!response.data?.data?.raw_url) {
      throw new Error('Failed to get download URL')
    }
    return response.data.data.raw_url
  }
}

export const alistService = new AlistService() 