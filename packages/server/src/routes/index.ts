import type { Express, Request, Response } from 'express'
import { alistRouter } from './alist.js'
import { errorHandler, notFoundHandler } from '@/middleware/error.js'
import { configRouter } from './config.js'
import { taskRouter } from './task.js'
import { taskLogRouter } from './task-log.js'
import { fileHistoryRouter } from './file-history.js'

export function setupRoutes(app: Express): void {
  // 配置AList路由
  app.use('/api/alist', alistRouter)

  // 配置相关路由
  app.use('/api/configs', configRouter)

  // 任务相关路由
  app.use('/api/tasks', taskRouter)
  app.use('/api/task-logs', taskLogRouter)

  // 文件历史路由
  app.use('/api/file-histories', fileHistoryRouter)

  // 健康检查
  app.get('/health', (_req: Request, res: Response) => {
    res.json({ status: 'ok' })
  })

  // 404 处理
  app.use(notFoundHandler)

  // 错误处理
  app.use(errorHandler)
} 