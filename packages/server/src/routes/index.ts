import type { Express, Request, Response } from 'express'
import { alistRouter } from './alist'
import { errorHandler, notFoundHandler } from '@/middleware/error'
import configRouter from './config'
import taskRouter from './task'
import taskLogRouter from './task-log'
import fileHistoryRouter from './file-history'

export function setupRoutes(app: Express): void {
  // API 路由
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