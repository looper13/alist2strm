import { Router } from 'express'
import { TaskService } from '@/services/task.service.js'
import { taskQueue } from '@/services/task-queue.service.js'
import { taskLogService } from '@/services/task-log.service.js'
import { logger } from '@/utils/logger.js'

const router = Router()
const taskService = new TaskService()

// 创建任务
router.post('/', async (req, res) => {
  try {
    const task = await taskService.create(req.body)
    res.json({
      code: 0,
      data: task,
      message: '创建成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 创建任务:', error)
    res.status(500).json({
      code: 500,
      message: '创建失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 更新任务
router.put('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const task = await taskService.update(id, req.body)
    if (!task) {
      res.status(404).json({
        code: 404,
        message: '任务不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: task,
      message: '更新成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 更新任务:', error)
    res.status(500).json({
      code: 500,
      message: '更新失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 删除任务
router.delete('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const success = await taskService.delete(id)
    if (!success) {
      res.status(404).json({
        code: 404,
        message: '任务不存在',
      })
      return
    }
    res.json({
      code: 0,
      message: '删除成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 删除任务:', error)
    res.status(500).json({
      code: 500,
      message: '删除失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 分页查询任务
router.get('/', async (req, res) => {
  try {
    const { page, pageSize, keyword, enabled, running } = req.query
    const result = await taskService.findByPage({
      page: page ? parseInt(page as string, 10) : undefined,
      pageSize: pageSize ? parseInt(pageSize as string, 10) : undefined,
      keyword: keyword as string,
      enabled: enabled === 'true' ? true : enabled === 'false' ? false : undefined,
      running: running === 'true' ? true : running === 'false' ? false : undefined,
    })
    res.json({
      code: 0,
      data: result,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 分页查询任务:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 查询所有任务
router.get('/all', async (req, res) => {
  try {
    const name = req.query.name as string
    const tasks = await taskService.findAll(name)
    res.json({
      code: 0,
      data: tasks,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 查询所有任务:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 根据ID查询任务
router.get('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const task = await taskService.findById(id)
    if (!task) {
      res.status(404).json({
        code: 404,
        message: '任务不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: task,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 查询任务:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 更新任务运行状态
router.put('/:id/status', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const { running } = req.body
    const task = await taskService.updateRunningStatus(id, running)
    if (!task) {
      res.status(404).json({
        code: 404,
        message: '任务不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: task,
      message: '更新成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 更新任务运行状态:', error)
    res.status(500).json({
      code: 500,
      message: '更新失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 执行任务
router.post('/:id/execute', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const success = await taskService.execute(id)
    if (!success) {
      res.status(400).json({
        code: 400,
        message: '执行任务失败',
      })
      return
    }
    res.json({
      code: 0,
      message: '任务开始执行',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 执行任务:', error)
    res.status(500).json({
      code: 500,
      message: '执行失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 终止任务
router.post('/:id/terminate', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const success = await taskService.stop(id)
    
    if (!success) {
      res.status(400).json({
        code: 400,
        message: '终止任务失败'
      })
      return
    }

    res.json({
      code: 0,
      message: '任务已终止'
    })
  }
  catch (error) {
    logger.error.error('终止任务失败:', error)
    res.status(500).json({
      code: 500,
      message: '终止失败',
      error: error instanceof Error ? error.message : '未知错误'
    })
  }
})


// 获取任务状态
router.get('/:id/status', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const status = taskService.getTaskStatus(id)
    
    if (!status) {
      // 如果任务不在运行中，从数据库获取最后的状态
      const task = await taskService.findById(id)
      if (!task) {
        res.status(404).json({
          code: 404,
          message: '任务不存在'
        })
        return
      }

      res.json({
        code: 0,
        data: {
          status: task.running ? 'running' : 'stopped',
          progress: 0,
          lastRunAt: task.lastRunAt
        },
        message: '查询成功'
      })
      return
    }

    res.json({
      code: 0,
      data: status,
      message: '查询成功'
    })
  }
  catch (error) {
    logger.error.error('获取任务状态失败:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误'
    })
  }
})

// 停止任务
router.post('/:id/stop', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const success = await taskService.stop(id)
    
    if (!success) {
      res.status(400).json({
        code: 400,
        message: '停止任务失败'
      })
      return
    }

    res.json({
      code: 0,
      message: '任务已停止'
    })
  }
  catch (error) {
    logger.error.error('停止任务失败:', error)
    res.status(500).json({
      code: 500,
      message: '停止失败',
      error: error instanceof Error ? error.message : '未知错误'
    })
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

    const logs = await taskLogService.findByTaskId(taskId)
    res.json(logs)
  }
  catch (error) {
    logger.error.error('获取任务日志失败:', error)
    res.status(500).json({ message: '获取任务日志失败' })
  }
})

export default router 