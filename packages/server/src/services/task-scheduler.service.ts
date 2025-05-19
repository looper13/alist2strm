import cron from 'node-cron'
import { Op } from 'sequelize'
import { Task } from '@/models/task.js'
import { logger } from '@/utils/logger.js'
import { taskQueue } from './task-queue.service.js'

class TaskSchedulerService {
  private jobs: Map<number, cron.ScheduledTask> = new Map()

  constructor() {
    this.init()
  }

  /**
   * 初始化定时任务
   */
  private async init() {
    try {
      // 获取所有启用的任务
      const tasks = await Task.findAll({
        where: {
          enabled: true,
          cron: {
            [Op.ne]: null,
          },
        },
      })

      // 为每个任务创建定时器
      for (const task of tasks) {
        this.scheduleTask(task)
      }

      logger.info.info('定时任务初始化完成', { count: tasks.length })
    }
    catch (error) {
      logger.error.error('定时任务初始化失败:', error)
    }
  }

  /**
   * 调度任务
   */
  scheduleTask(task: Task) {
    // 如果任务已存在，先停止
    this.unscheduleTask(task.id)

    // 验证 cron 表达式
    if (!cron.validate(task.dataValues.cron!)) {
      logger.warn.warn('无效的 cron 表达式', { taskId: task.id, taskName: task.dataValues.name, cron: task.dataValues.cron })
      return
    }

    // 创建定时任务
    const job = cron.schedule(task.dataValues.cron!, async () => {
      try {
        // 检查任务是否正在执行
        const taskStatus = taskQueue.getTaskStatus(task.id)
        if (taskStatus?.status === 'running') {
          logger.warn.warn('任务正在执行中，跳过本次调度', { 
            taskId: task.id, 
            taskName: task.dataValues.name 
          })
          return
        }

        logger.info.info('执行定时任务', { 
          taskId: task.id,
          taskName: task.dataValues.name 
        })
        await taskQueue.addTask(task.id)
      }
      catch (error) {
        logger.error.error('定时任务执行失败:', error)
      }
    })

    // 保存任务引用
    this.jobs.set(task.id, job)
    logger.info.info('任务已调度', { taskId: task.id, cron: task.dataValues.cron })
  }

  /**
   * 取消任务调度
   */
  unscheduleTask(taskId: number) {
    const job = this.jobs.get(taskId)
    if (job) {
      job.stop()
      this.jobs.delete(taskId)
      logger.info.info('任务已取消调度', { taskId })
    }
  }

  /**
   * 更新任务调度
   */
  async updateTaskSchedule(task: Task) {
    if (task.enabled && task.cron) {
      this.scheduleTask(task)
    }
    else {
      this.unscheduleTask(task.id)
    }
  }
}

// 导出单例实例
export const taskScheduler = new TaskSchedulerService() 