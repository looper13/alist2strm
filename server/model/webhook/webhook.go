package webhook

import (
	"strconv"
	"strings"
	"time"
)

// CustomTime is a custom time type to handle both RFC3339 and Unix timestamps in JSON
type CustomTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	// Try to parse as a Unix timestamp (integer) first, as that's the source of the error
	unix, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		ct.Time = time.Unix(unix, 0)
		return nil
	}

	// If it fails, try to parse as RFC3339 string
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err // Return the parsing error if both fail
	}
	ct.Time = t
	return nil
}

// CustomBool is a custom bool type to handle both bool and string representations in JSON
type CustomBool bool

// UnmarshalJSON implements the json.Unmarshaler interface.
func (cb *CustomBool) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	val, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*cb = CustomBool(val)
	return nil
}

// FileWebhookPayload 对应 file_system_watcher 的 Webhook 负载
type FileWebhookPayload struct {
	DeviceName    string            `json:"device_name"`
	UserName      string            `json:"user_name"`
	Version       string            `json:"version"`
	EventCategory string            `json:"event_category"`
	EventName     string            `json:"event_name"`
	EventTime     CustomTime        `json:"event_time"` // TOML 模板是字符串，但通常会是 ISO 8601 格式，Go 可以自动解析
	SendTime      CustomTime        `json:"send_time"`
	Data          []FileChangeEvent `json:"data"`
}

// FileChangeEvent 描述了单个文件变更事件
type FileChangeEvent struct {
	Action          string     `json:"action"`           // create, delete, rename
	IsDir           CustomBool `json:"is_dir"`           // 是否为目录
	SourceFile      string     `json:"source_file"`      // 源文件路径
	DestinationFile string     `json:"destination_file"` // 目标文件路径 (仅对 rename 有效)
}

// MountWebhookPayload 对应 mount_point_watcher 的 Webhook 负载
type MountWebhookPayload struct {
	DeviceName    string                  `json:"device_name"`
	UserName      string                  `json:"user_name"`
	Version       string                  `json:"version"`
	EventCategory string                  `json:"event_category"`
	EventName     string                  `json:"event_name"`
	EventTime     CustomTime              `json:"event_time"`
	SendTime      CustomTime              `json:"send_time"`
	Data          []MountPointChangeEvent `json:"data"`
}

// MountPointChangeEvent 描述了单个挂载点变更事件
type MountPointChangeEvent struct {
	Action     string     `json:"action"`      // mount, unmount
	MountPoint string     `json:"mount_point"` // 挂载点路径
	Status     CustomBool `json:"status"`      // 动作状态 (true 成功, false 失败)
	Reason     string     `json:"reason"`      // 失败原因
}
