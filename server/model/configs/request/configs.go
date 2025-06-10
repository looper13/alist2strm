package request

import "github.com/MccRay-s/alist2strm/model/common/request"

// ConfigCreateReq 配置创建请求
type ConfigCreateReq struct {
	Name  string `json:"name" binding:"required" validate:"required,min=1,max=100" example:"配置名称"`
	Code  string `json:"code" binding:"required" validate:"required,min=1,max=50" example:"配置代码"`
	Value string `json:"value" binding:"required" validate:"required" example:"配置值"`
}

// ConfigUpdateReq 配置更新请求
type ConfigUpdateReq struct {
	ID    uint   `json:"-"` // 通过路径参数传递，不参与JSON绑定和验证
	Name  string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"配置名称"`
	Value string `json:"value,omitempty" validate:"omitempty" example:"配置值"`
}

// ConfigInfoReq 配置信息查询请求
type ConfigInfoReq struct {
	request.GetById
}

// ConfigByCodeReq 根据代码查询配置请求
type ConfigByCodeReq struct {
	Code string `json:"code" form:"code" binding:"required" validate:"required" example:"配置代码"`
}

// ConfigListReq 配置列表查询请求
type ConfigListReq struct {
	Name string `json:"name" form:"name" example:"配置名称筛选"`
	Code string `json:"code" form:"code" example:"配置代码筛选"`
}
