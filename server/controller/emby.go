package controller

import (
	"strconv"

	"github.com/MccRay-s/alist2strm/model/common/response"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/gin-gonic/gin"
)

// EmbyController Emby 控制器
type EmbyController struct{}

// Emby 控制器实例
var Emby = &EmbyController{}

// GetLibraries 获取Emby媒体库列表
// @Summary 获取Emby媒体库列表
// @Description 获取Emby服务器中所有可用的媒体库
// @Tags Emby
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]service.EmbyLibrary}
// @Failure 400 {object} response.Response
// @Router /api/emby/libraries [get]
func (ctrl *EmbyController) GetLibraries(c *gin.Context) {
	libraries, err := service.Emby.GetLibraries()
	if err != nil {
		response.FailWithMessage("获取Emby媒体库列表失败: "+err.Error(), c)
		return
	}
	response.SuccessWithData(libraries, c)
}

// RefreshLibrary 刷新指定Emby媒体库
// @Summary 刷新指定Emby媒体库
// @Description 触发刷新Emby中的特定媒体库
// @Tags Emby
// @Accept json
// @Produce json
// @Param id path string true "媒体库ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/emby/libraries/{id}/refresh [post]
func (ctrl *EmbyController) RefreshLibrary(c *gin.Context) {
	// 从路径参数获取媒体库ID
	libraryID := c.Param("id")
	if libraryID == "" {
		response.FailWithMessage("媒体库ID不能为空", c)
		return
	}

	err := service.Emby.RefreshLibrary(libraryID)
	if err != nil {
		response.FailWithMessage("刷新媒体库失败: "+err.Error(), c)
		return
	}
	response.SuccessWithMessage("已成功触发媒体库刷新", c)
}

// RefreshAllLibraries 刷新所有Emby媒体库
// @Summary 刷新所有Emby媒体库
// @Description 触发刷新Emby中的所有媒体库
// @Tags Emby
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/emby/libraries/refresh [post]
func (ctrl *EmbyController) RefreshAllLibraries(c *gin.Context) {
	err := service.Emby.RefreshAllLibraries()
	if err != nil {
		response.FailWithMessage("刷新所有媒体库失败: "+err.Error(), c)
		return
	}
	response.SuccessWithMessage("已成功触发所有媒体库刷新", c)
}

// GetLatestMedia 获取Emby最新入库媒体
// @Summary 获取Emby最新入库媒体
// @Description 获取Emby服务器中最新添加的媒体文件
// @Tags Emby
// @Accept json
// @Produce json
// @Param limit query int false "返回结果数量限制，默认10条"
// @Success 200 {object} response.Response{data=[]service.EmbyLatestMedia}
// @Failure 400 {object} response.Response
// @Router /api/emby/latest [get]
func (ctrl *EmbyController) GetLatestMedia(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	media, err := service.Emby.GetLatestMedia(limit)
	if err != nil {
		response.FailWithMessage("获取最新入库媒体失败: "+err.Error(), c)
		return
	}
	response.SuccessWithData(media, c)
}

// TestConnection 测试Emby服务器连接
// @Summary 测试Emby服务器连接
// @Description 测试Emby服务器的可用性和连接状态
// @Tags Emby
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=response.EmbyConnectionTestResult}
// @Failure 400 {object} response.Response
// @Router /api/emby/test [get]
func (ctrl *EmbyController) TestConnection(c *gin.Context) {
	result, err := service.Emby.TestConnection()
	if err != nil {
		response.FailWithMessage("测试连接时发生错误: "+err.Error(), c)
		return
	}
	response.SuccessWithData(result, c)
}

// GetImage 获取Emby图片
// @Summary 获取Emby图片
// @Description 代理获取Emby服务器上的图片资源
// @Tags Emby
// @Accept json
// @Produce image/*
// @Param item_id path string true "项目ID"
// @Param image_type path string true "图片类型,例如:Primary,Backdrop等"
// @Param tag query string false "图片标签"
// @Param max_width query int false "最大宽度"
// @Param max_height query int false "最大高度"
// @Param quality query int false "图片质量"
// @Success 200 {file} binary "图片文件"
// @Failure 400 {object} response.Response
// @Router /api/emby/items/{item_id}/images/{image_type} [get]
func (ctrl *EmbyController) GetImage(c *gin.Context) {
	// 获取路径参数
	itemId := c.Param("item_id")
	imageType := c.Param("image_type")
	if itemId == "" || imageType == "" {
		response.FailWithMessage("项目ID和图片类型不能为空", c)
		return
	}

	// 获取查询参数
	tag := c.Query("tag")
	maxWidth, _ := strconv.Atoi(c.Query("max_width"))
	maxHeight, _ := strconv.Atoi(c.Query("max_height"))
	quality, _ := strconv.Atoi(c.Query("quality"))

	// 调用服务获取图片
	imageData, contentType, err := service.Emby.GetImage(itemId, imageType, tag, maxWidth, maxHeight, quality)
	if err != nil {
		response.FailWithMessage("获取图片失败: "+err.Error(), c)
		return
	}

	// 设置响应头
	c.Header("Content-Disposition", "inline")
	c.Data(200, contentType, imageData)
}
