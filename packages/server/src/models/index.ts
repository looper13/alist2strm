import { Sequelize } from 'sequelize'
import type { TaskInstance, TaskLogInstance } from '../types'
import Task from './task'
import TaskLog from './taskLog'

const sequelize = new Sequelize({
  dialect: 'sqlite',
  storage: './database.sqlite',
  logging: false,
})

// 初始化模型
Task.initModel(sequelize)
TaskLog.initModel(sequelize)

// 设置关联关系
Task.hasMany(TaskLog, { foreignKey: 'taskId' })
TaskLog.belongsTo(Task, { foreignKey: 'taskId' })

// 导出模型和数据库连接
const db = {
  sequelize,
  Sequelize,
  Task,
  TaskLog,
}

export { sequelize }
export type { TaskInstance, TaskLogInstance }
export default db
