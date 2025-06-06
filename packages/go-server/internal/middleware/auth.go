package middleware

import (
	"alist2strm/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Warn("认证失败：缺少Authorization头",
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "请先登录"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Warn("认证失败：Authorization格式错误",
				zap.String("path", c.Request.URL.Path),
				zap.String("auth", authHeader))
			c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "认证格式错误"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.Warn("认证失败：Token无效或已过期",
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, utils.NewErrorResponse(http.StatusUnauthorized, "Token无效或已过期"))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("nickname", claims.Nickname)

		utils.Debug("用户认证成功",
			zap.String("path", c.Request.URL.Path),
			zap.Uint("userId", claims.UserID),
			zap.String("username", claims.Username))

		c.Next()
	}
}
