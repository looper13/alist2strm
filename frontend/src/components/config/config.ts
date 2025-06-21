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
  } as Api.Config.StrmConfig,
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
        telegram: '✅ *任务完成通知*\n\n📂 任务：`{{.TaskName}}`\n⏱️ 耗时：{{.Duration}}秒\n📊 处理结果：\n - 总文件：{{.TotalFiles}}个\n - 已生成：{{.GeneratedFiles}}个\n - 已跳过：{{.SkippedFiles}}个\n - 元数据：{{.MetadataFiles}}个\n - 字幕：{{.SubtitleFiles}}个',
        wework: '【任务完成通知】\n\n任务：{{.TaskName}}\n耗时：{{.Duration}}秒\n处理结果：\n- 总文件：{{.TotalFiles}}个\n- 已生成：{{.GeneratedFiles}}个\n- 已跳过：{{.SkippedFiles}}个\n- 元数据：{{.MetadataFiles}}个\n- 字幕：{{.SubtitleFiles}}个',
      },
      taskFailed: {
        telegram: '❌ *任务失败通知*\n\n📂 任务：`{{.TaskName}}`\n⏱️ 耗时：{{.Duration}}秒\n❗ 错误信息：\n`{{.ErrorMessage}}`',
        wework: '【任务失败通知】\n\n任务：{{.TaskName}}\n耗时：{{.Duration}}秒\n错误信息：\n{{.ErrorMessage}}',
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
        min: 100,
        step: 100,
        describe: '每次请求之间的间隔时间，默认100',
      },
      {
        key: 'reqRetryCount',
        label: '重试次数',
        type: 'number',
        required: true,
        min: 0,
        step: 1,
        describe: '请求失败时的重试次数，默认3次',
      },
      {
        key: 'reqRetryInterval',
        label: '重试间隔(ms)',
        type: 'number',
        required: true,
        min: 100,
        step: 100,
        describe: '重试间隔时间，默认10000',
      },
    ] as ConfigField<Api.Config.AlistConfig>[],
  } as ConfigItem<Api.Config.AlistConfig>,
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
    ] as ConfigField<Api.Config.StrmConfig>[],
  } as ConfigItem<Api.Config.StrmConfig>,
  {
    name: '消息通知',
    code: 'NOTIFICATION_SETTINGS',
    fields: [] as ConfigField<Api.Config.NotificationConfig>[],
  } as ConfigItem<Api.Config.NotificationConfig>,
]
