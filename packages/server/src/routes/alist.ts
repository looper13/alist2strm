import { Router } from 'express'
import type { Request, Response, NextFunction, Router as RouterType } from 'express'
import { logger } from '@/utils/logger.js'
import { alistService } from '@/services/alist.service.js'
import { HttpError } from '@/middleware/error.js'

const router: RouterType = Router()

// 获取文件列表
router.get('/listFiles', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { path = '/' } = req.query
    const files = await alistService.listFiles(path as string)
    res.json({
      code: 0,
      message: 'success',
      data: files,
    })
  }
  catch (error) {
    logger.error.error('Failed to list files:', error)
    next(new HttpError('Failed to list files', 500))
  }
})


// 获取文件信息
router.get('/fileInfo', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { path } = req.query
    if (!path) {
      throw new HttpError('路径不能为空', 400)
    }
    const fileInfo = await alistService.getFileInfo(path as string)
    res.json({
      code: 0,
      message: 'success',
      data: fileInfo,
    })
  }
  catch (error) {
    if (error instanceof HttpError)
      next(error)
    else {
      logger.error.error('Failed to get file info:', error)
      next(new HttpError('Failed to get file info', 500))
    }
  }
})

export { router as alistRouter } 