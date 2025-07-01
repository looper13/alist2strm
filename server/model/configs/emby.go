package configs

// EmbyConfig 包含Emby服务器配置
type EmbyConfig struct {
	EmbyServer   string        `json:"embyServer"`
	EmbyToken    string        `json:"embyToken"`
	PathMappings []PathMapping `json:"pathMappings"`
}

// PathMapping 包含路径映射配置
type PathMapping struct {
	Path     string `json:"path"`
	EmbyPath string `json:"embyPath"`
}
