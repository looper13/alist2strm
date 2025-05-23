import { Config } from '@/models/config.js'
import type { WhereOptions } from 'sequelize'
import { Op } from 'sequelize'
import { logger } from '@/utils/logger.js'
import { configCache } from './config-cache.service.js'

export class ConfigService {
  /**
   * 创建配置
   */
  async create(data: App.Config.Create): Promise<Config> {
    try {
      const {code} = data
      const existingConfig = await Config.findOne({ where: { code } })
      if (existingConfig) {
        logger.warn.warn(`创建配置失败: 配置code: 【${code}】已存在`, { code: data.code })
        throw new Error(`配置code重复：【${code}】`)
      }
      const config = await Config.create(data as any)
      // 更新缓存
      configCache.set(config.code, config.value)
      logger.info.info('创建配置成功:', { id: config.id, code: config.code })
      return config
    }
    catch (error) {
      logger.error.error('创建配置失败:', error)
      throw error
    }
  }

  /**
   * 更新配置
   */
  async update(id: number, data: App.Config.Update): Promise<Config | null> {
    try {
      const config = await Config.findByPk(id)
      if (!config) {
        logger.warn.warn('更新配置失败: 配置不存在', { id })
        return null
      }
      await config.update(data)
      if (data.value !== undefined)
        configCache.set(config.code, data.value)
      logger.info.info('更新配置成功:', { id, code: config.code })
      return config
    }
    catch (error) {
      logger.error.error('更新配置失败:', error)
      throw error
    }
  }

  /**
   * 删除配置
   */
  async delete(id: number): Promise<boolean> {
    try {
      const config = await Config.findByPk(id)
      if (!config) {
        logger.warn.warn('删除配置失败: 配置不存在', { id })
        return false
      }

      await config.destroy()
      configCache.delete(config.code)
      logger.info.info('删除配置成功:', { id, code: config.code })
      return true
    }
    catch (error) {
      logger.error.error('删除配置失败:', error)
      throw error
    }
  }

  /**
   * 分页查询配置
   */
  async findByPage(query: App.Config.Query): Promise<App.Common.PaginationResult<Config>> {
    try {
      const { page = 1, pageSize = 10, keyword } = query
      const where: WhereOptions<Models.ConfigAttributes> = {}

      // 如果有关键字，搜索名称和代码
      if (keyword) {
        Object.assign(where, {
          [Op.or]: [
            { name: { [Op.like]: `%${keyword}%` } },
            { code: { [Op.like]: `%${keyword}%` } },
          ],
        } as WhereOptions<Models.ConfigAttributes>)
      }

      const { count, rows } = await Config.findAndCountAll({
        where,
        offset: (page - 1) * pageSize,
        limit: pageSize,
        order: [['createdAt', 'DESC']],
      })
      logger.debug.debug('分页查询配置:', { page, pageSize, total: count })
      return {
        list: rows,
        total: count,
        page,
        pageSize,
      }
    }
    catch (error) {
      logger.error.error('分页查询配置失败:', error)
      throw error
    }
  }

  /**
   * 查询所有配置
   */
  async findAll(): Promise<Config[]> {
    try {
      const configs = await Config.findAll({
        order: [['createdAt', 'DESC']],
      })
      logger.debug.debug('查询所有配置:', { total: configs.length })
      return configs
    }
    catch (error) {
      logger.error.error('查询所有配置失败:', error)
      throw error
    }
  }

  /**
   * 根据ID查询配置
   */
  async findById(id: number): Promise<Config | null> {
    try {
      const config = await Config.findByPk(id)
      if (!config)
        logger.warn.warn('查询配置失败: 配置不存在', { id })
      return config
    }
    catch (error) {
      logger.error.error('查询配置失败:', error)
      throw error
    }
  }

  /**
   * 根据代码查询配置
   */
  async findByCode(code: string): Promise<Config | null> {
    try {
      const config = await Config.findOne({ where: { code } })
      if (!config)
        logger.warn.warn('查询配置失败: 配置不存在', { code })
      return config
    }
    catch (error) {
      logger.error.error('查询配置失败:', error)
      throw error
    }
  }
}

export const configService = new ConfigService() 