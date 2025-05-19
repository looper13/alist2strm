import type { Request, Response, NextFunction } from 'express'
import { logger } from '../utils/logger.js'
import type { AppError } from '../types/index.js'

export class HttpError extends Error implements AppError {
  constructor(
    public message: string,
    public code: number,
    public status: number = code,
    public details?: Record<string, any>,
  ) {
    super(message)
    this.name = 'HttpError'
  }
}

export function errorHandler(
  err: Error | HttpError,
  req: Request,
  res: Response,
  _next: NextFunction,
): void {
  const isHttpError = err instanceof HttpError
  const statusCode = isHttpError ? err.status : 500
  const errorCode = isHttpError ? err.code : 500
  const errorMessage = isHttpError ? err.message : 'Internal server error'
  const details = isHttpError ? err.details : undefined

  // 记录错误日志
  logger.error.error('Request error:', {
    path: req.path,
    method: req.method,
    query: req.query,
    body: req.body,
    error: {
      name: err.name,
      message: err.message,
      code: errorCode,
      status: statusCode,
      details,
      stack: err.stack,
    },
  })

  res.status(statusCode).json({
    code: errorCode,
    message: errorMessage,
    details,
  })
}

export function notFoundHandler(req: Request, res: Response): void {
  logger.warn.warn('Resource not found:', {
    path: req.path,
    method: req.method,
    query: req.query,
  })

  res.status(404).json({
    code: 404,
    message: 'Resource not found',
  })
} 