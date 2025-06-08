package service

import (
	"alist2strm/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"go.uber.org/zap"
)

type TelegramNotificationService struct {
	config *TelegramConfig
	client *http.Client
	logger *zap.Logger
}

// TelegramConfig Telegram 配置
type TelegramConfig struct {
	Enabled   bool              `json:"enabled"`
	BotToken  string            `json:"botToken"`
	ChatID    string            `json:"chatId"`
	Timeout   int               `json:"timeout"`
	Templates map[string]string `json:"templates"`
}

// TelegramMessage Telegram 消息
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// TelegramResponse Telegram API 响应
type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
}

var (
	telegramNotificationService *TelegramNotificationService
	telegramNotificationOnce    sync.Once
)

// GetTelegramNotificationService 获取 TelegramNotificationService 单例
func GetTelegramNotificationService() *TelegramNotificationService {
	telegramNotificationOnce.Do(func() {
		telegramNotificationService = &TelegramNotificationService{
			client: &http.Client{
				Timeout: 30 * time.Second,
			},
			logger: utils.Logger,
		}
		telegramNotificationService.loadConfig()
	})
	return telegramNotificationService
}

// loadConfig 加载 Telegram 配置
func (s *TelegramNotificationService) loadConfig() {
	configService := GetConfigService()
	configValue, err := configService.GetByCode("TELEGRAM")
	if err != nil {
		s.logger.Warn("获取 Telegram 配置失败", zap.Error(err))
		s.config = &TelegramConfig{Enabled: false}
		return
	}

	var config TelegramConfig
	if err := json.Unmarshal([]byte(configValue.Value), &config); err != nil {
		s.logger.Error("解析 Telegram 配置失败", zap.Error(err))
		s.config = &TelegramConfig{Enabled: false}
		return
	}

	s.config = &config
	if s.config.Timeout > 0 {
		s.client.Timeout = time.Duration(s.config.Timeout) * time.Second
	}

	s.logger.Info("加载 Telegram 配置成功",
		zap.Bool("enabled", s.config.Enabled),
		zap.String("chatId", s.config.ChatID))
}

// ReloadConfig 重新加载配置
func (s *TelegramNotificationService) ReloadConfig() {
	s.loadConfig()
}

// IsEnabled 检查是否启用
func (s *TelegramNotificationService) IsEnabled() bool {
	return s.config != nil && s.config.Enabled && s.config.BotToken != "" && s.config.ChatID != ""
}

// SendTaskCompletedNotification 发送任务完成通知
func (s *TelegramNotificationService) SendTaskCompletedNotification(payload map[string]interface{}) (bool, string) {
	if !s.IsEnabled() {
		return false, "Telegram 通知未启用或配置不完整"
	}

	template := s.getTemplate("task_completed")
	message, err := s.renderTemplate(template, payload)
	if err != nil {
		return false, fmt.Sprintf("渲染消息模板失败: %v", err)
	}

	if err := s.SendMessage(message); err != nil {
		return false, fmt.Sprintf("发送 Telegram 消息失败: %v", err)
	}

	return true, "Telegram 任务完成通知发送成功"
}

// SendTaskFailedNotification 发送任务失败通知
func (s *TelegramNotificationService) SendTaskFailedNotification(payload map[string]interface{}) (bool, string) {
	if !s.IsEnabled() {
		return false, "Telegram 通知未启用或配置不完整"
	}

	template := s.getTemplate("task_failed")
	message, err := s.renderTemplate(template, payload)
	if err != nil {
		return false, fmt.Sprintf("渲染消息模板失败: %v", err)
	}

	if err := s.SendMessage(message); err != nil {
		return false, fmt.Sprintf("发送 Telegram 消息失败: %v", err)
	}

	return true, "Telegram 任务失败通知发送成功"
}

