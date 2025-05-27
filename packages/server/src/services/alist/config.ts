import { configCache } from '../config-cache.service.js'
import type { AlistConfig, AlistOptions } from '@/types/alist.js'
import { ALIST_CONFIG } from '@/constant/index.js'
import { logger } from '@/utils/logger.js'
import EventEmitter from 'events'

export class AlistConfigManager extends EventEmitter {
  private config: AlistConfig = {
    host: '',
    token: '',
    perPage: 100,
    maxRetries: 3,
    retryDelay: 1000,
    reqDelay: 100,
  }

  private isConfigReady = false

  constructor() {
    super()
    this.initConfig()
    this.setupConfigListener()
  }

  private initConfig() {
    try {
      const host = configCache.get(ALIST_CONFIG.ALIST_HOST)
      const token = configCache.get(ALIST_CONFIG.ALIST_TOKEN)
      const perPage = configCache.get(ALIST_CONFIG.ALIST_PER_PAGE)
      const maxRetries = configCache.get(ALIST_CONFIG.ALIST_REQ_RETRY_COUNT)
      const retryDelay = configCache.get(ALIST_CONFIG.ALIST_REQ_RETRY_INTERVAL)
      const reqDelay = configCache.get(ALIST_CONFIG.ALIST_REQ_INTERVAL)

      if (host) this.config.host = host
      if (token) this.config.token = token
      if (perPage) this.config.perPage = parseInt(perPage)
      if (maxRetries) this.config.maxRetries = parseInt(maxRetries)
      if (retryDelay) this.config.retryDelay = parseInt(retryDelay)
      if (reqDelay) this.config.reqDelay = parseInt(reqDelay)

      this.checkAndEmitReady()

      logger.info.info('AList 配置已初始化', { config: this.config })
    } catch (error) {
      logger.error.error('AList 配置初始化失败', { error })
    }
  }

  private setupConfigListener() {
    configCache.on('configCacheInitialized', () => {
      logger.info.info('配置缓存初始化成功, 开始初始化 AList 配置')
      this.initConfig()
    })
    configCache.on('configUpdated', ({ code, value }) => {
      if (this.isAlistConfig(code)) {
        this.updateConfig(code, value)
        this.checkAndEmitReady()
      }
    })
  }

  private checkAndEmitReady() {
    const wasReady = this.isConfigReady
    this.isConfigReady = Boolean(this.config.host && this.config.token)

    if (!wasReady && this.isConfigReady) {
      logger.info.info('AList 必要配置已就绪', {
        host: this.config.host,
        hasToken: Boolean(this.config.token)
      })
      this.emit('configUpdated', this.getConfig())
    }
  }

  private isAlistConfig(code: string): boolean {
    return [
      ALIST_CONFIG.ALIST_HOST,
      ALIST_CONFIG.ALIST_TOKEN,
      ALIST_CONFIG.ALIST_PER_PAGE,
      ALIST_CONFIG.ALIST_REQ_RETRY_COUNT,
      ALIST_CONFIG.ALIST_REQ_RETRY_INTERVAL,
      ALIST_CONFIG.ALIST_REQ_INTERVAL,
    ].includes(code)
  }

  private updateConfig(code: string, value: string) {
    switch (code) {
      case ALIST_CONFIG.ALIST_HOST:
        this.config.host = value
        break
      case ALIST_CONFIG.ALIST_TOKEN:
        this.config.token = value
        break
      case ALIST_CONFIG.ALIST_PER_PAGE:
        this.config.perPage = parseInt(value)
        if (this.isConfigReady) {
          this.emit('configUpdated', this.getConfig())
        }
        break
      case ALIST_CONFIG.ALIST_REQ_RETRY_COUNT:
        this.config.maxRetries = parseInt(value)
        if (this.isConfigReady) {
          this.emit('configUpdated', this.getConfig())
        }
        break
      case ALIST_CONFIG.ALIST_REQ_RETRY_INTERVAL:
        this.config.retryDelay = parseInt(value)
        if (this.isConfigReady) {
          this.emit('configUpdated', this.getConfig())
        }
        break
      case ALIST_CONFIG.ALIST_REQ_INTERVAL:
        this.config.reqDelay = parseInt(value)
        if (this.isConfigReady) {
          this.emit('configUpdated', this.getConfig())
        }
        break
    }

    logger.info.info('AList 配置已更新', {
      code,
      value,
      config: this.config,
    })
  }

  getConfig(): AlistConfig {
    return { ...this.config }
  }

  getOptions(): AlistOptions {
    const { perPage, maxRetries, retryDelay, reqDelay } = this.config
    return { perPage, maxRetries, retryDelay, reqDelay }
  }

  isReady(): boolean {
    return this.isConfigReady
  }
} 