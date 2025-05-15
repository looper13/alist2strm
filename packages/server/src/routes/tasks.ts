import { Router } from 'express'
import { AppError } from '../middleware/errorHandler'
import taskService from '../services/taskService'
import { logger } from '../utils/logger'
import { success, fail } from '../utils/response'
import { HTTP_STATUS, API_CODE, ERROR_MSG, SUCCESS_MSG } from '../constants'

const router = Router()

/**
 * 获取任务列表
 */
router.get('/', async (req, res, next) => {
  try {
    const page = parseInt(req.query.page as string) || 1
    const pageSize = parseInt(req.query.pageSize as string) || 10

    logger.info.info('正在获取任务列表', {
      query: req.query,
      page,
      pageSize,
    })

    const result = await taskService.getTasksWithPagination(page, pageSize)

    logger.debug.debug('成功获取任务列表', {
      page,
      pageSize,
      total: result.total,
      currentPageSize: result.records.length,
      requestQuery: req.query,
    })

    res.json(success(result, SUCCESS_MSG.TASK_LIST_SUCCESS))
  } catch (error) {
    logger.error.error('获取任务列表失败', {
      error: (error as Error).message,
      stack: (error as Error).stack,
      query: req.query,
    })
    next(error)
  }
})

/**
 * 获取单个任务
 */
router.get('/:id', async (req, res, next) => {
  const taskId = Number(req.params.id)
  try {
    logger.info.info('正在获取任务详情', { taskId })

    const task = await taskService.getTask(taskId)
    if (!task) {
      logger.warn.warn('未找到任务', { taskId })
      throw new AppError(HTTP_STATUS.NOT_FOUND, ERROR_MSG.TASK_NOT_FOUND)
    }

    logger.debug.debug('成功获取任务详情', { taskId, task })
    res.json(success(task, SUCCESS_MSG.TASK_DETAIL_SUCCESS))
  } catch (error) {
    logger.error.error('获取任务详情失败', {
      taskId,
      error: (error as Error).message,
      stack: (error as Error).stack,
    })
    next(error)
  }
})

/**
 * 创建任务
 */
router.post('/', async (req, res, next) => {
  try {
    logger.info.info('正在创建新任务', {
      taskData: req.body,
    })

    const task = await taskService.createTask(req.body)

    logger.debug.debug('成功创建任务', {
      taskId: task.id,
      taskData: task,
    })

    res
      .status(HTTP_STATUS.CREATED)
      .json(success(task, SUCCESS_MSG.TASK_CREATE_SUCCESS, API_CODE.SUCCESS))
  } catch (error) {
    logger.error.error('创建任务失败', {
      taskData: req.body,
      error: (error as Error).message,
      stack: (error as Error).stack,
    })
    next(error)
  }
})

/**
 * 更新任务
 */
router.put('/:id', async (req, res, next) => {
  const taskId = Number(req.params.id)
  try {
    logger.info.info('正在更新任务', {
      taskId,
      updates: req.body,
    })

    const task = await taskService.updateTask(taskId, req.body)
    if (!task) {
      logger.warn.warn('未找到要更新的任务', { taskId })
      throw new AppError(HTTP_STATUS.NOT_FOUND, ERROR_MSG.TASK_NOT_FOUND)
    }

    logger.debug.debug('成功更新任务', {
      taskId,
      updatedTask: task,
    })

    res.json(success(task, SUCCESS_MSG.TASK_UPDATE_SUCCESS))
  } catch (error) {
    logger.error.error('更新任务失败', {
      taskId,
      updates: req.body,
      error: (error as Error).message,
      stack: (error as Error).stack,
    })
    next(error)
  }
})

/**
 * 删除任务
 */
router.delete('/:id', async (req, res, next) => {
  const taskId = Number(req.params.id)
  try {
    logger.info.info('正在删除任务', { taskId })

    await taskService.deleteTask(taskId)

    logger.debug.debug('成功删除任务', { taskId })

    res.status(HTTP_STATUS.OK).json(success(null, SUCCESS_MSG.TASK_DELETE_SUCCESS))
  } catch (error) {
    logger.error.error('删除任务失败', {
      taskId,
      error: (error as Error).message,
      stack: (error as Error).stack,
    })
    next(error)
  }
})

/**
 * 执行任务
 */
router.post('/:id/execute', async (req, res, next) => {
  const taskId = Number(req.params.id)
  const startTime = Date.now()

  try {
    logger.info.info('正在执行任务', {
      taskId,
      params: req.body,
    })

    const result = await taskService.executeTask(taskId)

    // 检查任务是否已在执行中
    if (result && typeof result === 'object' && 'success' in result && result.success === false) {
      logger.info.info('任务已在执行中', {
        taskId,
        message: result.message,
        executionTime: Date.now() - startTime,
      })

      return res.status(HTTP_STATUS.ACCEPTED).json(
        success(
          {
            alreadyRunning: true,
            taskId: result.taskId,
            taskName: result.taskName,
          },
          ERROR_MSG.TASK_RUNNING,
          API_CODE.TASK_RUNNING,
        ),
      )
    }

    logger.info.info('成功执行任务', {
      taskId,
      result,
      executionTime: Date.now() - startTime,
    })

    res.json(success(result, SUCCESS_MSG.TASK_EXECUTE_SUCCESS))
  } catch (error) {
    logger.error.error('执行任务失败', {
      taskId,
      error: (error as Error).message,
      stack: (error as Error).stack,
      executionTime: Date.now() - startTime,
    })
    next(error)
  }
})

/**
 * 获取任务日志
 */
router.get('/:id/logs', async (req, res, next) => {
  const taskId = Number(req.params.id)
  try {
    logger.info.info('正在获取任务日志', {
      taskId,
      query: req.query,
    })

    const limit = req.query.limit ? Number(req.query.limit) : undefined
    const logs = await taskService.getTaskLogs(taskId, limit)

    logger.debug.debug('成功获取任务日志', {
      taskId,
      logCount: logs.length,
    })

    res.json(success(logs, SUCCESS_MSG.TASK_LOG_SUCCESS))
  } catch (error) {
    logger.error.error('获取任务日志失败', {
      taskId,
      error: (error as Error).message,
      stack: (error as Error).stack,
    })
    next(error)
  }
})

export default router
