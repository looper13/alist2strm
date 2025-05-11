import axios from 'axios'
import type { AlistStorage } from '../types'
import config from '../config'
import { alistLogger as logger, errorLogger } from '../utils/logger'

interface AlistListResponse<T> {
  code: number
  message: string
  data: {
    content: T
  }
}

interface AlistGetResponse<T> {
  code: number
  message: string
  data: T
}

interface AlistFile {
  name: string
  size?: number
  is_dir: boolean
  modified?: string
  created?: string
  sign?: string
  thumb?: string
  type?: number
  hashinfo?: string
  hash_info?: any
  raw_url?: string
  readme?: string
  header?: string
  provider?: string
  related?: any
}

class AlistService {
  private client
  private readonly maxRetries = 3
  private readonly retryDelay = 1000 // 1秒

  constructor() {
    this.client = axios.create({
      baseURL: config.alist.host,
      headers: config.alist.token
        ? {
            Authorization: config.alist.token,
          }
        : undefined,
    })

    // 添加响应拦截器
    this.client.interceptors.response.use(
      (response) => {
        logger.debug('Response:', {
          url: response.config.url,
          method: response.config.method,
          data: response.config.data,
          status: response.status,
          responseData: response.data,
        })

        if (response.data?.code !== 200) {
          throw new Error(response.data?.message || 'AList API returned non-200 code')
        }
        return response
      },
      (error) => {
        errorLogger.error('Request failed:', {
          url: error.config?.url,
          method: error.config?.method,
          data: error.config?.data,
          error: error.message,
          response: error.response?.data,
        })

        if (axios.isAxiosError(error)) {
          const message = error.response?.data?.message || error.message
          throw new Error(`AList API error: ${message}`)
        }
        throw error
      },
    )
  }

  /**
   * 规范化路径字符串
   * @param path 原始路径
   * @returns 规范化后的路径
   */
  private normalizePath(path: string): string {
    // 移除多余的斜杠
    const normalized = path.replace(/\/+/g, '/')
    // 确保以斜杠开头
    return normalized.startsWith('/') ? normalized : `/${normalized}`
  }

  /**
   * 带重试机制的 API 调用
   */
  private async retryableRequest<T>(operation: () => Promise<T>, retryCount = 0): Promise<T> {
    try {
      return await operation()
    } catch (error) {
      if (retryCount >= this.maxRetries) {
        throw error
      }

      logger.warn(`Retrying operation (attempt ${retryCount + 1}/${this.maxRetries})...`)
      await new Promise((resolve) => setTimeout(resolve, this.retryDelay * (retryCount + 1)))
      return this.retryableRequest(operation, retryCount + 1)
    }
  }

  async listStorages(): Promise<AlistStorage[]> {
    try {
      const response =
        await this.client.get<AlistListResponse<AlistStorage[]>>('/api/admin/storage/list')
      if (!response.data?.data?.content) {
        throw new Error('No storage list returned from AList API')
      }
      return response.data.data.content
    } catch (error) {
      errorLogger.error('Failed to list storages:', error)
      throw error
    }
  }

  async listFiles(path: string): Promise<AlistFile[]> {
    logger.info('Listing files for path:', path)
    try {
      const response = await this.retryableRequest(async () => {
        const resp = await this.client.post<AlistListResponse<AlistFile[]>>('/api/fs/list', {
          path: path,
          password: '',
          page: 1,
          per_page: 0,
          refresh: false,
        })

        if (!resp.data?.data?.content) {
          throw new Error('No file list returned from AList API')
        }
        return resp
      })

      logger.debug(`Found ${response.data.data.content.length} files at path: ${path}`)
      return response.data.data.content
    } catch (error) {
      errorLogger.error(`Failed to list files for path ${path}:`, error)
      throw error
    }
  }

  async getFileInfo(path: string): Promise<AlistFile> {
    logger.info('Getting file info for path:', path)
    try {
      const response = await this.client.post<AlistGetResponse<AlistFile>>('/api/fs/get', {
        path: path,
        password: '',
        page: 1,
        per_page: 0,
        refresh: false,
      })

      if (!response.data?.data) {
        throw new Error('No file info returned from AList API')
      }

      logger.debug('File info retrieved:', response.data.data)
      return response.data.data
    } catch (error) {
      errorLogger.error(`Failed to get file info for path ${path}:`, error)
      throw error
    }
  }

  /**
   * 根据给定的路径查找对应的存储。
   *
   * @param path 要查找的存储路径。
   * @returns 返回与给定路径匹配的存储对象，如果没有找到则返回 undefined。
   * @throws 如果没有找到有效的存储或路径未找到匹配的存储，则抛出错误。
   */
  async findStorage(path: string): Promise<AlistStorage | undefined> {
    logger.info('Finding storage for path:', path)
    try {
      const storages = await this.listStorages()
      if (!Array.isArray(storages) || storages.length === 0) {
        throw new Error('No valid storage found in AList')
      }

      // 确保路径以斜杠开头进行匹配
      const normalizedPath = this.normalizePath(path)
      const storage = storages.find((s) => {
        const mountPath = this.normalizePath(s.mount_path)
        return normalizedPath.startsWith(mountPath)
      })

      if (!storage) {
        throw new Error(`No storage found for path: ${path}`)
      }

      logger.debug('Found storage:', storage)
      return storage
    } catch (error) {
      errorLogger.error(`Failed to find storage for path ${path}:`, error)
      throw error
    }
  }
}

export default new AlistService()
