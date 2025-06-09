package controller

import (
	"strconv"

	"github.com/MccRay-s/alist2strm/model/common/response"
	"github.com/MccRay-s/alist2strm/model/user/request"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/MccRay-s/alist2strm/utils"
	"github.com/gin-gonic/gin"
)

// 包级别的用户控制器实例
var User = &UserController{}

type UserController struct{}

// Login 用户登录
func (uc *UserController) Login(c *gin.Context) {
	var req request.UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("用户登录参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	loginResp, err := service.User.Login(&req)
	if err != nil {
		utils.Error("用户登录失败", "username", req.Username, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("用户登录成功", "username", req.Username, "user_id", loginResp.User.ID, "request_id", c.GetString("request_id"))
	response.SuccessWithData(loginResp, c)
}

// Register 用户注册
func (uc *UserController) Register(c *gin.Context) {
	var req request.UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("用户注册参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	err := service.User.Register(&req)
	if err != nil {
		utils.Error("用户注册失败", "username", req.Username, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("用户注册成功", "username", req.Username, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("注册成功", c)
}

// GetUserInfo 获取用户信息
func (uc *UserController) GetUserInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("获取用户信息ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("用户ID参数错误", c)
		return
	}

	req := &request.UserInfoReq{}
	req.ID = id

	userInfo, err := service.User.GetUserInfo(req)
	if err != nil {
		utils.Error("获取用户信息失败", "user_id", id, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取用户信息成功", "user_id", id, "request_id", c.GetString("request_id"))
	response.SuccessWithData(userInfo, c)
}

// UpdateUser 更新用户信息
func (uc *UserController) UpdateUser(c *gin.Context) {
	// 从路径参数获取用户ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("更新用户信息ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("用户ID参数错误", c)
		return
	}

	var req request.UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("更新用户信息参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 设置用户ID
	req.ID = uint(id)

	err = service.User.UpdateUser(&req)
	if err != nil {
		utils.Error("更新用户信息失败", "user_id", req.ID, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("更新用户信息成功", "user_id", req.ID, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("更新成功", c)
}

// GetUserList 获取用户列表
func (uc *UserController) GetUserList(c *gin.Context) {
	var req request.UserListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error("获取用户列表参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	userList, err := service.User.GetUserList(&req)
	if err != nil {
		utils.Error("获取用户列表失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取用户列表成功", "total", userList.Total, "page", userList.Page, "request_id", c.GetString("request_id"))
	response.SuccessWithData(userList, c)
}

// Me 获取当前用户信息
func (uc *UserController) Me(c *gin.Context) {
	// 从JWT中间件获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Error("获取当前用户信息失败: 未找到用户ID", "request_id", c.GetString("request_id"))
		response.FailWithMessage("用户信息获取失败", c)
		return
	}

	// 构建请求
	req := &request.UserInfoReq{}
	req.ID = int(userID.(uint))

	userInfo, err := service.User.GetUserInfo(req)
	if err != nil {
		utils.Error("获取当前用户信息失败", "user_id", userID, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取当前用户信息成功", "user_id", userID, "request_id", c.GetString("request_id"))
	response.SuccessWithData(userInfo, c)
}
