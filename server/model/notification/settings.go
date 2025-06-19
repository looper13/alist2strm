package notification

// Settings 通知系统配置
type Settings struct {
	Enabled        bool                      `json:"enabled"`
	DefaultChannel string                    `json:"defaultChannel"`
	Channels       map[string]ChannelConfig  `json:"channels"`
	Templates      map[string]TemplateConfig `json:"templates"`
	QueueSettings  QueueSettings             `json:"queueSettings"`
}

// ChannelConfig 通知渠道配置
type ChannelConfig struct {
	Enabled bool              `json:"enabled"`
	Type    string            `json:"type"`
	Config  map[string]string `json:"config"`
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	Telegram string `json:"telegram"`
	Wework   string `json:"wework"`
}

// QueueSettings 队列设置
type QueueSettings struct {
	MaxRetries    int `json:"maxRetries"`
	RetryInterval int `json:"retryInterval"` // 秒
	Concurrency   int `json:"concurrency"`
}

// NotificationChannelType 通知渠道类型
type NotificationChannelType string

const (
	// ChannelTypeTelegram Telegram 通知渠道
	ChannelTypeTelegram NotificationChannelType = "telegram"
	// ChannelTypeWework 企业微信通知渠道
	ChannelTypeWework NotificationChannelType = "wework"
)

// TemplateType 通知模板类型
type TemplateType string

const (
	// TemplateTypeTaskComplete 任务完成通知模板
	TemplateTypeTaskComplete TemplateType = "taskComplete"
	// TemplateTypeTaskFailed 任务失败通知模板
	TemplateTypeTaskFailed TemplateType = "taskFailed"
)

// DefaultSettings 返回默认通知设置
func DefaultSettings() *Settings {
	return &Settings{
		Enabled:        true,
		DefaultChannel: string(ChannelTypeTelegram),
		Channels: map[string]ChannelConfig{
			string(ChannelTypeTelegram): {
				Enabled: false,
				Type:    string(ChannelTypeTelegram),
				Config: map[string]string{
					"botToken":  "",
					"chatId":    "",
					"parseMode": "Markdown",
				},
			},
			string(ChannelTypeWework): {
				Enabled: false,
				Type:    string(ChannelTypeWework),
				Config: map[string]string{
					"corpId":     "",
					"agentId":    "",
					"corpSecret": "",
					"toUser":     "@all",
				},
			},
		},
		Templates: map[string]TemplateConfig{
			string(TemplateTypeTaskComplete): {
				Telegram: "✅ *任务完成通知*\n\n📂 任务：`{{.TaskName}}`\n⏱️ 耗时：{{.Duration}}秒\n📊 处理结果：\n - 总文件：{{.TotalFiles}}个\n - 已生成：{{.GeneratedFiles}}个\n - 已跳过：{{.SkippedFiles}}个\n - 元数据：{{.MetadataFiles}}个\n - 字幕：{{.SubtitleFiles}}个",
				Wework:   "【任务完成通知】\n\n任务：{{.TaskName}}\n耗时：{{.Duration}}秒\n处理结果：\n- 总文件：{{.TotalFiles}}个\n- 已生成：{{.GeneratedFiles}}个\n- 已跳过：{{.SkippedFiles}}个\n- 元数据：{{.MetadataFiles}}个\n- 字幕：{{.SubtitleFiles}}个",
			},
			string(TemplateTypeTaskFailed): {
				Telegram: "❌ *任务失败通知*\n\n📂 任务：`{{.TaskName}}`\n⏱️ 耗时：{{.Duration}}秒\n❗ 错误信息：\n`{{.ErrorMessage}}`",
				Wework:   "【任务失败通知】\n\n任务：{{.TaskName}}\n耗时：{{.Duration}}秒\n错误信息：\n{{.ErrorMessage}}",
			},
		},
		QueueSettings: QueueSettings{
			MaxRetries:    3,
			RetryInterval: 60,
			Concurrency:   1,
		},
	}
}
