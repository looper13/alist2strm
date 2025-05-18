import 'reflect-metadata'
import 'dotenv/config'
import express from 'express'
import type { Express } from 'express'
import { getLogger } from 'log4js'
import { setupLogger, httpLogger } from './utils/logger'
import { setupDatabase } from './database'
import { setupRoutes } from './routes'
import { configCache } from './services/config-cache.service'

const logger = getLogger('app')
const app: Express = express()
const port = process.env.PORT || 3000

async function bootstrap() {
  // 设置日志
  setupLogger()

  // 设置数据库
  await setupDatabase()

  // 中间件
  app.use(httpLogger) // 添加日志中间件
  app.use(express.json())
  app.use(express.urlencoded({ extended: true }))

  configCache.initialize()

  // 路由
  setupRoutes(app)

  // 启动服务器
  app.listen(port, () => {
    logger.info(`服务器已启动，监听端口 ${port}`)
  })
}

bootstrap().catch((error) => {
  logger.error('服务器启动失败:', error)
  process.exit(1)
}) 