package controller

import (
	"strconv"

	"github.com/MccRay-s/alist2strm/model/common/response"
	fileHistoryRequest "github.com/MccRay-s/alist2strm/model/filehistory/request"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/gin-gonic/gin"
)

type FileHistoryController struct{}

// GetMainFileList 获取主文件分页列表
func (c *FileHistoryController) GetMainFileList(ctx *gin.Context) {
	var req fileHistoryRequest.FileHistoryListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), ctx)
		return
	}

	result, err := service.FileHistoryServiceApp.GetMainFileList(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}

	response.SuccessWithData(result, ctx)
}

// GetFileHistoryInfo 获取文件历史详情
func (c *FileHistoryController) GetFileHistoryInfo(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", ctx)
		return
	}

	result, err := service.FileHistoryServiceApp.GetFileHistoryInfo(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}

	response.SuccessWithData(result, ctx)
}

// GetRelatedFilesByMainID 根据主文件ID查询关联文件
func (c *FileHistoryController) GetRelatedFilesByMainID(ctx *gin.Context) {
	var req fileHistoryRequest.FileHistoryByMainFileReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage("参数错误", ctx)
		return
	}

	result, err := service.FileHistoryServiceApp.GetRelatedFilesByMainID(req.MainFileID)
	if err != nil {
		response.FailWithMessage(err.Error(), ctx)
		return
	}

	response.SuccessWithData(result, ctx)
}
