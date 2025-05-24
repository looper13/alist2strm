import { Router } from 'express'
import type { Request, Response, NextFunction, Router as RouterType } from 'express'
import { logger } from '@/utils/logger.js'
import { taskService } from '@/services/task.service.js'
import { HttpError } from '@/middlewares/error.js'
import { success, error, pageResult } from '@/utils/response.js'
import { taskLogService } from '@/services/task-log.service.js'

const router: RouterType = Router()

// 创建任务
router.post('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const task = await taskService.create(req.body)
    success(res, task)
  }
  catch (err) {
    logger.error.error('创建任务失败:', err)
    next(new HttpError('创建任务失败', 500))
  }
})

// 更新任务
router.put('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const task = await taskService.update(Number(req.params.id), req.body)
    success(res, task)
  }
  catch (err) {
    logger.error.error('更新任务失败:', err)
    next(new HttpError('更新任务失败', 500))
  }
})

// 删除任务
router.delete('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    await taskService.delete(Number(req.params.id))
    success(res, null, '删除成功')
  }
  catch (err) {
    logger.error.error('删除任务失败:', err)
    next(new HttpError('删除任务失败', 500))
  }
})

// 分页查询任务
router.get('/', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { page, pageSize, keyword, enabled, running } = req.query
    const result = await taskService.findByPage({
      page: Number(page),
      pageSize: Number(pageSize),
      keyword: keyword as string,
      enabled: enabled === 'true' ? true : enabled === 'false' ? false : undefined,
      running: running === 'true' ? true : running === 'false' ? false : undefined,
    })
    pageResult(res, result)
  }
  catch (err) {
    logger.error.error('分页查询任务失败:', err)
    next(new HttpError('分页查询任务失败', 500))
  }
})

// 获取所有任务
router.get('/all', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const tasks = await taskService.findAll()
    success(res, tasks)
  }
  catch (err) {
    logger.error.error('获取所有任务失败:', err)
    next(new HttpError('获取所有任务失败', 500))
  }
})

// 获取任务详情
router.get('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const task = await taskService.findById(Number(req.params.id))
    if (!task) {
      throw new HttpError('Task not found', 404)
    }
    success(res, task)
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('获取任务详情失败:', err)
      next(new HttpError('获取任务详情失败', 500))
    }
  }
})

// 执行任务
router.post('/:id/execute', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const task = await taskService.findById(Number(req.params.id))
    if (!task) {
      throw new HttpError('任务不存在', 404)
    }
    taskService.execute(task.id)
    success(res, null, '任务已开始执行')
  }
  catch (err) {
    if (err instanceof HttpError)
      next(err)
    else {
      logger.error.error('执行任务失败:', err)
      next(new HttpError('执行任务失败', 500))
    }
  }
})

/**
 * 获取任务日志
 */
router.get('/:id/logs', async (req, res) => {
  try {
    const taskId = parseInt(req.params.id)
    if (isNaN(taskId)) {
      return res.status(400).json({ message: '无效的任务ID' })
    }
    const { page, pageSize } = req.query
    const result = await taskLogService.findByPage({
      taskId,
      page: page ? parseInt(page as string, 10) : undefined,
      pageSize: pageSize ? parseInt(pageSize as string, 10) : undefined,
    })
    pageResult(res, result)
  }
  catch (error) {
    logger.error.error('获取任务日志失败:', error)
    res.status(500).json({ message: '获取任务日志失败' })
  }
})

// 重置任务状态
router.post('/:id/reset-status', async (req, res) => {
  try {
    const taskId = parseInt(req.params.id)
    if (isNaN(taskId)) {
      return res.status(400).json({ message: '无效的任务ID' })
    }
    await taskService.resetStatus(taskId)
    success(res, null, '任务状态已重置')
  }
  catch (error) {
    logger.error.error('重置任务状态失败:', error)
    res.status(500).json({ message: '重置任务状态失败' })
  }
})

export { router as taskRouter } 