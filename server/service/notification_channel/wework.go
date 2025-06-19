package notification_channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/MccRay-s/alist2strm/model/notification"
	"go.uber.org/zap"
)

// WeworkChannel 企业微信通知渠道
type WeworkChannel struct {
	*BaseChannel
	corpID      string
	agentID     string
	corpSecret  string
	toUser      string
	token       string
	tokenExpire time.Time
}

// GetType 获取渠道类型
func (c *WeworkChannel) GetType() notification.NotificationChannelType {
	return notification.ChannelTypeWework
}

// NewWeworkChannel 创建企业微信通知渠道
func NewWeworkChannel(logger *zap.Logger, settings *notification.Settings) Channel {
	channelConfig, exists := settings.Channels[string(notification.ChannelTypeWework)]
	if !exists || !channelConfig.Enabled {
		channel := &WeworkChannel{
			BaseChannel: NewBaseChannel(logger, settings),
			corpID:      "",
			agentID:     "",
			corpSecret:  "",
			toUser:      "",
		}
		channel.BaseChannel.enabled = false
		return channel
	}

	corpID := channelConfig.Config["corpId"]
	agentID := channelConfig.Config["agentId"]
	corpSecret := channelConfig.Config["corpSecret"]
	toUser := channelConfig.Config["toUser"]

	// 检查必要参数
	if corpID == "" || agentID == "" || corpSecret == "" {
		logger.Warn("企业微信配置不完整，通知功能已禁用",
			zap.String("corpID", corpID),
			zap.String("agentID", agentID))
		channel := &WeworkChannel{
			BaseChannel: NewBaseChannel(logger, settings),
			corpID:      "",
			agentID:     "",
			corpSecret:  "",
			toUser:      "",
		}
		channel.BaseChannel.enabled = false
		return channel
	}

	if toUser == "" {
		toUser = "@all"
	}

	channel := &WeworkChannel{
		BaseChannel: NewBaseChannel(logger, settings),
		corpID:      corpID,
		agentID:     agentID,
		corpSecret:  corpSecret,
		toUser:      toUser,
	}
	channel.BaseChannel.enabled = true
	return channel
}

// Send 发送通知
func (c *WeworkChannel) Send(templateType notification.TemplateType, data interface{}) error {
	if !c.enabled {
		return fmt.Errorf("企业微信通知渠道未启用")
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
func (c *WeworkChannel) getTemplateContent(templateType notification.TemplateType) string {
	templateConfig, exists := c.settings.Templates[string(templateType)]
	if !exists {
		return ""
	}
	return templateConfig.Wework
}

// renderTemplate 渲染模板
func (c *WeworkChannel) renderTemplate(templateContent string, data interface{}) (string, error) {
	tmpl, err := template.New("wework").Parse(templateContent)
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

// getAccessToken 获取访问令牌
func (c *WeworkChannel) getAccessToken() (string, error) {
	// 如果现有令牌有效，直接返回
	if c.token != "" && time.Now().Before(c.tokenExpire) {
		return c.token, nil
	}

	// 获取新令牌
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", c.corpID, c.corpSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取企业微信访问令牌失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析企业微信API响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("企业微信API错误: %s (%d)", result.ErrMsg, result.ErrCode)
	}

	// 更新令牌和过期时间（提前5分钟过期，以防临界问题）
	c.token = result.AccessToken
	c.tokenExpire = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)

	return c.token, nil
}

// sendMessage 发送消息
func (c *WeworkChannel) sendMessage(message string) error {
	// 获取访问令牌
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}

	// 构建API URL
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)

	// 构建请求体
	requestBody := map[string]interface{}{
		"touser":  c.toUser,
		"msgtype": "text",
		"agentid": c.agentID,
		"text": map[string]string{
			"content": message,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("JSON编码失败: %w", err)
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送企业微信消息失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析企业微信API响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("企业微信API错误: %s (%d)", result.ErrMsg, result.ErrCode)
	}

	return nil
}
