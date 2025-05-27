import { AlistClient } from './client.js'
import { AlistConfigManager } from './config.js'
import { RetryStrategy } from './retry.js'
import { logger } from '@/utils/logger.js'
import type { AlistConfig } from '@/types/alist.js'

export class AlistService {
  private client: AlistClient | null = null
  private configManager: AlistConfigManager
  private retryStrategy: RetryStrategy | null = null

  constructor() {
    this.configManager = new AlistConfigManager()
    
    // 如果配置已就绪，立即初始化客户端
    if (this.configManager.isReady()) {
      this.initializeClient(this.configManager.getConfig())
    }

    // 监听配置更新
    this.configManager.on('configUpdated', (config: AlistConfig) => {
      if (this.client && this.retryStrategy) {
        // 如果客户端已存在，只更新配置
        this.client.updateConfig(config)
        this.retryStrategy.updateConfig(config)
        logger.info.info('AList 服务配置已更新', { config })
      } else {
        // 如果客户端不存在，进行初始化
        this.initializeClient(config)
        logger.info.info('AList 服务已初始化', { config })
      }
    })
  }

  private initializeClient(config: AlistConfig) {
    this.client = new AlistClient(config)
    this.retryStrategy = new RetryStrategy({
      maxRetries: config.maxRetries || 3,
      retryDelay: config.retryDelay || 1000,
      reqDelay: config.reqDelay || 100,
    })
  }

  private validateClient() {
    if (!this.client || !this.retryStrategy) {
      throw new Error('AList 服务未就绪，请确保已配置 host 和 token')
    }
  }

  async listFiles(path: string): Promise<App.AList.AlistFile[]> {
    this.validateClient()
    const config = this.configManager.getConfig()
    let page = 1
    let allFiles: App.AList.AlistFile[] = []
    let hasMore = true

    while (hasMore) {
      try {
        const response = await this.retryStrategy!.execute(
          () => this.client!.listFiles(path, page, config.perPage || 100),
          `获取文件列表 - 页码: ${page}`,
        )

        const files = response.data.content || []
        allFiles.push(...files)

        logger.info.info(`获取文件列表,第${page}页`, {
          path,
          currentCount: files.length,
          totalCount: allFiles.length,
        })

        hasMore = files.length === (config.perPage || 100)
        page++
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

    return allFiles
  }

  async getFileInfo(path: string): Promise<App.AList.AlistFile> {
    this.validateClient()
    try {
      const response = await this.retryStrategy!.execute(
        () => this.client!.getFileInfo(path),
        '获取文件信息',
      )

      if (!response.data) {
        throw new Error('获取文件信息失败')
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

// 导出服务实例
export const alistService = new AlistService() 