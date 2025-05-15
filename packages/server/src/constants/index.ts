/**
 * HTTP状态码
 */
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  ACCEPTED: 202,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  INTERNAL_SERVER_ERROR: 500,
}

/**
 * 业务状态码
 */
export const API_CODE = {
  SUCCESS: 200, // 成功
  PARAM_ERROR: 400, // 参数错误
  UNAUTHORIZED: 401, // 未授权
  FORBIDDEN: 403, // 禁止访问
  NOT_FOUND: 404, // 资源不存在
  CONFLICT: 409, // 资源冲突
  INTERNAL_ERROR: 500, // 服务器内部错误
  TASK_RUNNING: 1001, // 任务已在运行中
  DB_ERROR: 1002, // 数据库错误
}

/**
 * 错误消息
 */
export const ERROR_MSG = {
  PARAM_ERROR: '请求参数错误',
  NOT_FOUND: '资源不存在',
  UNAUTHORIZED: '未授权',
  FORBIDDEN: '禁止访问',
  INTERNAL_ERROR: '服务器内部错误',
  TASK_NOT_FOUND: '任务不存在',
  TASK_RUNNING: '任务已在执行中',
  DB_ERROR: '数据库错误',
  DIR_NOT_FOUND: '目录不存在',
  FILE_NOT_FOUND: '文件不存在',
}

/**
 * 成功消息
 */
export const SUCCESS_MSG = {
  OPERATION_SUCCESS: '操作成功',
  CREATE_SUCCESS: '创建成功',
  UPDATE_SUCCESS: '更新成功',
  DELETE_SUCCESS: '删除成功',
  TASK_CREATE_SUCCESS: '任务创建成功',
  TASK_UPDATE_SUCCESS: '更新任务成功',
  TASK_DELETE_SUCCESS: '删除任务成功',
  TASK_EXECUTE_SUCCESS: '执行任务成功',
  TASK_LIST_SUCCESS: '获取任务列表成功',
  TASK_DETAIL_SUCCESS: '获取任务成功',
  TASK_LOG_SUCCESS: '获取任务日志成功',

  FOLDER_LIST_SUCCESS: '获取目录列表成功',
}

/**
 * 任务状态
 */
export const TASK_STATUS = {
  PENDING: 'pending',
  SUCCESS: 'success',
  ERROR: 'error',
}

export default {
  HTTP_STATUS,
  API_CODE,
  ERROR_MSG,
  SUCCESS_MSG,
  TASK_STATUS,
}
