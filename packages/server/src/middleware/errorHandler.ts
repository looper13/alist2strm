import type { ErrorRequestHandler } from 'express'
import { logger } from '../utils/logger'

export class AppError extends Error {
  constructor(
    public statusCode: number,
    message: string,
  ) {
    super(message)
    this.name = 'AppError'
  }
}

export const errorHandler: ErrorRequestHandler = (err, req, res, _next) => {
  const errorContext = {
    path: req.path,
    method: req.method,
    query: req.query,
    body: req.body,
    headers: req.headers,
    error: {
      name: err.name,
      message: err.message,
      stack: err.stack,
    },
  }

  if (err instanceof AppError) {
    logger.error.error('Application error occurred', {
      ...errorContext,
      statusCode: err.statusCode,
    })

    return res.status(err.statusCode).json({
      status: 'error',
      message: err.message,
    })
  }

  // Sequelize validation errors
  if (err.name === 'SequelizeValidationError') {
    const validationErrors = err.errors.map((e: any) => ({
      field: e.path,
      message: e.message,
    }))

    logger.error.error('Validation error occurred', {
      ...errorContext,
      validationErrors,
    })

    return res.status(400).json({
      status: 'error',
      message: '数据验证失败',
      errors: validationErrors,
    })
  }

  // Sequelize unique constraint errors
  if (err.name === 'SequelizeUniqueConstraintError') {
    const constraintErrors = err.errors.map((e: any) => ({
      field: e.path,
      message: e.message,
    }))

    logger.error.error('Unique constraint violation', {
      ...errorContext,
      constraintErrors,
    })

    return res.status(400).json({
      status: 'error',
      message: '数据已存在',
      errors: constraintErrors,
    })
  }

  // Default error
  logger.error.error('Unhandled error occurred', errorContext)

  return res.status(500).json({
    status: 'error',
    message: '服务器内部错误',
  })
}
