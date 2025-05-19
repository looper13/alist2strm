// HTTP 相关类型
declare namespace Api {
  // 基础响应类型
  interface HttpResponse<T = any> {
    code: number
    message: string
    data?: T
  }

  // 分页查询参数
  interface PaginationQuery {
    page?: number
    pageSize?: number
    sortBy?: string
    sortOrder?: 'asc' | 'desc'
  }

  // 分页响应类型
  interface PaginationResponse<T> {
    list: T[]
    total: number
    page: number
    pageSize: number
  }

  // 配置相关类型
  interface Config {
    id: number
    name: string
    code: string
    value: string
    description?: string
    createdAt: string
    updatedAt: string
  }

  interface ConfigCreateDto extends Omit<Config, 'id' | 'createdAt' | 'updatedAt'> {}
  interface ConfigUpdateDto extends Omit<Config, 'id' | 'createdAt' | 'updatedAt'> {}

  // 任务相关类型
  interface Task {
    id: number
    name: string
    sourcePath: string
    targetPath: string
    fileSuffix: string
    overwrite: boolean
    enabled: boolean
    cron: string
    running: boolean
    lastRunAt?: string
    createdAt: string
    updatedAt: string
  }

  interface TaskCreateDto extends Omit<Task, 'id' | 'createdAt' | 'updatedAt'> {}
  interface TaskUpdateDto extends Omit<Task, 'id' | 'createdAt' | 'updatedAt'> {}

  // 任务日志相关类型
  interface TaskLog {
    id: number
    taskId: number
    status: string
    message: string
    startTime: string
    endTime: string | null
    totalFile: number
    generatedFile: number
    skipFile: number
    createdAt: string
    updatedAt: string
  }

  // 文件历史相关类型
  interface FileHistory {
    id: number
    fileName: string
    sourcePath: string
    targetFilePath: string
    fileSize: number
    fileType: string
    fileSuffix: string
    createdAt: string
    updatedAt: string
  }

  interface FileHistoryQuery extends PaginationQuery {
    keyword?: string
    fileType?: string
    fileSuffix?: string
    startTime?: string
    endTime?: string
  }

  // 任务进度相关类型
  interface TaskProgress {
    progress: number
    status: 'running' | 'stopped' | 'completed' | 'failed'
  }

  interface AlistDir {
    name: string
    modified: string
  }

}