// SendFileInvalidNotification 发送文件失效通知
func (s *TelegramNotificationService) SendFileInvalidNotification(payload map[string]interface{}) (bool, string) {
	if !s.IsEnabled() {
		return false, "Telegram 通知未启用或配置不完整"
	}

	template := s.getTemplate("file_invalid")
	message, err := s.renderTemplate(template, payload)
	if err != nil {
		return false, fmt.Sprintf("渲染消息模板失败: %v", err)
	}

	if err := s.SendMessage(message); err != nil {
		return false, fmt.Sprintf("发送 Telegram 消息失败: %v", err)
	}

	return true, "Telegram 文件失效通知发送成功"
}

// SendMessage 发送消息
func (s *TelegramNotificationService) SendMessage(text string) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Telegram 服务未启用")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.config.BotToken)

	message := TelegramMessage{
		ChatID:    s.config.ChatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var telegramResp TelegramResponse
	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return err
	}

	if !telegramResp.OK {
		return fmt.Errorf("Telegram API 错误 (%d): %s", telegramResp.ErrorCode, telegramResp.Description)
	}

	s.logger.Info("Telegram 消息发送成功", zap.String("chatId", s.config.ChatID))
	return nil
}

// TestConnection 测试连接
func (s *TelegramNotificationService) TestConnection() error {
	if !s.IsEnabled() {
		return fmt.Errorf("Telegram 服务未启用或配置不完整")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", s.config.BotToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var telegramResp TelegramResponse
	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return err
	}

	if !telegramResp.OK {
		return fmt.Errorf("Telegram API 错误 (%d): %s", telegramResp.ErrorCode, telegramResp.Description)
	}

	// 发送测试消息
	testMessage := "🤖 Alist2Strm 连接测试成功！"
	return s.SendMessage(testMessage)
}

// getTemplate 获取消息模板
func (s *TelegramNotificationService) getTemplate(templateName string) string {
	if s.config.Templates != nil {
		if template, exists := s.config.Templates[templateName]; exists {
			return template
		}
	}

	// 返回默认模板
	switch templateName {
	case "task_completed":
		return `📊 *任务执行完成*
🎬 任务：{{.TaskName}}
📁 路径：{{.SourcePath}}
✅ 成功：{{.SuccessCount}} 个文件
❌ 失败：{{.FailedCount}} 个文件
⏩ 跳过：{{.SkippedCount}} 个文件
🕒 用时：{{.Duration}}`

	case "task_failed":
		return `❌ *任务执行失败*
🎬 任务：{{.TaskName}}
📁 路径：{{.SourcePath}}
💥 错误：{{.ErrorMessage}}`

	case "file_invalid":
		return `⚠️ *文件失效检测*
📁 共检测：{{.TotalFiles}} 个文件
❌ 失效文件：{{.InvalidFiles}} 个
🔗 主要原因：{{.MainReason}}`

	default:
		return "{{.Message}}"
	}
}

// renderTemplate 渲染模板
func (s *TelegramNotificationService) renderTemplate(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("telegram").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// formatDuration 格式化持续时间
func (s *TelegramNotificationService) formatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%.1f秒", duration.Seconds())
	} else if duration < time.Hour {
		return fmt.Sprintf("%.1f分钟", duration.Minutes())
	} else {
		return fmt.Sprintf("%.1f小时", duration.Hours())
	}
}

// escapeMarkdown 转义 Markdown 特殊字符
func (s *TelegramNotificationService) escapeMarkdown(text string) string {
	// 转义 Telegram Markdown 特殊字符
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

// GetConfig 获取当前配置
func (s *TelegramNotificationService) GetConfig() *TelegramConfig {
	return s.config
}

// UpdateConfig 更新配置
func (s *TelegramNotificationService) UpdateConfig(config *TelegramConfig) error {
	configService := GetConfigService()

	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := configService.UpdateByCode("TELEGRAM", string(configData)); err != nil {
		return err
	}

	s.config = config
	if s.config.Timeout > 0 {
		s.client.Timeout = time.Duration(s.config.Timeout) * time.Second
	}

	return nil
}
