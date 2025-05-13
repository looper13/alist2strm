import type { ScheduledTask } from 'node-cron'
import cron from 'node-cron'
import { Op } from 'sequelize'
import type {
  TaskAttributes,
  TaskCreationAttributes,
  TaskInstance,
  TaskLogInstance,
  TaskStatusType,
} from '../types'
import db from '../models'
import generatorService from './generator'
import { logger } from '../utils/logger'
import { TASK_STATUS, ERROR_MSG } from '../constants'

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
      throw new Error(ERROR_MSG.TASK_NOT_FOUND)
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
      throw new Error(ERROR_MSG.TASK_NOT_FOUND)
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
  ): Promise<{ records: TaskInstance[]; total: number }> {
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
        records: rows,
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
      throw new Error(ERROR_MSG.TASK_NOT_FOUND)
    }

    // 检查任务是否正在运行中
    if (task.running) {
      logger.warn.warn('任务已在执行中，跳过本次执行', {
        taskId: id,
        name: task.name,
        lastRunAt: task.lastRunAt,
      })
      return {
        success: false,
        message: ERROR_MSG.TASK_RUNNING,
        taskId: id,
        taskName: task.name,
      }
    }

    // 设置任务为运行中状态
    await task.update({ running: true })
    logger.debug.debug('任务状态已设置为运行中', { taskId: id, name: task.name })

    const startTime = new Date()
    const taskLog = await db.TaskLog.create({
      taskId: task.id,
      startTime,
      status: TASK_STATUS.PENDING,
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
          status: TASK_STATUS.SUCCESS,
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

      // 更新任务状态并记录最后运行时间
      await task.update({
        lastRunAt: endTime,
        running: false,
      })
      logger.debug.debug('任务执行完成，状态已重置', {
        taskId: id,
        duration: endTime.getTime() - startTime.getTime(),
      })

      return result
    } catch (error) {
      const errorMessage = (error as Error).message
      const endTime = new Date()

      await db.TaskLog.update(
        {
          status: TASK_STATUS.ERROR,
          error: errorMessage,
          endTime,
        },
        {
          where: {
            id: taskLog.id,
          },
        },
      )

      // 即使出错也要重置任务状态
      await task.update({ running: false })
      logger.debug.debug('任务执行失败，状态已重置', { taskId: id })

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
        const result = await this.executeTask(task.id)

        // 检查任务是否已在运行中
        if (
          result &&
          typeof result === 'object' &&
          'success' in result &&
          result.success === false
        ) {
          logger.debug.debug('定时任务已在执行中，跳过本次执行', {
            taskId: task.id,
            message: result.message,
          })
          return
        }

        logger.debug.debug('定时任务执行完成', {
          taskId: task.id,
          result,
        })
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
   * 重置所有卡住的任务状态
   * 在系统启动时调用，确保没有任务因为系统崩溃等原因卡在running状态
   */
  async resetStuckTasks() {
    try {
      const runningTasks = await db.Task.findAll({
        where: { running: true },
      })

      if (runningTasks.length > 0) {
        logger.warn.warn('发现处于运行状态的任务，准备重置', { count: runningTasks.length })

        for (const task of runningTasks) {
          await task.update({ running: false })
          logger.info.info('重置任务运行状态', {
            taskId: task.id,
            name: task.name,
          })
        }
      } else {
        logger.debug.debug('没有发现需要重置的任务')
      }
    } catch (error) {
      logger.error.error('重置任务状态失败', {
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
    }
  }

  /**
   * 初始化任务
   */
  async initializeCronJobs() {
    // 先重置所有任务的运行状态
    await this.resetStuckTasks()

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
