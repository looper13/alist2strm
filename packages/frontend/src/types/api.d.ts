// HTTP 相关类型
declare namespace Api {

  namespace Common {
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

    // 分页响应类型
    interface PaginationResponse<T> {
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

  // 授权
  namespace Auth {
    interface LoginParams {
      username: string
      password: string
    }
    interface RegisterParams {
      username: string
      password: string
      nickname?: string
    }

    interface LoginResult {
      token: string
      user: {
        id: number
        username: string
        nickname?: string
        email?: string
        status: 'active' | 'disabled'
        lastLoginAt?: string
        createdAt: string
        updatedAt: string
      }
    }
  }

  // 系统配置
  namespace Config {
    // 配置相关类型
    type Record = Common.CommonRecord<{
      name: string
      code: string
      value: string
      description?: strin
    }>

    type Create = Pick<Config, 'name' | 'code' | 'value' | 'description'>
    type Update = Pick<Config, 'id' | 'name' | 'code' | 'value' | 'description'>

  }

  // 任务
  namespace Task {
    type Record = Common.CommonRecord<{
      name: string
      sourcePath: string
      targetPath: string
      fileSuffix: string
      overwrite: boolean
      enabled: boolean
      cron?: string
      running: boolean
      lastRunAt?: string
    }>

    type Query = Pick<Record, 'name' | 'enabled' | 'overwrite'>
    type Update = Pick<Record, 'id' | 'name' | 'sourcePath' | 'targetPath' | 'fileSuffix' | 'overwrite' | 'enabled' | 'cron'>
    type Create = Pick<Record, 'name' | 'sourcePath' | 'targetPath' | 'fileSuffix' | 'overwrite' | 'enabled' | 'cron'>

    type Log = Common.CommonRecord<{
      taskId: number
      status: string
      message: string
      startTime: string
      endTime: string | null
      totalFile: number
      generatedFile: number
      skipFile: number
    }>

    type LogQuery = Common.PaginationQuery<{
      taskId?: number
    }>

  }

  // 文件历史
  namespace FileHistory {
    type Record = Common.CommonRecord<{
      fileName: string
      sourcePath: string
      targetFilePath: string
      fileSize: number
      fileType: string
      fileSuffix: string
    }>

    type Query = Common.PaginationQuery<{
      keyword?: string
      fileType?: string
      fileSuffix?: string
      startTime?: string
      endTime?: string
    }>
  }
}
