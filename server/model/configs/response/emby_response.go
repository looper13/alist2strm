package response

// EmbyConnectionTestResult 包含Emby连接测试的结果
type EmbyConnectionTestResult struct {
	Connected       bool   `json:"connected"`
	Error           string `json:"error,omitempty"`
	Version         string `json:"version,omitempty"`
	ServerName      string `json:"serverName,omitempty"`
	OperatingSystem string `json:"operatingSystem,omitempty"`
}
