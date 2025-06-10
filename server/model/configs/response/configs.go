package response

import "time"

// ConfigInfo 配置信息响应
type ConfigInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Value     string    `json:"value"`
}
