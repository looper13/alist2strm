import type { Response } from 'express'

/**
 * 成功响应
 */
export function success<T>(res: Response, data: T, message = '操作成功'): void {
  const response: App.Common.HttpResponse<T> = {
    code: 0,
    message,
    data,
  }
  res.json(response)
}

/**
 * 错误响应
 */
export function error(res: Response, message: string, code = 500): void {
  const response: App.Common.HttpResponse = {
    code,
    message,
  }
  res.status(code).json(response)
}

/**
 * 分页响应
 */
export function pageResult<T>(res: Response, data: App.Common.PaginationResult<T>, message = '查询成功'): void {
  const response: App.Common.HttpResponse<App.Common.PaginationResult<T>> = {
    code: 0,
    data,
    message,
  }
  res.json(response)
} 