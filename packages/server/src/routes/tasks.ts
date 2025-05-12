import { Router } from 'express'
import { AppError } from '../middleware/errorHandler'
import taskService from '../services/taskService'
import { logger, errorLogger } from '../utils/logger'

const router = Router()

// 获取所有任务
router.get('/', async (req, res, next) => {
  try {
    const tasks = await taskService.getTasks()
    res.json(tasks)
  } catch (error) {
    next(error)
  }
})

// 创建任务
router.post('/', async (req, res, next) => {
  try {
    const task = await taskService.createTask(req.body)
    res.status(201).json(task)
  } catch (error) {
    next(error)
  }
})

// 更新任务
router.patch('/:id', async (req, res, next) => {
  try {
    const task = await taskService.updateTask(Number(req.params.id), req.body)
    if (!task) {
      throw new AppError(404, '任务不存在')
    }
    res.json(task)
  } catch (error) {
    next(error)
  }
})

// 删除任务
router.delete('/:id', async (req, res, next) => {
  try {
    await taskService.deleteTask(Number(req.params.id))
    res.status(204).end()
  } catch (error) {
    next(error)
  }
})

// 执行任务
router.post('/:id/execute', async (req, res, next) => {
  try {
    const result = await taskService.executeTask(Number(req.params.id))
    logger.info(`Task executed successfully: ${req.params.id}`)
    res.json(result)
  } catch (error) {
    errorLogger.error(`Failed to execute task: ${req.params.id}`, { error })
    next(error)
  }
})

// 获取任务日志
router.get('/:id/logs', async (req, res, next) => {
  try {
    const logs = await taskService.getTaskLogs(Number(req.params.id))
    logger.info(`Fetched logs for task: ${req.params.id}`)
    res.json(logs)
  } catch (error) {
    errorLogger.error(`Failed to fetch logs for task: ${req.params.id}`, { error })
    next(error)
  }
})

export default router
