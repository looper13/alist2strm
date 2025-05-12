import type { ScheduledTask } from 'node-cron'
import cron from 'node-cron'
import { Op } from 'sequelize'
import type {
  TaskAttributes,
  TaskCreationAttributes,
  TaskInstance,
  TaskLogInstance,
} from '../types'
import db from '../models'
import generatorService from './generator'
import { logger } from '../utils/logger'

/**
 * 任务服务
 */
class TaskService {
  private cronJobs: Map<number, ScheduledTask>

  constructor() {
    this.cronJobs = new Map()
  }

  /**
   * 创建任务
   * @param taskData 任务数据
   * @returns 任务
   */
  async createTask(taskData: Omit<TaskCreationAttributes, 'id'>): Promise<TaskInstance> {
    const task = await db.Task.create(taskData)
    logger.info.info('创建任务成功', { taskId: task.id, taskData })

    if (task.enabled && task.cronExpression) {
      this.scheduleCronJob(task)
      logger.debug.debug('定时任务已调度', {
        taskId: task.id,
        cronExpression: task.cronExpression,
      })
    }
    return task
  }

  /**
   * 更新任务
   * @param id 任务ID
   * @param taskData 任务数据
   * @returns 任务
   */
  async updateTask(id: number, taskData: Partial<TaskAttributes>): Promise<TaskInstance | null> {
    const task = await db.Task.findByPk(id)
    if (!task) {
      logger.error.error('更新任务失败：任务不存在', { taskId: id })
      throw new Error('任务不存在')
    }

    if (this.cronJobs.has(task.id)) {
      this.cronJobs.get(task.id)?.stop()
      this.cronJobs.delete(task.id)
      logger.debug.debug('已停止现有定时任务', { taskId: task.id })
    }

    await task.update(taskData)
    logger.info.info('更新任务成功', { taskId: id, updates: taskData })

    if (task.enabled && task.cronExpression) {
      this.scheduleCronJob(task)
      logger.debug.debug('新的定时任务已调度', {
        taskId: task.id,
        cronExpression: task.cronExpression,
      })
    }

    return task
  }

  /**
   * 删除任务
   * @param id 任务ID
   */
  async deleteTask(id: number): Promise<void> {
    const task = await db.Task.findByPk(id)
    if (!task) {
      logger.error.error('删除任务失败：任务不存在', { taskId: id })
      throw new Error('任务不存在')
    }

    if (this.cronJobs.has(task.id)) {
      this.cronJobs.get(task.id)?.stop()
      this.cronJobs.delete(task.id)
      logger.debug.debug('已移除定时任务', { taskId: task.id })
    }

    await task.destroy()
    logger.info.info('删除任务成功', { taskId: id })
  }

  /**
   * 获取所有任务
   * @returns 任务列表
   */
  async getTasks(): Promise<TaskInstance[]> {
    return db.Task.findAll()
  }

  /**
   * 分页获取任务列表
   * @param page 页码（从1开始）
   * @param pageSize 每页大小
   * @returns 任务列表和总数
   */
  async getTasksWithPagination(
    page = 1,
    pageSize = 10,
  ): Promise<{ tasks: TaskInstance[]; total: number }> {
    // 确保参数为数字且有效
    const validPage = Math.max(1, Number(page))
    const validPageSize = Math.max(1, Math.min(50, Number(pageSize)))
    const offset = (validPage - 1) * validPageSize

    logger.debug.debug('执行分页查询', {
      page: validPage,
      pageSize: validPageSize,
      offset,
    })

    try {
      const { count, rows } = await db.Task.findAndCountAll({
        offset,
        limit: validPageSize,
        order: [['createdAt', 'DESC']],
        distinct: true, // 确保正确的总数计算
      })

      logger.debug.debug('查询结果', {
        total: count,
        currentPageSize: rows.length,
        page: validPage,
        pageSize: validPageSize,
      })

      return {
        tasks: rows,
        total: count,
      }
    } catch (error) {
      logger.error.error('分页查询失败', {
        error: (error as Error).message,
        stack: (error as Error).stack,
        params: { page: validPage, pageSize: validPageSize },
      })
      throw error
    }
  }

  async getTask(id: number): Promise<TaskInstance | null> {
    return db.Task.findByPk(id)
  }

  async getTaskLogs(taskId: number, limit = 10): Promise<TaskLogInstance[]> {
    return db.TaskLog.findAll({
      where: { taskId },
      order: [['createdAt', 'DESC']],
      limit,
    })
  }

  /**
   * 执行任务
   * @param id 任务ID
   */
  async executeTask(id: number) {
    const task = await db.Task.findByPk(id)
    if (!task) {
      logger.error.error('执行任务失败：任务不存在', { taskId: id })
      throw new Error('任务不存在')
    }

    const startTime = new Date()
    const taskLog = await db.TaskLog.create({
      taskId: task.id,
      startTime,
      status: 'pending',
    })

    try {
      const result = await generatorService.generateStrm({
        sourcePath: task.sourcePath,
        targetPath: task.targetPath,
        fileSuffix: task.fileSuffix.split(','),
        overwrite: task.overwrite,
      })

      const endTime = new Date()

      await db.TaskLog.update(
        {
          status: 'success',
          totalFiles: result.totalFiles,
          generatedFiles: result.generatedFiles,
          skippedFiles: result.skippedFiles,
          endTime,
        },
        {
          where: {
            id: taskLog.id,
          },
        },
      )

      await task.update({ lastRunAt: endTime })

      return result
    } catch (error) {
      const errorMessage = (error as Error).message
      const endTime = new Date()

      await db.TaskLog.update(
        {
          status: 'error',
          error: errorMessage,
          endTime,
        },
        {
          where: {
            id: taskLog.id,
          },
        },
      )

      logger.error.error('任务执行失败', {
        taskId: id,
        error: errorMessage,
        stack: (error as Error).stack,
        duration: endTime.getTime() - startTime.getTime(),
      })
      throw error
    }
  }

  /**
   * 调度任务
   * @param task 任务
   */
  private scheduleCronJob(task: TaskInstance) {
    if (!task.enabled || !task.cronExpression) {
      return
    }

    const job = cron.schedule(task.cronExpression, async () => {
      try {
        logger.debug.debug('开始执行定时任务', { taskId: task.id })
        await this.executeTask(task.id)
      } catch (error) {
        logger.error.error('定时任务执行失败', {
          taskId: task.id,
          error: (error as Error).message,
          stack: (error as Error).stack,
        })
      }
    })

    this.cronJobs.set(task.id, job)
    logger.debug.debug('定时任务已设置', {
      taskId: task.id,
      cronExpression: task.cronExpression,
    })
  }

  /**
   * 初始化任务
   */
  async initializeCronJobs() {
    const tasks = await db.Task.findAll({
      where: {
        enabled: true,
        cronExpression: {
          [Op.not]: null,
        },
      },
    })
    logger.info.info('正在初始化定时任务', { taskCount: tasks.length })
    for (const task of tasks) {
      this.scheduleCronJob(task)
    }
  }
}

export default new TaskService()
