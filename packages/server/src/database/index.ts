import { Sequelize } from 'sequelize-typescript'
import path from 'path'
import config from '@/config'
import { logger } from '@/utils/logger'
import { Config } from '@/models/config'
import { Task } from '@/models/task'
import { TaskLog } from '@/models/task-log'
import { FileHistory } from '@/models/file-history'
import fs from 'fs'

// 确保数据库目录存在
const dbDir = path.dirname(path.join(config.database.path, config.database.name))
if (!fs.existsSync(dbDir)) {
  fs.mkdirSync(dbDir, { recursive: true })
}

const sequelize = new Sequelize({
  dialect: 'sqlite',
  storage: path.join(config.database.path, config.database.name),
  models: [Config, Task, TaskLog, FileHistory],
  logging: (msg) => logger.debug.debug(msg),
})

export async function setupDatabase(): Promise<void> {
  try {
    await sequelize.authenticate()
    logger.info.info('数据库连接已成功建立')

    // 同步数据库结构
    await sequelize.sync({ alter: true })
    logger.info.info('数据库结构同步完成')
  }
  catch (error) {
    logger.error.error('无法连接到数据库:', error)
    throw error
  }
}

export { sequelize }
export * from '@/models/config'
export * from '@/models/task'
export * from '@/models/task-log'
export * from '@/models/file-history' 