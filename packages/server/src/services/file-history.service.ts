import { FileHistory } from '@/models/file-history'
import type { WhereOptions } from 'sequelize'
import { Op } from 'sequelize'
import { logger } from '@/utils/logger'

export interface CreateFileHistoryDto {
  fileName: string
  sourcePath: string
  targetFilePath: string
  fileSize: number
  fileType: string
  fileSuffix: string
}

export interface QueryFileHistoryDto {
  page?: number
  pageSize?: number
  keyword?: string
  fileType?: string
  fileSuffix?: string
  startTime?: Date
  endTime?: Date
}

export class FileHistoryService {
  /**
   * 创建文件历史
   */
  async create(data: CreateFileHistoryDto): Promise<FileHistory> {
    try {
      const fileHistory = await FileHistory.create(data as any)
      logger.info.info('创建文件历史成功:', { id: fileHistory.id, fileName: fileHistory.fileName })
      return fileHistory
    }
    catch (error) {
      logger.error.error('创建文件历史失败:', error)
      throw error
    }
  }

  /**
   * 分页查询文件历史
   */
  async findByPage(query: QueryFileHistoryDto): Promise<Services.PageResult<FileHistory>> {
    try {
      const { page = 1, pageSize = 10, keyword, fileType, fileSuffix, startTime, endTime } = query
      const where: WhereOptions<Models.FileHistoryAttributes> = {}

      // 关键字搜索
      if (keyword) {
        Object.assign(where, {
          [Op.or]: [
            { fileName: { [Op.like]: `%${keyword}%` } },
            { sourcePath: { [Op.like]: `%${keyword}%` } },
            { targetFilePath: { [Op.like]: `%${keyword}%` } },
          ],
        } as WhereOptions<Models.FileHistoryAttributes>)
      }

      // 文件类型过滤
      if (fileType)
        where.fileType = fileType

      // 文件后缀过滤
      if (fileSuffix)
        where.fileSuffix = fileSuffix

      // 时间范围过滤
      if (startTime || endTime) {
        where.createdAt = {}
        if (startTime)
          Object.assign(where.createdAt, { [Op.gte]: startTime })
        if (endTime)
          Object.assign(where.createdAt, { [Op.lte]: endTime })
      }

      const { count, rows } = await FileHistory.findAndCountAll({
        where,
        offset: (page - 1) * pageSize,
        limit: pageSize,
        order: [['createdAt', 'DESC']],
      })

      logger.debug.debug('分页查询文件历史:', { page, pageSize, total: count })
      return {
        list: rows,
        total: count,
        page,
        pageSize,
      }
    }
    catch (error) {
      logger.error.error('分页查询文件历史失败:', error)
      throw error
    }
  }

  /**
   * 根据ID查询文件历史
   */
  async findById(id: number): Promise<FileHistory | null> {
    try {
      const fileHistory = await FileHistory.findByPk(id)
      if (!fileHistory)
        logger.warn.warn('查询文件历史失败: 历史记录不存在', { id })
      return fileHistory
    }
    catch (error) {
      logger.error.error('查询文件历史失败:', error)
      throw error
    }
  }

  /**
   * 检查文件是否已存在
   */
  async checkFileExists(sourcePath: string, fileName: string): Promise<boolean> {
    try {
      const count = await FileHistory.count({
        where: {
          sourcePath,
          fileName,
        },
      })
      return count > 0
    }
    catch (error) {
      logger.error.error('检查文件是否存在失败:', error)
      throw error
    }
  }
} 