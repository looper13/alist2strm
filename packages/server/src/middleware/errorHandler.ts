import type { ErrorRequestHandler } from 'express'
import { logger } from '../utils/logger'
import { fail } from '../utils/response'
import { HTTP_STATUS, API_CODE, ERROR_MSG } from '../constants'

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

    return res.status(err.statusCode).json(fail(err.message, err.statusCode))
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

    return res
      .status(HTTP_STATUS.BAD_REQUEST)
      .json(fail(ERROR_MSG.PARAM_ERROR, API_CODE.PARAM_ERROR, { errors: validationErrors }))
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

    return res
      .status(HTTP_STATUS.BAD_REQUEST)
      .json(fail('数据已存在', API_CODE.CONFLICT, { errors: constraintErrors }))
  }

  // Default error
  logger.error.error('Unhandled error occurred', errorContext)

  return res
    .status(HTTP_STATUS.INTERNAL_SERVER_ERROR)
    .json(fail(ERROR_MSG.INTERNAL_ERROR, API_CODE.INTERNAL_ERROR))
}
