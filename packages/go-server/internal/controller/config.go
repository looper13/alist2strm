package controller

import (
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ConfigController struct {
	configService *service.ConfigService
	logger        *zap.Logger
}

// NewConfigController 创建配置控制器
func NewConfigController() *ConfigController {
	return &ConfigController{
		configService: service.GetConfigService(),
		logger:        utils.Logger,
	}
}

// RegisterRoutes 注册路由
func (c *ConfigController) RegisterRoutes(router *gin.RouterGroup) {
	configGroup := router.Group("/configs")
	{
		configGroup.POST("", c.CreateConfig)
		configGroup.PUT("/:id", c.UpdateConfig)
		configGroup.DELETE("/:id", c.DeleteConfig)
		configGroup.GET("/:id", c.GetConfig)
		configGroup.GET("/code/:code", c.GetConfigByCode)
		configGroup.GET("/all", c.ListConfigs)
	}
}

// CreateConfig 创建配置
func (c *ConfigController) CreateConfig(ctx *gin.Context) {
	var req service.ConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析创建配置请求失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	resp, err := c.configService.CreateConfig(&req)
	if err != nil {
		c.logger.Error("创建配置失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("创建配置成功", zap.String("name", req.Name))
	utils.ResponseSuccess(ctx, resp)
}

// UpdateConfig 更新配置
func (c *ConfigController) UpdateConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的配置ID")
		return
	}

	var req service.ConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("解析更新配置请求失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	resp, err := c.configService.UpdateConfig(uint(id), &req)
	if err != nil {
		c.logger.Error("更新配置失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("更新配置成功", zap.Uint("id", uint(id)), zap.String("name", req.Name))
	utils.ResponseSuccess(ctx, resp)
}

// DeleteConfig 删除配置
func (c *ConfigController) DeleteConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的配置ID")
		return
	}

	if err := c.configService.DeleteConfig(uint(id)); err != nil {
		c.logger.Error("删除配置失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	c.logger.Info("删除配置成功", zap.Uint("id", uint(id)))
	utils.ResponseSuccessWithMessage(ctx, "配置删除成功", nil)
}

// GetConfig 获取单个配置
func (c *ConfigController) GetConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(ctx, "无效的配置ID")
		return
	}

	resp, err := c.configService.GetConfig(uint(id))
	if err != nil {
		c.logger.Error("获取配置失败", zap.Error(err), zap.Uint("id", uint(id)))
		utils.ResponseError(ctx, err.Error())
		return
	}

	utils.ResponseSuccess(ctx, resp)
}

// GetConfigByCode 通过代码获取配置
func (c *ConfigController) GetConfigByCode(ctx *gin.Context) {
	code := ctx.Param("code")
	resp, err := c.configService.GetConfigByCode(code)
	if err != nil {
		c.logger.Error("通过代码获取配置失败", zap.Error(err), zap.String("code", code))
		utils.ResponseError(ctx, err.Error())
		return
	}

	utils.ResponseSuccess(ctx, resp)
}

// ListConfigs 获取配置列表
func (c *ConfigController) ListConfigs(ctx *gin.Context) {
	resp, err := c.configService.ListConfigs()
	if err != nil {
		c.logger.Error("获取配置列表失败", zap.Error(err))
		utils.ResponseError(ctx, err.Error())
		return
	}

	utils.ResponseSuccess(ctx, resp)
}
