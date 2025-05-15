import { Router } from 'express'
import { AppError } from '../middleware/errorHandler'
import alistService from '../services/alist'
import { logger } from '../utils/logger'
import { success, fail } from '../utils/response'
import { HTTP_STATUS, API_CODE, ERROR_MSG, SUCCESS_MSG } from '../constants'

const router = Router()

/**
 * 获取文件列表
 * @param {string} path - 文件路径
 */
router.get('/path', async (req, res, next) => {
  try {
    const path = (req.query.path as string) || '/'

    logger.info.info('正在获取目录列表', {
      query: req.query,
    })

    const result = await alistService.listFiles(path)

    if (!result || result.length === 0) {
      logger.warn.warn('未找到目录', { path })
      throw new AppError(HTTP_STATUS.NOT_FOUND, ERROR_MSG.DIR_NOT_FOUND)
    }
    res.json(
      success(
        result.filter((item) => item.is_dir),
        SUCCESS_MSG.FOLDER_LIST_SUCCESS,
      ),
    )
  } catch (error) {
    logger.error.error('获取任务列表失败', {
      error: (error as Error).message,
      stack: (error as Error).stack,
      query: req.query,
    })
    next(error)
  }
})

export default router
