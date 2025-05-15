import express from 'express'
import cors from 'cors'
import type { Request, Response, NextFunction } from 'express'
import config from './config'
import db from './models'
import taskService from './services/taskService'
import { errorHandler } from './middleware/errorHandler'
import taskRoutes from './routes/tasks'
import alistRoutes from './routes/alist'
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

// 使用任务路由模块
app.use('/api/tasks', taskRoutes)
app.use('/api/alist', alistRoutes)
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
          database: {
            path: config.database.path,
            name: config.database.name,
          },
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
