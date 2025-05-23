import { getConfigCache } from './config-cache.service.js'
import { logger } from '@/utils/logger.js'
import axios, { AxiosInstance } from 'axios'
import type { ConfigCacheService } from './config-cache.service.js'

class AlistService {
  private static instance: AlistService
  private client: AxiosInstance | undefined
  private initialized = false
  private configCache: ConfigCacheService | undefined
  // 分页大小
  private perPage = 100
  // 最大重试次数
  private maxRetries = 3
  // 重试间隔
  private retryDelay = 1000
  // 请求间隔
  private reqDelay = 100

  private constructor() {
    // 空构造函数
  }

  static getInstance(): AlistService {
    if (!AlistService.instance)
      AlistService.instance = new AlistService()
    return AlistService.instance
  }

  async initialize(): Promise<void> {
    await this._initializeHttp()
  }

  private async _initializeHttp() {
    if (this.initialized)
      return

    try {
      // 获取配置缓存实例
      this.configCache = await getConfigCache()

      // 设置配置更新监听
      this.configCache.on('configUpdated', (data: { code: string }) => {
        if ([
          'ALIST_HOST',
          'ALIST_TOKEN',
          'ALIST_PER_PAGE',
          'ALIST_REQ_RETRY_COUNT',
          'ALIST_REQ_RETRY_INTERVAL',
          'ALIST_REQ_INTERVAL',
        ].includes(data.code)) {
          this._initializeHttp()
        }
      })

      const host = this.configCache.getRequired('ALIST_HOST')
      const token = this.configCache.getRequired('ALIST_TOKEN')
      const perPage = this.configCache.get('ALIST_PER_PAGE')
      const maxRetries = this.configCache.get('ALIST_REQ_RETRY_COUNT')
      const retryDelay = this.configCache.get('ALIST_REQ_RETRY_INTERVAL')
      const reqDelay = this.configCache.get('ALIST_REQ_INTERVAL')

      if (perPage)
        this.perPage = parseInt(perPage)
      if (maxRetries)
        this.maxRetries = parseInt(maxRetries)
      if (retryDelay)
        this.retryDelay = parseInt(retryDelay)
      if (reqDelay)
        this.reqDelay = parseInt(reqDelay)

      if (!host || !token) {
        logger.error.error('AList 配置未找到,请先配置')
        throw new Error('AList 配置未找到,请先配置')
      }

      this.client = axios.create({
        baseURL: host,
        headers: {
          Authorization: token,
        },
      })

      this.initialized = true
      logger.info.info('AList 服务初始化成功', {
        host,
        perPage: this.perPage,
        maxRetries: this.maxRetries,
        retryDelay: this.retryDelay,
        reqDelay: this.reqDelay,
      })
    }
    catch (error) {
      logger.error.error('AList 服务初始化失败:', error)
      this.initialized = false
      // throw error
    }
  }

  // 确保 HTTP 客户端已初始化
  private async _ensureInitialized() {
    if (!this.initialized)
      await this._initializeHttp()
    if (!this.client)
      throw new Error('HTTP client not initialized')
  }

  /**
   * 带重试机制的 API 调用
   */
  private async _retryableRequest<T>(operation: () => Promise<T>, retryCount = 0): Promise<T> {
    try {
      return await operation()
    }
    catch (error) {
      if (retryCount >= this.maxRetries) {
        logger.error.error('达到最大重试次数', {
          retryCount,
          maxRetries: this.maxRetries,
          error: error instanceof Error ? error.message : String(error),
        })
        throw error
      }

      logger.warn.warn('正在重试操作', {
        attempt: retryCount + 1,
        maxRetries: this.maxRetries,
        delay: this.retryDelay * (retryCount + 1),
      })

      await new Promise(resolve => setTimeout(resolve, this.retryDelay * (retryCount + 1)))
      return this._retryableRequest(operation, retryCount + 1)
    }
  }

  // 获取文件列表
  async listFiles(path: string): Promise<App.AList.AlistFile[]> {
    await this._ensureInitialized()
    try {
      let page = 1
      let allFiles: App.AList.AlistFile[] = []
      let hasMore = true

      while (hasMore) {
        try {
          const response = await this._retryableRequest(async () => {
            const resp = await this.client!.post<App.AList.AlistListResponse<App.AList.AlistFile[]>>('/api/fs/list', {
              path,
              password: '',
              page,
              per_page: this.perPage,
              refresh: false,
            })

            if (resp.data.code !== 200) {
              throw new Error(resp.data.message || 'AList API 返回非200状态码')
            }
            return resp
          })

          const files = response.data.data.content || []
          allFiles.push(...files)

          logger.info.info(`获取文件列表,第${page}页`, {
            path,
            currentCount: files.length,
            totalCount: allFiles.length,
          })

          hasMore = files.length === this.perPage
          page++

          // 每次请求后添加基础延迟
          await new Promise(resolve => setTimeout(resolve, this.reqDelay))
        }
        catch (error) {
          logger.error.error('获取文件列表失败，准备重试', {
            path,
            page,
            error: error instanceof Error ? error.message : String(error),
          })
          throw error
        }
      }

      return allFiles
    }
    catch (error) {
      logger.error.error('获取文件列表失败', {
        path,
        error: error instanceof Error ? error.message : String(error),
      })
      throw error
    }
  }

  // 获取文件信息
  async getFileInfo(path: string): Promise<App.AList.AlistFile> {
    await this._ensureInitialized()
    try {
      const response = await this._retryableRequest(async () => {
        const resp = await this.client!.post<App.AList.AlistGetResponse<App.AList.AlistFile>>('/api/fs/get', {
          path,
          password: '',
        })

        if (resp.data.code !== 200) {
          throw new Error(resp.data.message || 'AList API 返回非200状态码')
        }
        return resp
      })

      if (!response.data.data) {
        throw new Error('获取文件信息失败')
      }

      return response.data.data
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

export const alistService = AlistService.getInstance() 