import { Router } from 'express'
import { TaskLogService } from '@/services/task-log.service.js'
import { logger } from '@/utils/logger.js'

const router = Router()
const taskLogService = new TaskLogService()

// 创建任务日志
router.post('/', async (req, res) => {
  try {
    const taskLog = await taskLogService.create(req.body)
    res.json({
      code: 0,
      data: taskLog,
      message: '创建成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 创建任务日志:', error)
    res.status(500).json({
      code: 500,
      message: '创建失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 分页查询任务日志
router.get('/', async (req, res) => {
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
    res.json({
      code: 0,
      data: result,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 分页查询任务日志:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 根据任务ID查询最新日志
router.get('/latest/:taskId', async (req, res) => {
  try {
    const taskId = parseInt(req.params.taskId, 10)
    const taskLog = await taskLogService.findLatestByTaskId(taskId)
    if (!taskLog) {
      res.status(404).json({
        code: 404,
        message: '任务日志不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: taskLog,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 查询最新任务日志:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 根据ID查询任务日志
router.get('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const taskLog = await taskLogService.findById(id)
    if (!taskLog) {
      res.status(404).json({
        code: 404,
        message: '任务日志不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: taskLog,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 查询任务日志:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

export default router 