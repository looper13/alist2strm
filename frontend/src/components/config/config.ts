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

export interface AlistConfig {
  token: string
  host: string
  domain: string
  reqInterval: number
  reqRetryCount: number
  reqRetryInterval: number
}

export interface StrmConfig {
  defaultSuffix: string
  replaceSuffix: boolean
  urlEncode: boolean
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
  } as AlistConfig,
  STRM: {
    defaultSuffix: 'mp4,mkv,avi,mov,rmvb,webm,flv,m3u8',
    replaceSuffix: true,
    urlEncode: true,
  } as StrmConfig,
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
    ] as ConfigField<AlistConfig>[],
  } as ConfigItem<AlistConfig>,
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
    ] as ConfigField<StrmConfig>[],
  } as ConfigItem<StrmConfig>,
]
