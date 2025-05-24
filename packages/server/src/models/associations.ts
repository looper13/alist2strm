import { Task } from './task.js'
import { TaskLog } from './task-log.js'

export function setupAssociations(): void {
  // 设置 Task 和 TaskLog 之间的关联
  Task.hasMany(TaskLog, {
    foreignKey: 'taskId',
    as: 'logs',
    constraints: false,
  })

  TaskLog.belongsTo(Task, {
    foreignKey: 'taskId',
    as: 'task',
    constraints: false,
  })
} 