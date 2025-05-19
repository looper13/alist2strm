import { Router } from 'express'
import type { Request, Response, NextFunction, Router as RouterType } from 'express'
import { logger } from '@/utils/logger.js'
import { configService } from '@/services/config.service.js'
import { HttpError } from '@/middleware/error.js'
import { success } from '@/utils/response.js'

const router: RouterType = Router()

// 创建配置
router.post('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const config = await configService.create(req.body)
    success(res, config)
  }
  catch (err) {
    logger.error.error('创建配置失败:', err)
    next(new HttpError('创建配置失败', 500))
  }
})

// 更新配置
router.put('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const config = await configService.update(Number(req.params.id), req.body)
    if (!config) {
      throw new HttpError('配置不存在', 404)
    }
    success(res, config)
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('更新配置失败:', err)
      next(new HttpError('更新配置失败', 500))
    }
  }
})

// 删除配置
router.delete('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    await configService.delete(Number(req.params.id))
    success(res, null, '删除成功')
  }
  catch (err) {
    logger.error.error('删除配置失败:', err)
    next(new HttpError('删除配置失败', 500))
  }
})

// 获取所有配置
router.get('/all', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const configs = await configService.findAll()
    success(res, configs)
  }
  catch (err) {
    logger.error.error('获取所有配置失败:', err)
    next(new HttpError('获取所有配置失败', 500))
  }
})

// 获取配置详情
router.get('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const config = await configService.findById(Number(req.params.id))
    if (!config) {
      throw new HttpError('配置不存在', 404)
    }
    success(res, config)
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('获取配置失败:', err)
      next(new HttpError('获取配置失败', 500))
    }
  }
})

export { router as configRouter } 