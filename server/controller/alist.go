package controller

import (
	"github.com/MccRay-s/alist2strm/service"
	"github.com/MccRay-s/alist2strm/utils/response"
	"github.com/gin-gonic/gin"
)

// AListController AList 控制器
type AListController struct{}

var AList = &AListController{}

// TestConnection 测试AList连接
func (a *AListController) TestConnection(c *gin.Context) {
	alistService := service.GetAListService()
	if alistService == nil {
		response.FailWithMessage(c, "AList 服务未初始化")
		return
	}

	if err := alistService.TestConnection(); err != nil {
		response.FailWithMessage(c, "连接测试失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "连接测试成功")
}
