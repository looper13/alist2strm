import type { Request, Response, NextFunction } from 'express'
import { logger } from '@/utils/logger.js'
import { error } from '@/utils/response.js'

export class HttpError extends Error implements App.AppError {
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
  logger.error.error('请求错误:', {
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
  error(res, errorMessage, statusCode)
}

export function notFoundHandler(req: Request, res: Response): void {
  logger.warn.warn('资源不存在:', {
    path: req.path,
    method: req.method,
    query: req.query,
  })

  error(res, '资源不存在', 404)
} 