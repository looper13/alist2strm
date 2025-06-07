package handler

import (
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateConfig 创建配置
func CreateConfig(c *gin.Context) {
	var req service.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("解析创建配置请求失败", zap.Error(err))
		utils.ResponseError(c, err.Error())
		return
	}

	resp, err := service.GetConfigService().CreateConfig(&req)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, resp)
}

// UpdateConfig 更新配置
func UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的配置ID")
		return
	}

	var req service.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("解析更新配置请求失败", zap.Error(err))
		utils.ResponseError(c, err.Error())
		return
	}

	resp, err := service.GetConfigService().UpdateConfig(uint(id), &req)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, resp)
}

// DeleteConfig 删除配置
func DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的配置ID")
		return
	}

	if err := service.GetConfigService().DeleteConfig(uint(id)); err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessage(c, "配置删除成功", nil)
}

// GetConfig 获取单个配置
func GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ResponseError(c, "无效的配置ID")
		return
	}

	resp, err := service.GetConfigService().GetConfig(uint(id))
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, resp)
}

// GetConfigByCode 通过代码获取配置
func GetConfigByCode(c *gin.Context) {
	code := c.Param("code")
	resp, err := service.GetConfigService().GetConfigByCode(code)
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, resp)
}

// ListConfigs 获取配置列表
func ListConfigs(c *gin.Context) {
	resp, err := service.GetConfigService().ListConfigs()
	if err != nil {
		utils.ResponseError(c, err.Error())
		return
	}

	utils.ResponseSuccess(c, resp)
}
