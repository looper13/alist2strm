import { Sequelize } from 'sequelize'
import type { TaskInstance, TaskLogInstance } from '../types'
import Task from './task'
import TaskLog from './taskLog'
import { logger } from '../utils/logger'

// 创建 Sequelize 实例
export const sequelize = new Sequelize({
  dialect: 'sqlite',
  storage: './database.sqlite',
  logging: (msg) => logger.debug.debug('数据库查询', { query: msg }),
  logQueryParameters: true,
  benchmark: true,
})

// 数据库事件监听
sequelize.afterBulkSync(() => {
  logger.info.info('数据库表同步完成')
})

// 测试数据库连接
sequelize
  .authenticate()
  .then(() => {
    logger.info.info('数据库连接建立成功')
  })
  .catch((err) => {
    logger.error.error('无法连接到数据库', {
      error: err.message,
      stack: err.stack,
    })
  })

// 初始化模型
Task.initModel(sequelize)
TaskLog.initModel(sequelize)

// 建立模型关联
Task.associate({ Task, TaskLog })
TaskLog.associate({ Task, TaskLog })

logger.debug.debug('模型初始化完成', {
  models: [Task.name, TaskLog.name],
})

// 导出数据库实例和模型
const db = {
  sequelize,
  Task,
  TaskLog,
}

logger.debug.debug('模型关联建立完成', {
  associations: [
    { model: Task.name, relation: 'hasMany', target: TaskLog.name },
    { model: TaskLog.name, relation: 'belongsTo', target: Task.name },
  ],
})

export default db
