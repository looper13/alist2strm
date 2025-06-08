package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/MccRay-s/alist2strm/utils"
	"github.com/gin-gonic/gin"
)

// responseWriter 包装 gin.ResponseWriter 以捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AccessLogger 访问日志中间件
func AccessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器
		wrapper := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		c.Writer = wrapper

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(startTime)

		// 记录访问日志
		if utils.AccessLogger != nil {
			utils.AccessLogger.Infow("HTTP Request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"query", c.Request.URL.RawQuery,
				"ip", c.ClientIP(),
				"user_agent", c.Request.UserAgent(),
				"referer", c.Request.Referer(),
				"status", c.Writer.Status(),
				"duration", duration,
				"response_size", wrapper.body.Len(),
				"request_id", c.GetHeader("X-Request-ID"),
				"timestamp", startTime.Format(time.RFC3339),
			)
		}
	}
}
