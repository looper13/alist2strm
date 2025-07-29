package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MccRay-s/alist2strm/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseFactory 数据库工厂接口
type DatabaseFactory interface {
	CreateConnection(config *config.DatabaseConfig) (*gorm.DB, error)
}

// DefaultDatabaseFactory 默认数据库工厂实现
type DefaultDatabaseFactory struct{}

// NewDatabaseFactory 创建数据库工厂实例
func NewDatabaseFactory() DatabaseFactory {
	return &DefaultDatabaseFactory{}
}

// CreateConnection 根据配置创建数据库连接
func (f *DefaultDatabaseFactory) CreateConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 验证配置
	if err := f.validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	switch cfg.Type {
	case "sqlite", "":
		return f.createSQLiteConnection(&cfg.SQLite)
	case "postgresql":
		return f.createPostgreSQLConnection(&cfg.PostgreSQL)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
	}
}

// validateConfig 验证数据库配置
func (f *DefaultDatabaseFactory) validateConfig(cfg *config.DatabaseConfig) error {
	switch cfg.Type {
	case "sqlite", "":
		return f.validateSQLiteConfig(&cfg.SQLite)
	case "postgresql":
		return f.validatePostgreSQLConfig(&cfg.PostgreSQL)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
	}
}

// validateSQLiteConfig 验证SQLite配置
func (f *DefaultDatabaseFactory) validateSQLiteConfig(cfg *config.SQLiteConfig) error {
	if cfg.BaseDir == "" {
		return NewConfigError("SQLite数据库目录不能为空", nil)
	}
	if cfg.Name == "" {
		return NewConfigError("SQLite数据库文件名不能为空", nil)
	}
	return nil
}

// validatePostgreSQLConfig 验证PostgreSQL配置
func (f *DefaultDatabaseFactory) validatePostgreSQLConfig(cfg *config.PostgreSQLConfig) error {
	if cfg.Host == "" {
		return NewConfigError("PostgreSQL主机地址不能为空", nil).WithContext("host", cfg.Host)
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return NewConfigError("PostgreSQL端口必须在1-65535范围内", nil).WithContext("port", fmt.Sprintf("%d", cfg.Port))
	}
	if cfg.Database == "" {
		return NewConfigError("PostgreSQL数据库名不能为空", nil)
	}
	if cfg.Username == "" {
		return NewConfigError("PostgreSQL用户名不能为空", nil)
	}
	if cfg.MaxOpenConns <= 0 {
		return NewConfigError("最大连接数必须大于0", nil).WithContext("max_open_conns", fmt.Sprintf("%d", cfg.MaxOpenConns))
	}
	if cfg.MaxIdleConns < 0 {
		return NewConfigError("最大空闲连接数不能小于0", nil).WithContext("max_idle_conns", fmt.Sprintf("%d", cfg.MaxIdleConns))
	}
	if cfg.MaxIdleConns > cfg.MaxOpenConns {
		return NewConfigError("最大空闲连接数不能大于最大连接数", nil).
			WithContext("max_idle_conns", fmt.Sprintf("%d", cfg.MaxIdleConns)).
			WithContext("max_open_conns", fmt.Sprintf("%d", cfg.MaxOpenConns))
	}
	return nil
}

// createSQLiteConnection 创建SQLite数据库连接
func (f *DefaultDatabaseFactory) createSQLiteConnection(cfg *config.SQLiteConfig) (*gorm.DB, error) {
	dbDir := cfg.BaseDir
	dbPath := filepath.Join(dbDir, cfg.Name)

	// 创建数据库目录
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, NewConnectionError("创建数据库目录失败", err).
			WithContext("db_dir", dbDir).
			WithContext("db_path", dbPath)
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, NewConnectionError("连接SQLite数据库失败", err).
			WithContext("db_path", dbPath)
	}

	return db, nil
}

// createPostgreSQLConnection 创建PostgreSQL数据库连接
func (f *DefaultDatabaseFactory) createPostgreSQLConnection(cfg *config.PostgreSQLConfig) (*gorm.DB, error) {
	// 构建DSN连接字符串
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		// 根据错误类型返回不同的错误
		if isAuthenticationError(err) {
			return nil, NewAuthenticationError("PostgreSQL认证失败", err).
				WithContext("host", cfg.Host).
				WithContext("port", fmt.Sprintf("%d", cfg.Port)).
				WithContext("database", cfg.Database).
				WithContext("username", cfg.Username)
		}
		if isPermissionError(err) {
			return nil, NewPermissionError("PostgreSQL权限不足", err).
				WithContext("host", cfg.Host).
				WithContext("database", cfg.Database).
				WithContext("username", cfg.Username)
		}
		return nil, NewConnectionError("连接PostgreSQL数据库失败", err).
			WithContext("host", cfg.Host).
			WithContext("port", fmt.Sprintf("%d", cfg.Port)).
			WithContext("database", cfg.Database)
	}

	// 获取底层的sql.DB实例来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, NewConnectionError("获取数据库实例失败", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	return db, nil
}

// isAuthenticationError 检查是否为认证错误
func isAuthenticationError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "password authentication failed") ||
		contains(errStr, "authentication failed") ||
		contains(errStr, "invalid authorization specification")
}

// isPermissionError 检查是否为权限错误
func isPermissionError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "permission denied") ||
		contains(errStr, "insufficient privilege") ||
		contains(errStr, "access denied")
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(substr) == 0 ||
			(len(s) > len(substr) && containsIgnoreCase(s, substr)))
}

// containsIgnoreCase 不区分大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower 转换为小写
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}
