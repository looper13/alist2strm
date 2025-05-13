import { Model, Optional } from 'sequelize'
import { TASK_STATUS } from '../constants'

/**
 * 任务属性
 */
export interface TaskAttributes {
  id: number
  name: string
  sourcePath: string
  targetPath: string
  fileSuffix: string
  overwrite: boolean
  enabled: boolean
  cronExpression: string | null
  lastRunAt: Date | null
  running: boolean
  createdAt: Date
  updatedAt: Date
}

/**
 * 任务创建属性
 */
export interface TaskCreationAttributes
  extends Optional<TaskAttributes, 'id' | 'createdAt' | 'updatedAt'> {}

/**
 * 任务状态类型
 */
export type TaskStatusType =
  | typeof TASK_STATUS.PENDING
  | typeof TASK_STATUS.SUCCESS
  | typeof TASK_STATUS.ERROR

/**
 * 任务日志属性
 */
export interface TaskLogAttributes {
  id: number
  taskId: number
  status: TaskStatusType
  startTime: Date
  endTime: Date | null
  totalFiles: number | null
  generatedFiles: number | null
  skippedFiles: number | null
  error: string | null
  createdAt: Date
  updatedAt: Date
}

export interface TaskLogCreationAttributes
  extends Optional<TaskLogAttributes, 'id' | 'createdAt' | 'updatedAt'> {}

export interface TaskInstance
  extends Model<TaskAttributes, TaskCreationAttributes>,
    TaskAttributes {}
export interface TaskLogInstance
  extends Model<TaskLogAttributes, TaskLogCreationAttributes>,
    TaskLogAttributes {}

/**
 * 生成结果
 */
export interface GenerateResult {
  totalFiles: number
  generatedFiles: number
  skippedFiles: number
}

/**
 * 日志配置
 */
export interface LoggerConfig {
  baseDir: string
  appName: string
  level: 'debug' | 'info' | 'warn' | 'error'
  maxDays: number
  maxFileSize: number
}

/**
 * 数据库配置
 */
export interface DatabaseConfig {
  path: string
  name: string
}

/**
 * 配置
 */
export interface Config {
  alist: {
    host: string
    token: string
  }
  generator: {
    path: string
    targetPath: string
    fileSuffix: string[]
  }
  cron: {
    expression: string
    enable: boolean
  }
  server: {
    port: number
  }
  logger: LoggerConfig
  database: DatabaseConfig
}

export interface AlistStorage {
  id: number
  mount_path: string
  enable_sign: boolean
  provider: string
}
