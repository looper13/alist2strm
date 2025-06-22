package notification_channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/MccRay-s/alist2strm/model/notification"
	"go.uber.org/zap"
)

// TelegramChannel Telegram 通知渠道
type TelegramChannel struct {
	*BaseChannel
	botToken  string
	chatID    string
	parseMode string
}

// GetType 获取渠道类型
func (c *TelegramChannel) GetType() notification.NotificationChannelType {
	return notification.ChannelTypeTelegram
}

// NewTelegramChannel 创建 Telegram 通知渠道
func NewTelegramChannel(logger *zap.Logger, settings *notification.Settings) Channel {
	channelConfig, exists := settings.Channels[string(notification.ChannelTypeTelegram)]
	if !exists || !channelConfig.Enabled {
		channel := &TelegramChannel{
			BaseChannel: NewBaseChannel(logger, settings),
			botToken:    "",
			chatID:      "",
			parseMode:   "",
		}
		channel.BaseChannel.enabled = false
		return channel
	}

	botToken := channelConfig.Config["botToken"]
	chatID := channelConfig.Config["chatId"]
	parseMode := channelConfig.Config["parseMode"]

	// 检查必要参数
	if botToken == "" || chatID == "" {
		logger.Warn("Telegram 配置不完整，通知功能已禁用",
			zap.String("botToken", botToken),
			zap.String("chatID", chatID))
		channel := &TelegramChannel{
			BaseChannel: NewBaseChannel(logger, settings),
			botToken:    "",
			chatID:      "",
			parseMode:   "",
		}
		channel.BaseChannel.enabled = false
		return channel
	}

	if parseMode == "" {
		parseMode = "Markdown"
	}

	channel := &TelegramChannel{
		BaseChannel: NewBaseChannel(logger, settings),
		botToken:    botToken,
		chatID:      chatID,
		parseMode:   parseMode,
	}
	channel.BaseChannel.enabled = true
	return channel
}

// Send 发送通知
func (c *TelegramChannel) Send(templateType notification.TemplateType, data interface{}) error {
	if !c.enabled {
		return fmt.Errorf("telegram 通知渠道未启用")
	}

	// 获取模板
	templateContent := c.getTemplateContent(templateType)
	if templateContent == "" {
		return fmt.Errorf("未找到模板: %s", templateType)
	}

	// 渲染模板
	message, err := c.renderTemplate(templateContent, data)
	if err != nil {
		return fmt.Errorf("渲染模板失败: %w", err)
	}

	// 发送通知
	return c.sendMessage(message)
}

// getTemplateContent 获取模板内容
func (c *TelegramChannel) getTemplateContent(templateType notification.TemplateType) string {
	templateConfig, exists := c.settings.Templates[string(templateType)]
	if !exists {
		return ""
	}
	return templateConfig.Telegram
}

// renderTemplate 渲染模板
func (c *TelegramChannel) renderTemplate(templateContent string, data interface{}) (string, error) {
	// 创建模板
	tmpl := template.New("telegram")

	// 解析模板内容
	tmpl, err := tmpl.Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sendMessage 发送消息
func (c *TelegramChannel) sendMessage(message string) error {
	// 构建API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)

	// 构建请求参数
	params := url.Values{}
	params.Add("chat_id", c.chatID)
	params.Add("text", message)
	params.Add("parse_mode", c.parseMode)

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("发送Telegram消息失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Description string `json:"description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("telegram API错误 (HTTP %d): %s", resp.StatusCode, errorResp.Description)
		}
		return fmt.Errorf("telegram API错误 (HTTP %d)", resp.StatusCode)
	}

	return nil
}
