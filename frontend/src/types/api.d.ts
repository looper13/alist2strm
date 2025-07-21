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
      newPassword?: string
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

    type Create = Pick<Record, 'name' | 'code' | 'value'>
    type Update = Pick<Record, 'name' | 'code' | 'value'>

    // STRM 特定配置类型
    interface StrmConfig {
      defaultSuffix: string
      replaceSuffix: boolean
      urlEncode: boolean
    }

    // Alist 特定配置类型
    export interface AlistConfig {
      token: string
      host: string
      domain: string
      reqInterval: number
      reqRetryCount: number
      reqRetryInterval: number
    }

    // clouddrive配置
    export interface CloudDriveConfig {
      host: string
      username: string
      password: string
    }

    // Emby 特定配置类型
    export interface PathMapping {
      path: string
      embyPath: string
    }

    export interface EmbyConfig {
      embyServer: string
      embyToken: string
      pathMappings: PathMapping[]
    }

    // 通知配置类型
    export interface ChannelConfig {
      enabled: boolean
      type: string
      config: Record<string, string>
    }

    export interface TemplateConfig {
      telegram: string
      wework: string
    }

    export interface QueueSettings {
      maxRetries: number
      retryInterval: number
      concurrency: number
    }

    export interface NotificationConfig {
      enabled: boolean
      defaultChannel: string
      channels: Record<string, ChannelConfig>
      templates: Record<string, TemplateConfig>
      queueSettings: QueueSettings
    }
  }

  // 任务
  namespace Task {
    type Record = Common.CommonRecord<{
      name: string
      mediaType: 'movie' | 'tv'
      configType: 'alist' | 'clouddrive' | 'local'
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
    type Update = Pick<Record, 'id' | 'configType' | 'name' | 'sourcePath' | 'targetPath' | 'fileSuffix' | 'overwrite' | 'enabled' | 'cron' | 'mediaType' | 'downloadMetadata' | 'metadataExtensions' | 'downloadSubtitle' | 'subtitleExtensions'>
    type Create = Pick<Record, 'configType' | 'name' | 'sourcePath' | 'targetPath' | 'fileSuffix' | 'overwrite' | 'enabled' | 'cron' | 'mediaType' | 'downloadMetadata' | 'metadataExtensions' | 'downloadSubtitle' | 'subtitleExtensions'>

    // 任务统计数据接口
    interface Stats {
      total: number // 任务总数
      enabled: number // 启用任务数
      disabled: number // 禁用任务数
      totalExecutions: number // 总执行次数
      successCount: number // 成功次数
      failedCount: number // 失败次数
    }

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

  // 任务日志相关类型
  namespace TaskLog {
    // 文件处理统计数据接口
    interface FileProcessingStats {
      totalFiles: number // 扫描的文件总数
      processedFiles: number // 已处理的文件数
      skippedFiles: number // 跳过处理的文件数
      strmGenerated: number // 生成的STRM文件数
      metadataDownloaded: number // 下载的元数据文件数
      subtitleDownloaded: number // 下载的字幕文件数
    }
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
