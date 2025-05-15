import axios from 'axios'
import type { AlistStorage } from '../types'
import config from '../config'
import { logger } from '../utils/logger'

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

/**
 * AList 服务
 */
class AlistService {
  private client
  private readonly maxRetries = 3
  private readonly retryDelay = 3000 // 3秒

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
        logger.debug.debug('AList API 响应', {
          url: response.config.url,
          method: response.config.method,
          data: response.config.data,
          status: response.status,
          responseData: response.data,
        })

        if (response.data?.code !== 200) {
          throw new Error(response.data?.message || 'AList API 返回非200状态码')
        }
        return response
      },
      (error) => {
        logger.error.error('AList API 请求失败', {
          url: error.config?.url,
          method: error.config?.method,
          data: error.config?.data,
          error: error.message,
          response: error.response?.data,
          stack: error.stack,
        })

        if (axios.isAxiosError(error)) {
          const message = error.response?.data?.message || error.message
          throw new Error(`AList API 错误: ${message}`)
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
        logger.error.error('达到最大重试次数', {
          retryCount,
          maxRetries: this.maxRetries,
          error: (error as Error).message,
        })
        throw error
      }

      logger.warn.warn('正在重试操作', {
        attempt: retryCount + 1,
        maxRetries: this.maxRetries,
        delay: this.retryDelay * (retryCount + 1),
      })

      await new Promise((resolve) => setTimeout(resolve, this.retryDelay * (retryCount + 1)))
      return this.retryableRequest(operation, retryCount + 1)
    }
  }

  /**
   * 列出存储
   * @returns 存储列表
   */
  async listStorages(): Promise<AlistStorage[]> {
    logger.info.info('正在获取存储列表')
    try {
      const response =
        await this.client.get<AlistListResponse<AlistStorage[]>>('/api/admin/storage/list')
      if (!response.data?.data?.content) {
        throw new Error('AList API 未返回存储列表')
      }

      logger.debug.debug('成功获取存储列表', {
        count: response.data.data.content.length,
      })

      return response.data.data.content
    } catch (error) {
      logger.error.error('获取存储列表失败', {
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
      throw error
    }
  }

  /**
   * 列出文件/目录
   * @param path 文件路径
   * @returns 文件列表
   */
  async listFiles(path: string): Promise<AlistFile[]> {
    logger.info.info('正在列出文件', { path })
    try {
      const perPage = 100 // 每页数量
      let page = 1
      let allFiles: AlistFile[] = []
      let hasMore = true
      let requestCount = 0
      let consecutiveErrors = 0
      const maxConsecutiveErrors = 3

      while (hasMore) {
        try {
          // 每3次请求后添加8秒间隔
          if (requestCount > 0 && requestCount % 3 === 0) {
            const waitTime = 8000 + consecutiveErrors * 2000 // 每次连续错误增加2秒等待
            logger.info.info('达到请求限制，等待中', { requestCount, waitTime })
            await new Promise((resolve) => setTimeout(resolve, waitTime))
          }

          // 每次请求前添加基础延迟
          const baseDelay = 3000 + consecutiveErrors * 1000 // 基础延迟3秒，每次错误增加1秒
          await new Promise((resolve) => setTimeout(resolve, baseDelay))

          const response = await this.retryableRequest(async () => {
            const resp = await this.client.post<AlistListResponse<AlistFile[]>>('/api/fs/list', {
              path: path,
              password: '',
              page: page,
              per_page: perPage,
              refresh: false,
            })

            if (!resp.data?.data?.content) {
              throw new Error('AList API 未返回文件列表')
            }
            return resp
          })

          const files = response.data.data.content
          allFiles = allFiles.concat(files)
          requestCount++
          consecutiveErrors = 0 // 重置连续错误计数

          logger.debug.debug('成功获取文件列表分页数据', {
            path,
            page,
            requestCount,
            currentPageCount: files.length,
            totalCount: allFiles.length,
            consecutiveErrors,
          })

          // 如果当前页的文件数小于每页限制，说明已经是最后一页
          hasMore = files.length === perPage
          page++

          // 每次请求后添加基础延迟
          await new Promise((resolve) => setTimeout(resolve, 2000))
        } catch (error) {
          consecutiveErrors++
          logger.warn.warn('获取文件列表失败，准备重试', {
            path,
            page,
            consecutiveErrors,
            error: (error as Error).message,
          })

          if (consecutiveErrors >= maxConsecutiveErrors) {
            logger.error.error('连续失败次数过多，中止操作', {
              path,
              consecutiveErrors,
              error: (error as Error).message,
            })
            throw error
          }

          // 发生错误后等待更长时间
          const errorWaitTime = 10000 * consecutiveErrors // 错误后等待10秒，每次错误增加10秒
          await new Promise((resolve) => setTimeout(resolve, errorWaitTime))
          continue // 重试当前页
        }
      }

      logger.debug.debug('完成获取所有文件列表', {
        path,
        totalFileCount: allFiles.length,
        totalRequests: requestCount,
        totalErrors: consecutiveErrors,
      })

      return allFiles
    } catch (error) {
      logger.error.error('获取文件列表失败', {
        path,
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
      throw error
    }
  }

  /**
   * 获取文件信息
   * @param path 文件路径
   * @returns 文件信息
   */
  async getFileInfo(path: string): Promise<AlistFile> {
    logger.info.info('正在获取文件信息', { path })
    try {
      const response = await this.client.post<AlistGetResponse<AlistFile>>('/api/fs/get', {
        path: path,
        password: '',
        page: 1,
        per_page: 0,
        refresh: false,
      })

      if (!response.data?.data) {
        throw new Error('AList API 未返回文件信息')
      }

      logger.debug.debug('成功获取文件信息', {
        path,
        fileInfo: {
          name: response.data.data.name,
          size: response.data.data.size,
          is_dir: response.data.data.is_dir,
        },
      })

      return response.data.data
    } catch (error) {
      logger.error.error('获取文件信息失败', {
        path,
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
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
    logger.info.info('正在查找存储', { path })
    try {
      const storages = await this.listStorages()
      if (!Array.isArray(storages) || storages.length === 0) {
        logger.error.error('未找到有效的存储')
        throw new Error('AList 中未找到有效的存储')
      }

      // 确保路径以斜杠开头进行匹配
      const normalizedPath = this.normalizePath(path)
      const storage = storages.find((s) => {
        const mountPath = this.normalizePath(s.mount_path)
        return normalizedPath.startsWith(mountPath)
      })

      if (!storage) {
        logger.error.error('未找到匹配的存储', { path: normalizedPath })
        throw new Error(`未找到路径对应的存储: ${path}`)
      }

      logger.debug.debug('成功找到存储', {
        path: normalizedPath,
        storage: {
          id: storage.id,
          mount_path: storage.mount_path,
          provider: storage.provider,
        },
      })

      return storage
    } catch (error) {
      logger.error.error('查找存储失败', {
        path,
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
      throw error
    }
  }
}

export default new AlistService()
