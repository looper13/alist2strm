package handler

import (
	"alist2strm/internal/service"
	"alist2strm/internal/utils"

	"github.com/gin-gonic/gin"
)

// RegisterHandler 处理用户注册请求
func RegisterHandler(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	if err := service.GetUserService().Register(&req); err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "注册成功", nil)
}

// LoginHandler 处理用户登录请求
func LoginHandler(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	resp, err := service.GetUserService().Login(&req)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, resp)
}

// GetUserInfoHandler 获取用户信息
func GetUserInfoHandler(c *gin.Context) {
	userID := utils.GetContextUserID(c)
	if userID == 0 {
		utils.ResponseError(c, "未登录")
		return
	}

	userInfo, err := service.GetUserService().GetUserInfo(userID)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, userInfo)
}

// UpdateUserInfoHandler 更新用户信息
func UpdateUserInfoHandler(c *gin.Context) {
	userID := utils.GetContextUserID(c)
	if userID == 0 {
		utils.ResponseError(c, "未登录")
		return
	}

	var req service.UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, "参数错误")
		return
	}

	if err := service.GetUserService().UpdateUserInfo(userID, &req); err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "更新成功", nil)
}
