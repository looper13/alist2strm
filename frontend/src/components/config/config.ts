export interface ConfigField<T> {
  key: keyof T
  label: string
  type: 'text' | 'number' | 'boolean'
  required?: boolean
  placeholder?: string
  describe?: string
  min?: number
  step?: number
}

export interface ConfigItem<T> {
  name: string
  code: string
  fields: ConfigField<T>[]
}

// 配置默认值
export const defaultConfigs = {
  ALIST: {
    token: '',
    host: '',
    domain: '',
    reqInterval: 1000,
    reqRetryCount: 3,
    reqRetryInterval: 10000,
  } as Api.Config.AlistConfig,
  STRM: {
    defaultSuffix: 'mp4,mkv,avi,mov,rmvb,webm,flv,m3u8',
    replaceSuffix: true,
    urlEncode: true,
    minFileSize: 0,
  } as Api.Config.StrmConfig,
  EMBY: {
    embyServer: 'http://emby:8096',
    embyToken: '',
    pathMappings: [
      {
        path: '',
        embyPath: '',
      },
    ],
  } as Api.Config.EmbyConfig,
  NOTIFICATION_SETTINGS: {
    enabled: true,
    defaultChannel: 'telegram',
    channels: {
      telegram: {
        enabled: false,
        type: 'telegram',
        config: {
          botToken: '',
          chatId: '',
          parseMode: 'Markdown',
        },
      },
      wework: {
        enabled: false,
        type: 'wework',
        config: {
          corpId: '',
          agentId: '',
          corpSecret: '',
          toUser: '@all',
        },
      },
    },
    templates: {
      taskComplete: {
        telegram: '🎬 *任务完成通知* ✅\n\n📋 *基本信息*\n• *任务名称*: `{{.TaskName}}`\n• *完成时间*: {{.EventTime}}\n• *处理耗时*: {{.Duration}}秒\n\n📊 *处理统计*\n• *STRM文件*: 总计 {{.GeneratedFile}}+{{.SkipFile}}\n  - 已生成: {{.GeneratedFile}}\n  - 已跳过: {{.SkipFile}}\n• *元数据*: 总计 {{.MetadataCount}}\n  - 已下载: {{.MetadataDownloaded}}\n  - 已跳过: {{.MetadataSkipped}}\n• *字幕*: 总计 {{.SubtitleCount}}\n  - 已下载: {{.SubtitleDownloaded}}\n  - 已跳过: {{.SubtitleSkipped}}\n\n📁 *路径信息*\n• *源路径*: `{{.SourcePath}}`\n• *目标路径*: `{{.TargetPath}}`',
        wework: '# 🎬 任务完成通知 ✅\n\n## 📋 任务概览\n**任务名称**：<font color="info">`{{.TaskName}}`</font>\n**完成时间**：{{.EventTime}}\n**处理耗时**：<font color="info">{{.Duration}}</font> 秒\n\n## 📊 处理统计\n**STRM文件** (总计 {{.GeneratedFile}}+{{.SkipFile}})\n> 已生成：<font color="info">{{.GeneratedFile}}</font> | 已跳过：<font color="info">{{.SkipFile}}</font>\n\n**元数据文件** (总计 {{.MetadataCount}})\n> 已下载：<font color="info">{{.MetadataDownloaded}}</font> | 已跳过：<font color="info">{{.MetadataSkipped}}</font>\n\n**字幕文件** (总计 {{.SubtitleCount}})\n> 已下载：<font color="info">{{.SubtitleDownloaded}}</font> | 已跳过：<font color="info">{{.SubtitleSkipped}}</font>\n\n## 📂 路径信息\n**源路径**：`{{.SourcePath}}`\n**目标路径**：`{{.TargetPath}}`',
      },
      taskFailed: {
        telegram: '❌ *任务失败通知*\n\n📂 任务：`{{.TaskName}}`\n⏰ 时间：{{.EventTime}}\n⏱️ 耗时：{{.Duration}}秒\n❗ 错误信息：\n`{{.ErrorMessage}}`',
        wework: '❌ *任务失败通知*\n\n📂 任务：`{{.TaskName}}`\n⏰ 时间：{{.EventTime}}\n⏱️ 耗时：{{.Duration}}秒\n❗ 错误信息：\n`{{.ErrorMessage}}`',
      },
    },
    queueSettings: {
      maxRetries: 3,
      retryInterval: 60,
      concurrency: 1,
    },
  } as Api.Config.NotificationConfig,
}

