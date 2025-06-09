package middleware

import (
	"strings"

	"github.com/MccRay-s/alist2strm/model/common/response"
	"github.com/MccRay-s/alist2strm/utils"
	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Warn("JWT认证失败: 缺少Authorization头", "request_id", c.GetString("request_id"))
			response.FailWithMessage("未提供认证token", c)
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Warn("JWT认证失败: Authorization格式错误", "authorization", authHeader, "request_id", c.GetString("request_id"))
			response.FailWithMessage("认证token格式错误", c)
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			utils.Warn("JWT认证失败: token为空", "request_id", c.GetString("request_id"))
			response.FailWithMessage("认证token为空", c)
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			utils.Warn("JWT认证失败: token解析错误", "error", err.Error(), "request_id", c.GetString("request_id"))
			response.FailWithMessage("认证token无效", c)
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		utils.Debug("JWT认证成功", "user_id", claims.UserID, "username", claims.Username, "request_id", c.GetString("request_id"))
		c.Next()
	}
}
