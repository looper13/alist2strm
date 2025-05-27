import type { Express, Request, Response } from 'express'
import { errorHandler, notFoundHandler } from '@/middlewares/error.js'
import { configRouter } from './config.js'
import { taskRouter } from './task.js'
import { taskLogRouter } from './task-log.js'
import { fileHistoryRouter } from './file-history.js'
import userRouter from './user.js'
import { auth } from '@/middlewares/auth.js'
import { success } from '@/utils/response.js'

export function setupRoutes(app: Express): void {
  // 用户相关路由（不需要认证）
  app.use('/api/users', userRouter)

  // 健康检查（不需要认证）
  app.get('/health', (_req: Request, res: Response) => {
    success(res, { status: 'ok' })
  })

  // 以下路由需要认证
  app.use('/api/configs', auth, configRouter)
  app.use('/api/tasks', auth, taskRouter)
  app.use('/api/task-logs', auth, taskLogRouter)
  app.use('/api/file-histories', auth, fileHistoryRouter)

  // 404 处理
  app.use(notFoundHandler)

  // 错误处理
  app.use(errorHandler)
} 