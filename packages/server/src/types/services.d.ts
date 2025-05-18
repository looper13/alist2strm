declare namespace Services {
  interface CreateConfigDto {
    name: string
    code: string
    value: string
  }

  interface UpdateConfigDto {
    name?: string
    code?: string
    value?: string
  }

  interface QueryConfigDto {
    page?: number
    pageSize?: number
    keyword?: string
  }

  interface PageResult<T> {
    list: T[]
    total: number
    page: number
    pageSize: number
  }

  interface CreateTaskDto {
    name: string
    sourcePath: string
    targetPath: string
    fileSuffix: string
    overwrite?: boolean
    enabled?: boolean
    cron?: string
  }

  interface UpdateTaskDto {
    name?: string
    sourcePath?: string
    targetPath?: string
    fileSuffix?: string
    overwrite?: boolean
    enabled?: boolean
    cron?: string
  }

  interface QueryTaskDto {
    page?: number
    pageSize?: number
    keyword?: string
    enabled?: boolean
    running?: boolean
  }
} 