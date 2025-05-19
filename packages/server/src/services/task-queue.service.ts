import { EventEmitter } from 'events'
import { Task } from '@/models/task.js'
import { logger } from '@/utils/logger.js'
import { generatorService } from './generator.service.js'
import { taskLogService } from './task-log.service.js'

interface TaskJob {
  id: number
  task: Task
  status: 'pending' | 'running' | 'completed' | 'failed' | 'stopped'
  progress: number
  startTime: Date
  logId: number | null
  shouldStop: boolean
}

interface TaskProgress {
  totalFile: number
  generatedFile: number
  skipFile: number
}

type TaskProgressCallback = (progress: number, stats: TaskProgress) => void
type TaskCompleteCallback = (result: { success: boolean, result?: any }) => void

class TaskQueueService extends EventEmitter {
  private queue: TaskJob[] = []
  private runningJobs: Map<number, TaskJob> = new Map()
  private maxConcurrent = 20 // 最大并发任务数

  constructor() {
    super()
    // 监听任务进度更新
    this.on('taskProgress', (taskId: number, progress: number, stats: TaskProgress) => {
      logger.info.info('任务进度更新', { taskId, progress, stats })
    })
  }

  /**
   * 添加任务到队列
   */
  async addTask(taskId: number): Promise<boolean> {
    const task = await Task.findByPk(taskId)
    if (!task) {
      logger.warn.warn('添加任务失败: 任务不存在', { taskId })
      return false
    }

    if (task.running) {
      logger.warn.warn('添加任务失败: 任务正在运行', { taskId })
      return false
    }

    const job: TaskJob = {
      id: taskId,
      task,
      status: 'pending',
      progress: 0,
      startTime: new Date(),
      logId: null,
      shouldStop: false
    }

    this.queue.push(job)
    logger.info.info('任务已添加到队列', { taskId })
    this.processQueue()
    return true
  }

  /**
   * 处理队列
   */
  private async processQueue() {
    if (this.runningJobs.size >= this.maxConcurrent) {
      logger.debug.debug('达到最大并发数，等待处理', { 
        running: this.runningJobs.size,
        maxConcurrent: this.maxConcurrent 
      })
      return
    }

    const nextJob = this.queue.shift()
    if (!nextJob) {
      logger.debug.debug('队列为空，无待处理任务')
      return
    }

    try {
      // 创建任务日志
      const log = await taskLogService.create({
        taskId: nextJob.id,
        status: 'running',
        startTime: new Date(),
        totalFile: 0,
        generatedFile: 0,
        skipFile: 0
      })

      nextJob.logId = log.id
      nextJob.status = 'running'
      this.runningJobs.set(nextJob.id, nextJob)
      
      logger.info.info('开始执行任务', { 
        taskId: nextJob.id,
        logId: log.id
      })

      // 启动任务执行
      this.executeTask(nextJob).catch(error => {
        logger.error.error('任务执行失败:', error)
      })
    }
    catch (error) {
      logger.error.error('启动任务失败:', error)
      this.queue.unshift(nextJob) // 将任务放回队列
    }
  }

  /**
   * 执行任务
   */
  private async executeTask(job: TaskJob) {
    try {
      // 更新任务状态
      await Task.update(
        { running: true },
        { where: { id: job.id } }
      )

      // 执行文件生成
      const stats = await generatorService.generateStrm(
        job.task.dataValues.sourcePath,
        job.task.dataValues.targetPath,
        job.task.dataValues.fileSuffix.split(','),
        job.task.dataValues.overwrite
      )
      
      // 更新进度
      if (job.logId) {
        await taskLogService.update(job.logId, {
          totalFile: stats.totalFiles,
          generatedFile: stats.generatedFiles,
          skipFile: stats.skippedFiles
        })
      }
      
      // 完成任务
      await this.completeTask(job.id, true, stats)
    }
    catch (error) {
      logger.error.error('任务执行异常:', error)
      await this.completeTask(job.id, false, error)
    }
  }

  /**
   * 完成任务
   */
  private async completeTask(taskId: number, success: boolean, result?: any) {
    const job = this.runningJobs.get(taskId)
    if (!job) return

    try {
      // 更新任务状态
      await Task.update(
        { 
          running: false,
          lastRunAt: new Date()
        },
        { where: { id: taskId } }
      )

      // 更新日志
      if (job.logId) {
        await taskLogService.update(job.logId, {
          status: success ? 'completed' : 'failed',
          endTime: new Date(),
          message: result instanceof Error ? result.message : undefined,
          ...(success && result ? {
            totalFile: result.totalFile,
            generatedFile: result.generatedFile,
            skipFile: result.skipFile
          } : {})
        })
      }

      job.status = success ? 'completed' : 'failed'
      this.runningJobs.delete(taskId)
      this.emit('taskComplete', taskId, { success, result })
      
      logger.info.info('任务执行完成', { 
        taskId,
        success,
        result: result instanceof Error ? result.message : result
      })

      // 处理下一个任务
      this.processQueue()
    }
    catch (error) {
      logger.error.error('更新任务完成状态失败:', error)
    }
  }

  /**
   * 获取任务状态
   */
  getTaskStatus(taskId: number) {
    const job = this.runningJobs.get(taskId)
    if (job) {
      return {
        status: job.status,
        progress: job.progress,
        startTime: job.startTime
      }
    }
    return null
  }

  /**
   * 停止任务
   */
  async stopTask(taskId: number): Promise<boolean> {
    const job = this.runningJobs.get(taskId)
    if (!job) {
      logger.warn.warn('停止任务失败: 任务未在运行', { taskId })
      return false
    }

    job.shouldStop = true
    job.status = 'stopped'
    await this.completeTask(taskId, false, new Error('任务被手动停止'))
    return true
  }

  /**
   * 订阅任务进度
   */
  onProgress(taskId: number, callback: TaskProgressCallback) {
    this.on(`taskProgress`, (id: number, progress: number, stats: TaskProgress) => {
      if (id === taskId) {
        callback(progress, stats)
      }
    })
  }

  /**
   * 订阅任务完成
   */
  onComplete(taskId: number, callback: TaskCompleteCallback) {
    this.on(`taskComplete`, (id: number, result: any) => {
      if (id === taskId) {
        callback(result)
      }
    })
  }

  /**
   * 取消订阅
   */
  offProgress(taskId: number, callback: TaskProgressCallback) {
    this.removeListener(`taskProgress`, callback as (...args: any[]) => void)
  }

  /**
   * 取消订阅完成事件
   */
  offComplete(taskId: number, callback: TaskCompleteCallback) {
    this.removeListener(`taskComplete`, callback as (...args: any[]) => void)
  }
}

// 导出单例实例
export const taskQueue = new TaskQueueService() 