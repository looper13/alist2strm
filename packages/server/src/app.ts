import express from 'express'
import cors from 'cors'
import type { Request, Response, NextFunction } from 'express'
import config from './config'
import db from './models'
import taskService from './services/taskService'
import { errorHandler } from './middleware/errorHandler'
import taskRoutes from './routes/tasks'
import { logger } from './utils/logger'

const app = express()

app.use(cors())
app.use(express.json())

// 添加基本的请求日志中间件
app.use((req: Request, _res: Response, next: NextFunction) => {
  logger.info.info(`收到${req.method}请求: ${req.url}`, {
    body: req.body,
    query: req.query,
    params: req.params,
    ip: req.ip,
  })
  next()
})

// 路由
app.use('/api/tasks', taskRoutes)

// 任务管理 API
app.post('/api/tasks', async (req: Request, res: Response) => {
  try {
    const task = await taskService.createTask(req.body)
    logger.info.info('Task created successfully', { taskId: task.id, taskData: req.body })
    res.json(task)
  } catch (error) {
    logger.error.error('Failed to create task', { error: (error as Error).message, data: req.body })
    res.status(400).json({ error: (error as Error).message })
  }
})

app.get('/api/tasks', async (_req: Request, res: Response) => {
  try {
    const tasks = await taskService.getTasks()
    res.json(tasks)
  } catch (error) {
    res.status(500).json({ error: (error as Error).message })
  }
})

app.get('/api/tasks/:id', async (req: Request, res: Response) => {
  try {
    const task = await taskService.getTask(Number(req.params.id))
    if (!task) return res.status(404).json({ error: 'Task not found' })
    res.json(task)
  } catch (error) {
    res.status(500).json({ error: (error as Error).message })
  }
})

app.put('/api/tasks/:id', async (req: Request, res: Response) => {
  try {
    const task = await taskService.updateTask(Number(req.params.id), req.body)
    res.json(task)
  } catch (error) {
    res.status(400).json({ error: (error as Error).message })
  }
})

app.delete('/api/tasks/:id', async (req: Request, res: Response) => {
  try {
    await taskService.deleteTask(Number(req.params.id))
    res.json({ success: true })
  } catch (error) {
    res.status(400).json({ error: (error as Error).message })
  }
})

app.post('/api/tasks/:id/execute', async (req: Request, res: Response) => {
  try {
    const result = await taskService.executeTask(Number(req.params.id))
    res.json({
      success: true,
      message: 'Task executed successfully',
      ...result,
    })
  } catch (error) {
    res.status(500).json({
      success: false,
      error: (error as Error).message,
    })
  }
})

app.get('/api/tasks/:id/logs', async (req: Request, res: Response) => {
  try {
    const limit = req.query.limit ? Number(req.query.limit) : undefined
    const logs = await taskService.getTaskLogs(Number(req.params.id), limit)
    res.json(logs)
  } catch (error) {
    res.status(500).json({ error: (error as Error).message })
  }
})

// 错误处理
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  logger.error.error('发生未处理的错误', {
    error: err.message,
    stack: err.stack,
    path: req.path,
    method: req.method,
  })
  next(err)
})

app.use(errorHandler)

// 初始化数据库并启动服务器
async function start() {
  try {
    // 同步数据库模型
    await db.sequelize.sync()
    logger.info.info('数据库同步完成')

    // 初始化定时任务
    await taskService.initializeCronJobs()
    logger.info.info('定时任务初始化完成')

    // 启动服务器
    app.listen(config.server.port, () => {
      logger.info.info(`服务器启动成功`, {
        port: config.server.port,
        environment: process.env.NODE_ENV || 'development',
        config: {
          alistHost: config.alist.host,
          fileSuffix: config.generator.fileSuffix,
        },
      })
    })
  } catch (error) {
    logger.error.error('服务器启动失败', {
      error: (error as Error).message,
      stack: (error as Error).stack,
    })
    process.exit(1)
  }
}

start()

export default app
