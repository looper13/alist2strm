package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// 服务器配置
	Server struct {
		Port string
	}

	// 日志配置
	Log struct {
		BaseDir     string
		AppName     string
		Level       string
		MaxDays     int
		MaxFileSize int
	}

	// 数据库配置
	Database struct {
		BaseDir string
		Name    string
		Path    string // 完整路径，由 BaseDir 和 Name 组合而成
	}

	// JWT配置
	JWT struct {
		SecretKey string
		ExpiresIn int
	}

	USER struct {
		UserName     string
		UserPassword string
	}
}

var GlobalConfig Config

// 从环境变量获取字符串值，如果环境变量不存在则返回默认值
func getEnvStr(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// 从环境变量获取整数值，如果环境变量不存在或转换失败则返回默认值
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func Init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		fmt.Printf("警告: .env 文件未找到，将使用默认配置: %v\n", err)
	}

	// 服务器配置
	GlobalConfig.Server.Port = ":" + getEnvStr("PORT", "8080")

	// 日志配置
	GlobalConfig.Log.BaseDir = getEnvStr("LOG_BASE_DIR", "./data/logs")
	GlobalConfig.Log.AppName = getEnvStr("LOG_APP_NAME", "alist-strm")
	GlobalConfig.Log.Level = getEnvStr("LOG_LEVEL", "info")
	GlobalConfig.Log.MaxDays = getEnvInt("LOG_MAX_DAYS", 7)
	GlobalConfig.Log.MaxFileSize = getEnvInt("LOG_MAX_FILE_SIZE", 10)

	// 数据库配置
	GlobalConfig.Database.BaseDir = getEnvStr("DB_BASE_DIR", "./data/db")
	GlobalConfig.Database.Name = getEnvStr("DB_NAME", "database.sqlite")
	GlobalConfig.Database.Path = filepath.Join(GlobalConfig.Database.BaseDir, GlobalConfig.Database.Name)

	// JWT配置（保持默认值，因为这些是安全敏感的配置）
	GlobalConfig.JWT.SecretKey = getEnvStr("JWT_SECRET_KEY", "63fe1d02ac6da7fe325f3e7545f9b954dc76f25495f73f6d0c0dc82ad44d5fd3") // 在生产环境中应该使用环境变量设置
	GlobalConfig.JWT.ExpiresIn = getEnvInt("JWT_EXPIRES_IN", 168)                                                                // 24小时

	GlobalConfig.USER.UserName = getEnvStr("USER_NAME", "admin")
	GlobalConfig.USER.UserPassword = getEnvStr("USER_PASSWORD", "") // 在生产环境中应该使用环境变量设置
}
