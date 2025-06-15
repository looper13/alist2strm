package controller

import (
	"github.com/MccRay-s/alist2strm/model/common/response"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/gin-gonic/gin"
)

// AListController AList 控制器
type AListController struct{}

var AList = &AListController{}

// TestConnection 测试AList连接
func (a *AListController) TestConnection(c *gin.Context) {
	alistService := service.GetAListService()
	if alistService == nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	if err := alistService.TestConnection(); err != nil {
		response.FailWithMessage("连接测试失败: "+err.Error(), c)
		return
	}
	response.SuccessWithMessage("连接测试成功", c)
}
