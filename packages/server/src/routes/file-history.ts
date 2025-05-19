import { Router } from 'express'
import type { Request, Response, NextFunction, Router as RouterType } from 'express'
import { logger } from '@/utils/logger.js'
import { fileHistoryService } from '@/services/file-history.service.js'
import { HttpError } from '@/middleware/error.js'
import { success, error, pageResult } from '@/utils/response.js'

const router: RouterType = Router()

// 创建文件历史记录
router.post('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const history = await fileHistoryService.create(req.body)
    success(res, history)
  }
  catch (err) {
    logger.error.error('创建文件历史记录失败:', err)
    next(new HttpError('创建文件历史记录失败', 500))
  }
})

// 分页查询文件历史记录
router.get('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    // const { page = 1, pageSize = 10, fileName, sourcePath, startTime, endTime } = req.query
    const { page = 1, pageSize = 10, keyword, fileType, fileSuffix, startTime, endTime } = req.query
    const result = await fileHistoryService.findByPage({
      page: page ? parseInt(page as string, 10) : undefined,
      pageSize: pageSize ? parseInt(pageSize as string, 10) : undefined,
      keyword: keyword as string,
      fileType: fileType as string,
      fileSuffix: fileSuffix as string,
      startTime: startTime ? new Date(startTime as string) : undefined,
      endTime: endTime ? new Date(endTime as string) : undefined,
    })
    pageResult(res, result)
  }
  catch (err) {
    logger.error.error('查询文件历史记录失败:', err)
    next(new HttpError('查询文件历史记录失败', 500))
  }
})

// 获取文件历史记录详情
router.get('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const history = await fileHistoryService.findById(Number(req.params.id))
    if (!history) {
      throw new HttpError('文件历史记录不存在', 404)
    }
    success(res, history)
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('获取文件历史记录失败:', err)
      next(new HttpError('获取文件历史记录失败', 500))
    }
  }
})

// 检查文件是否存在
router.get('/check', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { sourcePath, fileName } = req.query
    if (!sourcePath || !fileName) {
      throw new HttpError('源路径和文件名是必填的', 400)
    }
    const exists = await fileHistoryService.checkFileExists(sourcePath as string, fileName as string)
    success(res, { exists })
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('检查文件是否存在失败:', err)
      next(new HttpError('检查文件是否存在失败', 500))
    }
  }
})

export { router as fileHistoryRouter } 