import type { ScheduledTask } from 'node-cron'
import cron from 'node-cron'
import { Op } from 'sequelize'
import type { TaskAttributes, TaskCreationAttributes } from '../types'
import db from '../models'
import generatorService from './generator'

class TaskService {
  private cronJobs: Map<number, ScheduledTask>

  constructor() {
    this.cronJobs = new Map()
  }

  async createTask(taskData: Omit<TaskCreationAttributes, 'id'>) {
    const task = await db.Task.create(taskData)
    if (task.enabled && task.cronExpression) {
      this.scheduleCronJob(task)
    }
    return task
  }

  async updateTask(id: number, taskData: Partial<TaskAttributes>) {
    const task = await db.Task.findByPk(id)
    if (!task) {
      throw new Error('Task not found')
    }

    if (this.cronJobs.has(task.id)) {
      this.cronJobs.get(task.id)?.stop()
      this.cronJobs.delete(task.id)
    }

    await task.update(taskData)

    if (task.enabled && task.cronExpression) {
      this.scheduleCronJob(task)
    }

    return task
  }

  async deleteTask(id: number) {
    const task = await db.Task.findByPk(id)
    if (!task) {
      throw new Error('Task not found')
    }

    if (this.cronJobs.has(task.id)) {
      this.cronJobs.get(task.id)?.stop()
      this.cronJobs.delete(task.id)
    }

    await task.destroy()
  }

  async getTasks() {
    return db.Task.findAll()
  }

  async getTask(id: number) {
    return db.Task.findByPk(id)
  }

  async getTaskLogs(taskId: number, limit = 10) {
    return db.TaskLog.findAll({
      where: { taskId },
      order: [['createdAt', 'DESC']],
      limit,
    })
  }

  async executeTask(id: number) {
    const task = await db.Task.findByPk(id)
    if (!task) {
      throw new Error('Task not found')
    }

    const taskLog = await db.TaskLog.create({
      taskId: task.dataValues.id,
      startTime: new Date(),
      status: 'pending',
    })

    try {
      const result = await generatorService.generateStrm({
        sourcePath: task.dataValues.sourcePath,
        targetPath: task.dataValues.targetPath,
        fileSuffix: task.dataValues.fileSuffix.split(','),
        overwrite: task.dataValues.overwrite,
      })

      await taskLog.update({
        status: 'success',
        totalFiles: result.totalFiles,
        generatedFiles: result.generatedFiles,
        skippedFiles: result.skippedFiles,
        endTime: new Date(),
      })

      await task.update({ lastRunAt: new Date() })

      return result
    } catch (error) {
      await taskLog.update({
        status: 'error',
        error: (error as Error).message,
        endTime: new Date(),
      })
      throw error
    }
  }

  private scheduleCronJob(task: TaskAttributes) {
    if (!task.enabled || !task.cronExpression) {
      return
    }

    const job = cron.schedule(task.cronExpression, async () => {
      try {
        await this.executeTask(task.id)
      } catch (error) {
        console.error(`Cron job error for task ${task.id}:`, error)
      }
    })

    this.cronJobs.set(task.id, job)
  }

  async initializeCronJobs() {
    const tasks = await db.Task.findAll({
      where: {
        enabled: true,
        cronExpression: {
          [Op.not]: null,
        },
      },
    })

    for (const task of tasks) {
      this.scheduleCronJob(task)
    }
  }
}

export default new TaskService()
