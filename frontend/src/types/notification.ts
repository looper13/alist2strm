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
