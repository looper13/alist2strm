import type { ErrorRequestHandler } from 'express'

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
  console.error('Error:', err)

  if (err instanceof AppError) {
    return res.status(err.statusCode).json({
      status: 'error',
      message: err.message,
    })
  }

  // Sequelize validation errors
  if (err.name === 'SequelizeValidationError') {
    return res.status(400).json({
      status: 'error',
      message: '数据验证失败',
      errors: err.errors.map((e: any) => ({
        field: e.path,
        message: e.message,
      })),
    })
  }

  // Sequelize unique constraint errors
  if (err.name === 'SequelizeUniqueConstraintError') {
    return res.status(400).json({
      status: 'error',
      message: '数据已存在',
      errors: err.errors.map((e: any) => ({
        field: e.path,
        message: e.message,
      })),
    })
  }

  // Default error
  return res.status(500).json({
    status: 'error',
    message: '服务器内部错误',
  })
}
