import 'reflect-metadata'
import 'dotenv/config'
import express from 'express'
import type { Express } from 'express'
import { setupLogger, httpLogger, logger } from '@/utils/logger.js'
import { setupDatabase } from '@/database/index.js'
import { setupRoutes } from '@/routes/index.js'
import { initConfigCache } from '@/services/config-cache.service.js'
import { taskScheduler } from '@/services/task-scheduler.service.js'
import { userService } from '@/services/user.service.js'


const app: Express = express()
const port = process.env.PORT || 3210

async function bootstrap() {
  // 设置日志
  setupLogger()

  // 设置数据库
  await setupDatabase()

  // 初始化管理员用户
  await userService.initAdminUser()

  // 中间件
  app.use(httpLogger)
  app.use(express.json())
  app.use(express.urlencoded({ extended: true }))

  // 初始化配置缓存
  await initConfigCache()

  // 初始化定时任务调度器
  taskScheduler

  // 路由
  setupRoutes(app)

  // 启动服务器
  app.listen(port, () => {
    logger.info.info(`服务器已启动，监听端口 ${port}`)
  })
}

bootstrap().catch((error) => {
  logger.error.error('服务器启动失败:', error)
  process.exit(1)
}) 