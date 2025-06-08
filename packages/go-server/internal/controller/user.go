package controller

import (
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	userService *service.UserService
	logger      *zap.Logger
}

// NewUserController 创建用户控制器
func NewUserController() *UserController {
	return &UserController{
		userService: service.GetUserService(),
		logger:      utils.Logger,
	}
}

// RegisterRoutes 注册路由
func (c *UserController) RegisterRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/info", c.GetUserInfo)
		userGroup.PUT("/:id", c.UpdateUserInfo)
	}
}

// Register 处理用户注册请求
func (c *UserController) Register(ctx *gin.Context) {
	var req service.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析注册请求失败", zap.Error(err))
		utils.ResponseError(ctx, "参数错误")
		return
	}

	if err := c.userService.Register(&req); err != nil {
		c.logger.Error("用户注册失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("用户注册成功", zap.String("username", req.Username))
	utils.ResponseSuccessWithMessage(ctx, "注册成功", nil)
}

// Login 处理用户登录请求
func (c *UserController) Login(ctx *gin.Context) {
	var req service.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析登录请求失败", zap.Error(err))
		utils.ResponseError(ctx, "参数错误")
		return
	}

	resp, err := c.userService.Login(&req)
	if err != nil {
		c.logger.Error("用户登录失败", zap.Error(err), zap.String("username", req.Username))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("用户登录成功", zap.String("username", req.Username))
	utils.ResponseSuccess(ctx, resp)
}

// GetUserInfo 获取用户信息
func (c *UserController) GetUserInfo(ctx *gin.Context) {
	userID := utils.GetContextUserID(ctx)
	if userID == 0 {
		utils.ResponseError(ctx, "未登录")
		return
	}

	userInfo, err := c.userService.GetUserInfo(userID)
	if err != nil {
		c.logger.Error("获取用户信息失败", zap.Error(err), zap.Uint("userId", userID))
		utils.ResponseError(ctx, err.Error())
		return
	}

	utils.ResponseSuccess(ctx, userInfo)
}

// UpdateUserInfo 更新用户信息
func (c *UserController) UpdateUserInfo(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的用户ID")
		return
	}

	userID := utils.GetContextUserID(ctx)
	if userID == 0 {
		utils.ResponseError(ctx, "未登录")
		return
	}

	// 检查是否为同一用户
	if uint(id) != userID {
		utils.ResponseError(ctx, "无权限修改其他用户信息")
		return
	}

	var req service.UpdateUserInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析更新用户信息请求失败", zap.Error(err))
		utils.ResponseError(ctx, "参数错误")
		return
	}

	if err := c.userService.UpdateUserInfo(userID, &req); err != nil {
		c.logger.Error("更新用户信息失败", zap.Error(err), zap.Uint("userId", userID))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("更新用户信息成功", zap.Uint("userId", userID))
	utils.ResponseSuccessWithMessage(ctx, "更新成功", nil)
}
