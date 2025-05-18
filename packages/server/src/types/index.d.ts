/// <reference path="./models.d.ts" />
/// <reference path="./services.d.ts" />
/// <reference path="./api.d.ts" />
/// <reference path="./alist.d.ts" />
/// <reference path="./generate.d.ts" />

// 环境变量声明
declare namespace NodeJS {
  interface ProcessEnv {
    NODE_ENV?: 'development' | 'production'
    ALIST_HOST?: string
    ALIST_TOKEN?: string
    GENERATOR_PATH?: string
    GENERATOR_TARGET_PATH?: string
    GENERATOR_FILE_SUFFIX?: string
    CRON_EXPRESSION?: string
    CRON_ENABLE?: string
    PORT?: string
    LOG_BASE_DIR?: string
    LOG_APP_NAME?: string
    LOG_LEVEL?: 'info' | 'debug' | 'error' | 'warn'
    LOG_MAX_DAYS?: string
    LOG_MAX_FILE_SIZE?: string
    DB_PATH?: string
    DB_NAME?: string
  }
}

// 应用配置类型声明
declare namespace App {
  // 服务配置类型
  interface ServerConfig {
    port: number
  }
  // 日志配置类型
  interface LoggerConfig {
    baseDir: string
    appName: string
    level: 'info' | 'debug' | 'error' | 'warn'
    maxDays: number
    maxFileSize: number
  }
  // 数据库配置类型
  interface DatabaseConfig {
    path: string
    name: string
  }

  interface Config {
    server: ServerConfig
    logger: LoggerConfig
    database: DatabaseConfig
  }

  // HTTP 相关类型
  interface HttpResponse<T = any> {
    code: number
    message: string
    data?: T
  }

  interface PaginationQuery {
    page?: number
    pageSize?: number
    sortBy?: string
    sortOrder?: 'asc' | 'desc'
  }

  interface PaginationResponse<T> {
    items: T[]
    total: number
    page: number
    pageSize: number
    totalPages: number
  }

  // 错误处理相关类型
  interface AppError extends Error {
    code: number
    status?: number
    details?: Record<string, any>
  }

  // 数据库模型相关类型
  interface BaseModel {
    id: number
    createdAt: Date
    updatedAt: Date
  }
}

// 导出类型
export type Config = App.Config
export type HttpResponse<T = any> = App.HttpResponse<T>
export type PaginationQuery = App.PaginationQuery
export type PaginationResponse<T> = App.PaginationResponse<T>
export type AppError = App.AppError
export type BaseModel = App.BaseModel
export type AlistFile = AList.AlistFile

export type AlistListResponse<T> = AList.AlistListResponse<T>
export type AlistGetResponse<T> = AList.AlistGetResponse<T>
export type AlistFile = AList.AlistFile

export type GenerateResult = GenerateResult.GenerateResult
export type GenerateTask = GenerateResult.GenerateTask
export as namespace App
export = App 