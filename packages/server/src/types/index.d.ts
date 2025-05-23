/// <reference path="./models.d.ts" />

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
    DB_BASE_DIR?: string
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

  // 错误处理相关类型
  interface AppError extends Error {
    code: number
    status?: number
    details?: Record<string, any>
  }

  namespace Common {
    interface BaseModel {
      id: number
      createdAt?: Date
      updatedAt?: Date
    }

    // 分页查询参数
    interface PaginationQuery {
      page?: number
      pageSize?: number
      sortBy?: string
      sortOrder?: 'asc' | 'desc'
    }

    // 基础响应类型
    interface HttpResponse<T = any> {
      code: number
      message: string
      data?: T
    }

    // 分页数据结果
    interface PaginationResult<T> {
      list: T[]
      total: number
      page: number
      pageSize: number
    }

    type CommonRecord<T = any> = {
      id: number
      createdAt: string
      updatedAt: string
    } & T
  }

  namespace Config {
    interface Record extends Common.BaseModel {
      name: string
      code: string
      value: string
    }

    type Create = Omit<Record, 'id'>
    type Update = Pick<Record, 'id' | 'name' | 'code' | 'value'>
    type Query = Common.PaginationQuery<{
      keyword?: string
    }>
  }

  namespace Task {
    interface Record extends Common.BaseModel {
      name: string
      sourcePath: string
      targetPath: string
      fileSuffix: string
      overwrite?: boolean
      enabled?: boolean
      cron?: string
    }

    type Create = Omit<Record, 'id'| 'createdAt' | 'updatedAt'>
    type Update = Partial<Record, 'id' | 'sourcePath' | 'targetPath' | 'fileSuffix'>
    type Query = Common.PaginationQuery<{
      keyword?: string
      enabled?: boolean
      running?: boolean
    }>

    interface Log extends Common.BaseModel {
      taskId: number
      status: string
      message?: string | null
      startTime?: Date
      endTime?: Date | null
      totalFile?: number
      generatedFile?: number
      skipFile?: number
    }

    type LogCreate = Omit<Log, 'id' | 'createdAt' | 'updatedAt'>
    type LogUpdate = Partial<Log, 'id' | 'status' | 'taskId'>
    type LogQuery = Common.PaginationQuery<{
      taskId: string,
      status?: string
      startTime?: string
      endTime?: string
    }>
  }

  namespace FileHistory {
    type Record = Common.CommonRecord<{
      fileName: string
      sourcePath: string
      targetFilePath: string
      fileSize: number
      fileType: string
      fileSuffix: string
    }>

    type Create = Omit<Record, 'id' | 'createdAt' | 'updatedAt' >
    type Update = Partial<Record, 'id' | 'createdAt' | 'updatedAt' >
    type Query = Common.PaginationQuery<{
      keyword?: string
      fileType?: string
      fileSuffix?: string
      startTime?: string
      endTime?: string
    }>
  }

  namespace Generate {
    /**
     * 生成结果
     */
    interface Result {
      success: boolean
      message: string
      totalFiles: number
      generatedFiles: number
      skippedFiles: number
    }

    /**
     * strm 成任务
     */
    interface Task {
      sourceFilePath: string
      targetFilePath: string
      strmPath: string
      name: string
      sign?: string
      type: string
      fileSize: number
    }
  }

  namespace AList {
     interface AlistListResponse<T> {
      code: number
      message: string
      data: {
        content: T
      }
    }
    
    interface AlistGetResponse<T> {
      code: number
      message: string
      data: T
    }

    interface AlistFile {
      name: string
      size?: number
      is_dir: boolean
      modified?: string
      created?: string
      sign?: string
      thumb?: string
      type?: number
      hashinfo?: string
      hash_info?: any
      raw_url?: string
      readme?: string
      header?: string
      provider?: string
      related?: any
    }

    interface AlistDir {
      name: string
      modified: string
    }
  }
}