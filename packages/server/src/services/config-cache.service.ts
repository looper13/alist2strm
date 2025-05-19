import { Config } from '@/models/config.js'
import { logger } from '@/utils/logger.js'
import EventEmitter from 'events'

export class ConfigCacheService extends EventEmitter {
  private static instance: ConfigCacheService
  private configCache: Map<string, string> = new Map()
  private initialized = false

  private constructor() {
    super()
  }

  static getInstance(): ConfigCacheService {
    if (!ConfigCacheService.instance)
      ConfigCacheService.instance = new ConfigCacheService()
    return ConfigCacheService.instance
  }

  async initialize(): Promise<void> {
    if (this.initialized)
      return

    try {
      const configs = await Config.findAll()
      this.configCache.clear()
      configs.forEach(config => this.configCache.set(config.code, config.value))
      this.initialized = true
      logger.info.info('配置缓存初始化成功')
    }
    catch (error) {
      logger.error.error('配置缓存初始化失败:', error)
      throw error
    }
  }

  get(code: string): string | undefined {
    return this.configCache.get(code)
  }

  getRequired(code: string): string {
    const value = this.get(code)
    if (!value)
      throw new Error(`Required config not found: ${code}`)
    return value
  }

  set(code: string, value: string): void {
    this.configCache.set(code, value)
    this.emit('configUpdated', { code, value })
  }

  delete(code: string): void {
    this.configCache.delete(code)
    this.emit('configDeleted', { code })
  }

  getAll(): Map<string, string> {
    return new Map(this.configCache)
  }

  isInitialized(): boolean {
    return this.initialized
  }
}

export const configCache = ConfigCacheService.getInstance() 