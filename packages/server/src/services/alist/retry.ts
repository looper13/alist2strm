import { logger } from '@/utils/logger.js'
import type { AlistOptions } from '@/types/alist.js'

export class RetryStrategy {
  private maxRetries: number
  private retryDelay: number
  private reqDelay: number

  constructor(options: AlistOptions) {
    this.maxRetries = options.maxRetries || 3
    this.retryDelay = options.retryDelay || 1000
    this.reqDelay = options.reqDelay || 100
  }

  updateConfig(options: AlistOptions): void {
    this.maxRetries = options.maxRetries || 3
    this.retryDelay = options.retryDelay || 1000
    this.reqDelay = options.reqDelay || 100
  }

  private async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  async execute<T>(fn: () => Promise<T>, description: string): Promise<T> {
    let lastError: Error | null = null
    let retryCount = 0

    while (retryCount <= this.maxRetries) {
      try {
        // 如果不是第一次尝试，等待指定的延迟时间
        if (retryCount > 0) {
          await this.delay(this.retryDelay)
        }

        const result = await fn()

        // 如果不是第一次尝试，记录重试成功日志
        if (retryCount > 0) {
          logger.info.info(`${description} - 重试成功`, {
            retryCount,
          })
        }

        // 请求成功后等待指定的延迟时间
        await this.delay(this.reqDelay)

        return result
      } catch (error) {
        lastError = error as Error
        retryCount++

        if (retryCount <= this.maxRetries) {
          logger.warn.warn(`${description} - 重试中`, {
            error: lastError.message,
            retryCount,
            maxRetries: this.maxRetries,
          })
        }
      }
    }

    logger.error.error(`${description} - 重试次数已达上限`, {
      error: lastError?.message,
      retryCount: retryCount - 1,
      maxRetries: this.maxRetries,
    })

    throw lastError
  }
} 