// 配置项定义
export const CONFIG_ITEMS = [
  {
    name: 'Alist 配置',
    code: 'ALIST',
    fields: [
      {
        key: 'host',
        label: 'Alist 地址',
        type: 'text',
        required: true,
        placeholder: 'Alist 服务器地址，建议内网地址',
      },
      {
        key: 'token',
        label: 'AList Token',
        type: 'text',
        required: true,
        placeholder: 'Alist 访问令牌',
      },
      {
        key: 'domain',
        label: 'Alist 域名',
        type: 'text',
        required: true,
        placeholder: '替换 strm 文件中的域名',
        describe: '优先级高于 Alist 地址，建议外网域名或IP',
      },
      {
        key: 'reqInterval',
        label: '请求间隔(ms)',
        type: 'number',
        required: true,
        placeholder: '1000',
        min: 100,
        step: 100,
        describe: '每次请求之间的间隔时间，默认1000',
      },
      {
        key: 'reqRetryCount',
        label: '重试次数',
        type: 'number',
        required: true,
        placeholder: '3',
        min: 0,
        step: 1,
        describe: '请求失败时的重试次数，默认3次',
      },
      {
        key: 'reqRetryInterval',
        label: '重试间隔(ms)',
        type: 'number',
        required: true,
        placeholder: '10000',
        min: 100,
        step: 100,
        describe: '重试间隔时间，默认10000',
      },
    ] as ConfigField<Api.Config.AlistConfig>[],
  } as ConfigItem<Api.Config.AlistConfig>,
  {
    name: 'cloud drive 配置',
    code: 'CLOUD_DRIVE',
    fields: [
      {
        key: 'host',
        label: 'host',
        type: 'text',
        required: true,
        placeholder: 'host',
        describe: '例如：http://192.168.1.100:19798',
      },
      {
        key: 'username',
        label: '用户名',
        type: 'text',
        required: true,
        placeholder: '用户名',
        describe: '用户名',
      },
      {
        key: 'password',
        label: '密码',
        type: 'text',
        required: true,
        placeholder: '密码',
        describe: '密码',
      },
    ] as ConfigField<Api.Config.CloudDriveConfig>[],
  } as ConfigItem<Api.Config.CloudDriveConfig>,
  {
    name: 'strm 配置',
    code: 'STRM',
    fields: [
      {
        key: 'defaultSuffix',
        label: '默认后缀',
        type: 'text',
        required: true,
        placeholder: '支持文件后缀,多个逗号分隔',
        describe: '例如：mp4,mkv,avi',
      },
      {
        key: 'replaceSuffix',
        label: '替换后缀',
        type: 'boolean',
        describe: '未开启，xxx.mp4.strm，开启 xxx.strm',
      },
      {
        key: 'urlEncode',
        label: 'URL编码',
        type: 'boolean',
        describe: '对 strm 内容进行URL编码，建议开启',
      },
      {
        key: 'minFileSize',
        label: '小文件过滤(MB)',
        type: 'number',
        required: true,
        placeholder: '过滤小于指定大小的文件，0表示不过滤',
        min: 0,
        step: 10,
        describe: '如 200 表示过滤小于200MB的文件，0表示不过滤',
      },
    ] as ConfigField<Api.Config.StrmConfig>[],
  } as ConfigItem<Api.Config.StrmConfig>,
  {
    name: 'Emby 配置',
    code: 'EMBY',
    fields: [] as ConfigField<Api.Config.EmbyConfig>[],
  } as ConfigItem<Api.Config.EmbyConfig>,
  {
    name: '消息通知',
    code: 'NOTIFICATION_SETTINGS',
    fields: [] as ConfigField<Api.Config.NotificationConfig>[],
  } as ConfigItem<Api.Config.NotificationConfig>,
]
