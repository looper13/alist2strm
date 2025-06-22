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
					"toUser":     "",
				},
			},
		},
		Templates: map[string]TemplateConfig{
			string(TemplateTypeTaskComplete): {
				Telegram: "🎬 *任务完成通知* ✅\n\n📋 *基本信息*\n• *任务名称*: `{{.TaskName}}`\n• *完成时间*: {{.EventTime}}\n• *处理耗时*: {{.Duration}}秒\n\n📊 *处理统计*\n• *STRM文件*: 总计 {{.GeneratedFile}}+{{.SkipFile}}\n  - 已生成: {{.GeneratedFile}}\n  - 已跳过: {{.SkipFile}}\n• *元数据*: 总计 {{.MetadataCount}}\n  - 已下载: {{.MetadataDownloaded}}\n  - 已跳过: {{.MetadataSkipped}}\n• *字幕*: 总计 {{.SubtitleCount}}\n  - 已下载: {{.SubtitleDownloaded}}\n  - 已跳过: {{.SubtitleSkipped}}\n\n📁 *路径信息*\n• *源路径*: `{{.SourcePath}}`\n• *目标路径*: `{{.TargetPath}}`",
				Wework:   "🎬 任务完成通知 ✅\n\n## 📋 任务概览\n**任务名称**：<font color=\"info\">`{{.TaskName}}`</font>\n**完成时间**：{{.EventTime}}\n**处理耗时**：<font color=\"info\">{{.Duration}}</font> 秒\n\n## 📊 处理统计\n**STRM文件** (总计 {{.GeneratedFile}}+{{.SkipFile}})\n> 已生成：<font color=\"info\">{{.GeneratedFile}}</font> | 已跳过：<font color=\"info\">{{.SkipFile}}</font>\n\n**元数据文件** (总计 {{.MetadataCount}})\n> 已下载：<font color=\"info\">{{.MetadataDownloaded}}</font> | 已跳过：<font color=\"info\">{{.MetadataSkipped}}</font>\n\n**字幕文件** (总计 {{.SubtitleCount}})\n> 已下载：<font color=\"info\">{{.SubtitleDownloaded}}</font> | 已跳过：<font color=\"info\">{{.SubtitleSkipped}}</font>\n\n## 📂 路径信息\n**源路径**：`{{.SourcePath}}`\n**目标路径**：`{{.TargetPath}}`",
			},
			string(TemplateTypeTaskFailed): {
				Telegram: "❌ *任务失败通知*\n\n📂 任务：`{{.TaskName}}`\n⏰ 时间：{{.EventTime}}\n⏱️ 耗时：{{.Duration}}秒\n❗ 错误信息：\n`{{.ErrorMessage}}`",
				Wework:   "❌ *任务失败通知*\n\n📂 任务：`{{.TaskName}}`\n⏰ 时间：{{.EventTime}}\n⏱️ 耗时：{{.Duration}}秒\n❗ 错误信息：\n`{{.ErrorMessage}}`",
			},
		},
		QueueSettings: QueueSettings{
			MaxRetries:    3,
			RetryInterval: 60,
			Concurrency:   1,
		},
	}
}
