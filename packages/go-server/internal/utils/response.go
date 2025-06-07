package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务码，0表示成功
	Message string      `json:"message"` // 消息提示
	Data    interface{} `json:"data"`    // 数据内容
}

// PageData 分页数据结构
type PageData struct {
	List  interface{} `json:"list"`  // 列表数据
	Total int64       `json:"total"` // 总记录数
	Page  int         `json:"page"`  // 当前页码
	Size  int         `json:"size"`  // 每页大小
}

// ResponseSuccess 返回成功响应
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ResponseSuccessWithMessage 返回带消息的成功响应
func ResponseSuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// ResponsePage 返回分页响应
func ResponsePage(c *gin.Context, list interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	})
}

// ResponseError 返回错误响应
func ResponseError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    -1,
		Message: message,
		Data:    nil,
	})
}

// ResponseErrorWithCode 返回带错误码的错误响应
func ResponseErrorWithCode(c *gin.Context, code int, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// GetContextUserID 从上下文中获取用户ID
func GetContextUserID(c *gin.Context) uint {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return 0
	}
	if userID, ok := userIDValue.(uint); ok {
		return userID
	}
	return 0
}
