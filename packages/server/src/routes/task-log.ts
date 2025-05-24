import { Router } from 'express'
import type { Request, Response, NextFunction, Router as RouterType } from 'express'
import { logger } from '@/utils/logger.js'
import { taskLogService } from '@/services/task-log.service.js'
import { HttpError } from '@/middlewares/error.js'
import { success, error, pageResult } from '@/utils/response.js'

const router: RouterType = Router()

// 创建任务日志
router.post('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const log = await taskLogService.create(req.body)
    success(res, log)
  }
  catch (err) {
    logger.error.error('创建任务日志失败:', err)
    next(new HttpError('创建任务日志失败', 500))
  }
})

// 分页查询任务日志
router.get('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { page, pageSize, taskId, status, startTime, endTime } = req.query
    const result = await taskLogService.findByPage({
      page: page ? parseInt(page as string, 10) : undefined,
      pageSize: pageSize ? parseInt(pageSize as string, 10) : undefined,
      taskId: taskId ? parseInt(taskId as string, 10) : undefined,
      status: status as string,
      startTime: startTime ? new Date(startTime as string) : undefined,
      endTime: endTime ? new Date(endTime as string) : undefined,
    })
    pageResult(res, result)
  }
  catch (err) {
    logger.error.error('查询任务日志失败:', err)
    next(new HttpError('查询任务日志失败', 500))
  }
})

// 获取最新任务日志
router.get('/latest/:taskId', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const log = await taskLogService.findLatestByTaskId(Number(req.params.taskId))
    if (!log) {
      throw new HttpError('任务日志不存在', 404)
    }
    success(res, log)
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('获取最新任务日志失败:', err)
      next(new HttpError('获取最新任务日志失败', 500))
    }
  }
})

// 获取任务日志详情
router.get('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const log = await taskLogService.findById(Number(req.params.id))
    if (!log) {
      throw new HttpError('任务日志不存在', 404)
    }
    success(res, log)
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('获取任务日志失败:', err)
      next(new HttpError('获取任务日志失败', 500))
    }
  }
})

export { router as taskLogRouter } 