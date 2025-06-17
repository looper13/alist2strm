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

    interface UserInfo {
      id: number
      username: string
      nickname?: string
      email?: string
      status: 'active' | 'disabled'
      lastLoginAt?: string
    }

    // 用户信息更新参数
    interface UpdateUserParams {
      nickname?: string
      password?: string
      oldPassword?: string
    }
  }

  // 系统配置
  namespace Config {
    // 配置相关类型
    type Record = Common.CommonRecord<{
      name: string
      code: string
      value: string
    }>

    type Create = Pick<Config, 'name' | 'code' | 'value'>
    type Update = Pick<Config, 'name' | 'code' | 'value'>

  }

  // 任务
  namespace Task {
    type Record = Common.CommonRecord<{
      name: string
      mediaType: 'movie' | 'tv'
      sourcePath: string
      targetPath: string
      fileSuffix: string
      overwrite: boolean
      enabled: boolean
      cron?: string
      running: boolean
      lastRunAt?: string

      downloadMetadata: boolean
      metadataExtensions: string
      downloadSubtitle: boolean
      subtitleExtensions: string
    }>

    type Query = Pick<Record, 'name' | 'enabled' | 'overwrite'>
    type Update = Pick<Record, 'id' | 'name' | 'sourcePath' | 'targetPath' | 'fileSuffix' | 'overwrite' | 'enabled' | 'cron' | 'mediaType' | 'downloadMetadata' | 'metadataExtensions' | 'downloadSubtitle' | 'subtitleExtensions'>
    type Create = Pick<Record, 'name' | 'sourcePath' | 'targetPath' | 'fileSuffix' | 'overwrite' | 'enabled' | 'cron' | 'mediaType' | 'downloadMetadata' | 'metadataExtensions' | 'downloadSubtitle' | 'subtitleExtensions'>

    type Log = Common.CommonRecord<{
      taskId: number
      createdAt: string
      updatedAt: string
      message: string
      startTime: string
      endTime: string | null
      duration: number
      status: string
      totalFile: number
      generatedFile: number
      skipFile: number
      overwriteFile: number
      metadataCount: number
      subtitleCount: number
      failedCount: number
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
