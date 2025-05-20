import { Sequelize } from 'sequelize-typescript'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { existsSync, mkdirSync } from 'node:fs'
import config from '@/config.js'
import { logger } from '@/utils/logger.js'
import { Config } from '@/models/config.js'
import { Task } from '@/models/task.js'
import { TaskLog } from '@/models/task-log.js'
import { FileHistory } from '@/models/file-history.js'
import { setupAssociations } from '@/models/associations.js'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// 确保数据库目录存在
const dbDir = dirname(join(config.database.path, config.database.name))
if (!existsSync(dbDir)) {
  mkdirSync(dbDir, { recursive: true })
}

const sequelize = new Sequelize({
  dialect: 'sqlite',
  storage: join(config.database.path, config.database.name),
  models: [Config, Task, TaskLog, FileHistory],
  logging: (msg) => logger.debug.debug(msg),
})

export async function setupDatabase(): Promise<void> {
  try {
    await sequelize.authenticate()
    logger.info.info('数据库连接已成功建立')

    // 设置模型关联
    setupAssociations()

    // 同步数据库结构，但不自动修改表结构
    await sequelize.sync({ force: false, alter: false })
    logger.info.info('数据库结构同步完成')
  }
  catch (error) {
    logger.error.error('无法连接到数据库:', error)
    throw error
  }
}

export { sequelize }
export * from '@/models/config.js'
export * from '@/models/task.js'
export * from '@/models/task-log.js'
export * from '@/models/file-history.js' 