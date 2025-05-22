import type { Response } from 'express'
import type { API } from '@/types/api.js'

/**
 * 成功响应
 */
export function success<T>(res: Response, data: T, message = '操作成功'): void {
  const response: API.SuccessResponse<T> = {
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
  const response: API.ErrorResponse = {
    code,
    message,
  }
  res.status(code).json(response)
}

/**
 * 分页响应
 */
export function pageResult<T>(res: Response, data: {
  list: T[]
  total: number
  page: number
  pageSize: number
}, message = '查询成功'): void {
  const response: API.PageResponse<T> = {
    code: 0,
    data,
    message,
  }
  res.json(response)
} 