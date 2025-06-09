package middleware

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// RequestIDHeader 请求ID的头名称
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey 在gin.Context中存储请求ID的键
	RequestIDKey = "request_id"
)

// RequestID 请求ID中间件
// 为每个请求生成唯一的ID，用于日志追踪和调试
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先检查请求头中是否已有请求ID
		requestID := c.GetHeader(RequestIDHeader)

		// 如果没有请求ID，则生成一个新的
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置请求ID到响应头
		c.Header(RequestIDHeader, requestID)

		// 将请求ID存储到上下文中，供其他中间件和处理器使用
		c.Set(RequestIDKey, requestID)

		// 继续处理请求
		c.Next()
	}
}

// generateRequestID 生成唯一的请求ID
// 格式：时间戳-随机字符串
func generateRequestID() string {
	// 使用当前时间戳（微秒）
	timestamp := time.Now().UnixMicro()

	// 生成4字节随机数
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)

	// 格式：timestamp-hex
	return fmt.Sprintf("%d-%x", timestamp, randomBytes)
}

// GetRequestID 从gin.Context中获取请求ID的辅助函数
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
