package handler

import (
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateConfig 创建配置
func CreateConfig(c *gin.Context) {
	var req service.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	resp, err := service.CreateConfig(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(resp))
}

// UpdateConfig 更新配置
func UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "无效的配置ID"))
		return
	}

	var req service.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	resp, err := service.UpdateConfig(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(resp))
}

// DeleteConfig 删除配置
func DeleteConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "无效的配置ID"))
		return
	}

	if err := service.DeleteConfig(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse("配置删除成功"))
}

// GetConfig 获取单个配置
func GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, "无效的配置ID"))
		return
	}

	resp, err := service.GetConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(resp))
}

// GetConfigByCode 通过代码获取配置
func GetConfigByCode(c *gin.Context) {
	code := c.Param("code")
	resp, err := service.GetConfigByCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(resp))
}

// ListConfigs 获取配置列表
func ListConfigs(c *gin.Context) {
	resp, err := service.ListConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(resp))
}
