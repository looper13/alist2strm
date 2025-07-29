package database

import (
	"fmt"
	"time"
)

// DatabaseErrorType 数据库错误类型
type DatabaseErrorType string

const (
	// ErrorTypeConfig 配置错误
	ErrorTypeConfig DatabaseErrorType = "CONFIG"
	// ErrorTypeConnection 连接错误
	ErrorTypeConnection DatabaseErrorType = "CONNECTION"
	// ErrorTypeAuthentication 认证错误
	ErrorTypeAuthentication DatabaseErrorType = "AUTHENTICATION"
	// ErrorTypePermission 权限错误
	ErrorTypePermission DatabaseErrorType = "PERMISSION"
	// ErrorTypeMigration 迁移错误
	ErrorTypeMigration DatabaseErrorType = "MIGRATION"
	// ErrorTypeQuery 查询错误
	ErrorTypeQuery DatabaseErrorType = "QUERY"
	// ErrorTypeTransaction 事务错误
	ErrorTypeTransaction DatabaseErrorType = "TRANSACTION"
	// ErrorTypeTimeout 超时错误
	ErrorTypeTimeout DatabaseErrorType = "TIMEOUT"
)

// DatabaseError 数据库错误结构
type DatabaseError struct {
	Type      DatabaseErrorType `json:"type"`
	Message   string            `json:"message"`
	Cause     error             `json:"cause,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Context   map[string]string `json:"context,omitempty"`
}

// Error 实现error接口
func (e *DatabaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("数据库错误 [%s]: %s (原因: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("数据库错误 [%s]: %s", e.Type, e.Message)
}

// Unwrap 返回原始错误
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// NewDatabaseError 创建新的数据库错误
func NewDatabaseError(errorType DatabaseErrorType, message string, cause error) *DatabaseError {
	return &DatabaseError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		Context:   make(map[string]string),
	}
}

// WithContext 添加上下文信息
func (e *DatabaseError) WithContext(key, value string) *DatabaseError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// NewConfigError 创建配置错误
func NewConfigError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypeConfig, message, cause)
}

// NewConnectionError 创建连接错误
func NewConnectionError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypeConnection, message, cause)
}

// NewAuthenticationError 创建认证错误
func NewAuthenticationError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypeAuthentication, message, cause)
}

// NewPermissionError 创建权限错误
func NewPermissionError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypePermission, message, cause)
}

// NewMigrationError 创建迁移错误
func NewMigrationError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypeMigration, message, cause)
}

// NewQueryError 创建查询错误
func NewQueryError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypeQuery, message, cause)
}

// NewTransactionError 创建事务错误
func NewTransactionError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypeTransaction, message, cause)
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message string, cause error) *DatabaseError {
	return NewDatabaseError(ErrorTypeTimeout, message, cause)
}
