import { HTTP_STATUS, API_CODE, ERROR_MSG, SUCCESS_MSG } from '../constants'

/**
 * 统一的API响应格式
 */
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

/**
 * 创建成功响应
 * @param data 响应数据
 * @param msg 成功消息
 * @param code 成功代码，默认200
 * @returns 标准格式的成功响应
 */
export function success<T = any>(
  data: T,
  msg = SUCCESS_MSG.OPERATION_SUCCESS,
  code = API_CODE.SUCCESS,
): ApiResponse<T> {
  return {
    code,
    msg,
    data,
  }
}

/**
 * 创建失败响应
 * @param msg 错误消息
 * @param code 错误代码，默认500
 * @param data 可选的错误相关数据
 * @returns 标准格式的失败响应
 */
export function fail<T = any>(
  msg = ERROR_MSG.INTERNAL_ERROR,
  code = API_CODE.INTERNAL_ERROR,
  data: T = null as any,
): ApiResponse<T> {
  return {
    code,
    msg,
    data,
  }
}

/**
 * 特定状态码的失败响应
 */
export const errorResponse = {
  badRequest: (msg = ERROR_MSG.PARAM_ERROR) => fail(msg, API_CODE.PARAM_ERROR),

  unauthorized: (msg = ERROR_MSG.UNAUTHORIZED) => fail(msg, API_CODE.UNAUTHORIZED),

  forbidden: (msg = ERROR_MSG.FORBIDDEN) => fail(msg, API_CODE.FORBIDDEN),

  notFound: (msg = ERROR_MSG.NOT_FOUND) => fail(msg, API_CODE.NOT_FOUND),

  conflict: (msg = ERROR_MSG.NOT_FOUND) => fail(msg, API_CODE.CONFLICT),

  serverError: (msg = ERROR_MSG.INTERNAL_ERROR) => fail(msg, API_CODE.INTERNAL_ERROR),
}

export default {
  success,
  fail,
  errorResponse,
}
