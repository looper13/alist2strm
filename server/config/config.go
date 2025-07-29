package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
}

// LogConfig 日志配置
type LogConfig struct {
	BaseDir     string
	AppName     string
	Level       string
	MaxDays     int
	MaxFileSize int
	MaxBackups  int
	Compress    bool
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type       string           `json:"type"` // "sqlite" 或 "postgresql"
	SQLite     SQLiteConfig     `json:"sqlite"`
	PostgreSQL PostgreSQLConfig `json:"postgresql"`
}

// SQLiteConfig SQLite数据库配置
type SQLiteConfig struct {
	BaseDir string `json:"base_dir"`
	Name    string `json:"name"`
}

// PostgreSQLConfig PostgreSQL数据库配置
type PostgreSQLConfig struct {
	Host                      string `json:"host"`
	Port                      int    `json:"port"`
	Database                  string `json:"database"`
	Username                  string `json:"username"`
	Password                  string `json:"password"`
	SSLMode                   string `json:"ssl_mode"`
	MaxOpenConns              int    `json:"max_open_conns"`
	MaxIdleConns              int    `json:"max_idle_conns"`
	ConnMaxLifetime           int    `json:"conn_max_lifetime"`           // 以分钟为单位
	SlowQueryThreshold        int    `json:"slow_query_threshold"`        // 慢查询阈值，毫秒
	EnablePerformanceLog      bool   `json:"enable_performance_log"`      // 是否启用性能日志
	PerformanceReportInterval int    `json:"performance_report_interval"` // 性能报告间隔，分钟
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey string
	ExpiresIn string
}

// UserConfig 用户配置
type UserConfig struct {
	Name     string
	Password string
}

// AppConfig 应用配置
type AppConfig struct {
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
	JWT      JWTConfig
	User     UserConfig
}

// 全局配置变量
var GlobalConfig *AppConfig

// loadEnvFiles 根据环境加载配置文件
func loadEnvFiles() {
	// 获取当前环境
	appEnv := os.Getenv("APP_ENV")

	// 如果没有设置APP_ENV，默认为生产环境
	if appEnv == "" {
		// 生产环境：只读取系统环境变量，不加载文件
		return
	}

	// 开发环境：加载对应的环境文件
	envFile := ".env." + appEnv

	// 尝试从当前目录加载
	if _, err := os.Stat(envFile); err == nil {
		if loadErr := godotenv.Load(envFile); loadErr == nil {
			return
		}
	}

	// 尝试从父目录加载
	if wd, err := os.Getwd(); err == nil {
		parentDir := filepath.Dir(wd)
		envPath := filepath.Join(parentDir, envFile)
		if _, err := os.Stat(envPath); err == nil {
			godotenv.Load(envPath)
		}
	}
}

// LoadConfig 加载配置
func LoadConfig() *AppConfig {
	// 尝试加载环境文件
	loadEnvFiles()

	GlobalConfig = &AppConfig{
		Server: ServerConfig{
			Port: getEnv("PORT", "3210"),
		},
		Log: LogConfig{
			BaseDir:     getEnv("LOG_BASE_DIR", "../data/logs"),
			AppName:     getEnv("LOG_APP_NAME", "alist2strm"),
			Level:       getEnv("LOG_LEVEL", "info"),
			MaxDays:     getEnvAsInt("LOG_MAX_DAYS", 7),
			MaxFileSize: getEnvAsInt("LOG_MAX_FILE_SIZE", 10),
			MaxBackups:  getEnvAsInt("LOG_MAX_BACKUPS", 5),
			Compress:    getEnvAsBool("LOG_COMPRESS", true),
		},
		Database: DatabaseConfig{
			Type: getEnv("DB_TYPE", "sqlite"),
			SQLite: SQLiteConfig{
				BaseDir: getEnv("DB_BASE_DIR", "../data/db"),
				Name:    getEnv("DB_NAME", "database.sqlite"),
			},
			PostgreSQL: PostgreSQLConfig{
				Host:                      getEnv("DB_HOST", "localhost"),
				Port:                      getEnvAsInt("DB_PORT", 5432),
				Database:                  getEnv("DB_DATABASE", "alist2strm"),
				Username:                  getEnv("DB_USERNAME", "postgres"),
				Password:                  getEnv("DB_PASSWORD", ""),
				SSLMode:                   getEnv("DB_SSL_MODE", "disable"),
				MaxOpenConns:              getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
				MaxIdleConns:              getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
				ConnMaxLifetime:           getEnvAsInt("DB_CONN_MAX_LIFETIME", 60),     // 分钟
				SlowQueryThreshold:        getEnvAsInt("DB_SLOW_QUERY_THRESHOLD", 100), // 毫秒
				EnablePerformanceLog:      getEnvAsBool("DB_ENABLE_PERFORMANCE_LOG", true),
				PerformanceReportInterval: getEnvAsInt("DB_PERFORMANCE_REPORT_INTERVAL", 5), // 分钟
			},
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET_KEY", "alist2strm-default-jwt-secret-key-2025"),
			ExpiresIn: getEnv("JWT_EXPIRES_IN", "24"),
		},
		User: UserConfig{
			Name:     getEnv("USER_NAME", "admin"),
			Password: getEnv("USER_PASSWORD", ""),
		},
	}

	return GlobalConfig
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
