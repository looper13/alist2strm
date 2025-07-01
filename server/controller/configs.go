package controller

import (
	"strconv"

	"github.com/MccRay-s/alist2strm/model/common/response"
	configRequest "github.com/MccRay-s/alist2strm/model/configs/request"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/MccRay-s/alist2strm/utils"
	"github.com/gin-gonic/gin"
)

// 包级别的配置控制器实例
var Config = &ConfigController{}

type ConfigController struct{}

// Create 创建配置
func (cc *ConfigController) Create(c *gin.Context) {
	var req configRequest.ConfigCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("创建配置参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	err := service.Config.Create(&req)
	if err != nil {
		utils.Error("创建配置失败", "name", req.Name, "code", req.Code, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("创建配置成功", "name", req.Name, "code", req.Code, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("创建成功", c)
}

// GetConfigInfo 获取配置信息
func (cc *ConfigController) GetConfigInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("获取配置信息ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("配置ID参数错误", c)
		return
	}

	req := &configRequest.ConfigInfoReq{}
	req.ID = id

	configInfo, err := service.Config.GetConfigInfo(req)
	if err != nil {
		utils.Error("获取配置信息失败", "config_id", id, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取配置信息成功", "config_id", id, "request_id", c.GetString("request_id"))
	response.SuccessWithData(configInfo, c)
}

// GetConfigByCode 根据代码获取配置
func (cc *ConfigController) GetConfigByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.Error("获取配置信息代码参数为空", "request_id", c.GetString("request_id"))
		response.FailWithMessage("配置代码参数不能为空", c)
		return
	}

	req := &configRequest.ConfigByCodeReq{
		Code: code,
	}

	configInfo, err := service.Config.GetConfigByCode(req)
	if err != nil {
		utils.Error("根据代码获取配置失败", "code", code, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("根据代码获取配置成功", "code", code, "request_id", c.GetString("request_id"))
	response.SuccessWithData(configInfo, c)
}

// UpdateConfig 更新配置
func (cc *ConfigController) UpdateConfig(c *gin.Context) {
	// 从路径参数获取配置ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("更新配置ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("配置ID参数错误", c)
		return
	}

	var req configRequest.ConfigUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error("更新配置参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 设置配置ID
	req.ID = uint(id)

	err = service.Config.UpdateConfig(&req)
	if err != nil {
		utils.Error("更新配置失败", "config_id", req.ID, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("更新配置成功", "config_id", req.ID, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("更新成功", c)
}

// DeleteConfig 删除配置
func (cc *ConfigController) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error("删除配置ID参数错误", "id", idStr, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("配置ID参数错误", c)
		return
	}

	err = service.Config.DeleteConfig(uint(id))
	if err != nil {
		utils.Error("删除配置失败", "config_id", id, "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("删除配置成功", "config_id", id, "request_id", c.GetString("request_id"))
	response.SuccessWithMessage("删除成功", c)
}

// GetConfigList 获取配置列表
func (cc *ConfigController) GetConfigList(c *gin.Context) {
	var req configRequest.ConfigListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error("获取配置列表参数绑定失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	configList, err := service.Config.GetConfigList(&req)
	if err != nil {
		utils.Error("获取配置列表失败", "error", err.Error(), "request_id", c.GetString("request_id"))
		response.FailWithMessage(err.Error(), c)
		return
	}

	utils.Info("获取配置列表成功", "total", len(configList), "request_id", c.GetString("request_id"))
	response.SuccessWithData(configList, c)
}
