import { Router } from 'express'
import type { Request, Response, NextFunction, Router as RouterType } from 'express'
import { logger } from '../utils/logger'
import { alistService } from '../services/alist'
import { HttpError } from '../middleware/error'

const router: RouterType = Router()

// 获取文件列表
router.get('/list', async (req: Request, res: Response, next: NextFunction) => {
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

// 获取目录列表
router.get('/dirs', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { path = '/' } = req.query
    const files = await alistService.listDirs(path as string)
  }catch (error) {
    logger.error.error('Failed to list dir:', error)
    next(new HttpError('Failed to list dir', 500))
  }
})

// 获取文件信息
router.get('/info', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { path } = req.query
    if (!path) {
      throw new HttpError('Path is required', 400)
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

// 获取媒体文件信息
router.get('/media', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { path } = req.query
    if (!path) {
      throw new HttpError('Path is required', 400)
    }
    const mediaInfo = await alistService.getMediaInfo(path as string)
    res.json({
      code: 0,
      message: 'success',
      data: mediaInfo,
    })
  }
  catch (error) {
    if (error instanceof HttpError)
      next(error)
    else {
      logger.error.error('Failed to get media info:', error)
      next(new HttpError('Failed to get media info', 500))
    }
  }
})

// 获取下载链接
router.get('/download', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { path } = req.query
    if (!path) {
      throw new HttpError('Path is required', 400)
    }
    const downloadUrl = await alistService.getDownloadUrl(path as string)
    res.json({
      code: 0,
      message: 'success',
      data: {
        url: downloadUrl,
      },
    })
  }
  catch (error) {
    if (error instanceof HttpError)
      next(error)
    else {
      logger.error.error('Failed to get download URL:', error)
      next(new HttpError('Failed to get download URL', 500))
    }
  }
})

export { router as alistRouter } 