package request

import "github.com/MccRay-s/alist2strm/model/common/request"

// UserLoginReq 用户登录请求
type UserLoginReq struct {
	Username string `json:"username" binding:"required" validate:"required,min=3,max=20" example:"用户名"`
	Password string `json:"password" binding:"required" validate:"required,min=6,max=32" example:"密码"`
}

// UserRegisterReq 用户注册请求
type UserRegisterReq struct {
	Username string `json:"username" binding:"required" validate:"required,min=3,max=20" example:"用户名"`
	Password string `json:"password" binding:"required" validate:"required,min=6,max=32" example:"密码"`
	Nickname string `json:"nickname" validate:"max=50" example:"昵称"`
}

// UserUpdateReq 用户更新请求
type UserUpdateReq struct {
	ID          uint   `json:"-"` // 通过路径参数传递，不参与JSON绑定和验证
	Nickname    string `json:"nickname,omitempty" validate:"omitempty,max=50" example:"昵称"`
	OldPassword string `json:"oldPassword,omitempty" validate:"omitempty,min=6,max=32" example:"旧密码"`
	NewPassword string `json:"newPassword,omitempty" validate:"omitempty,min=6,max=32" example:"新密码"`
}

// UserInfoReq 用户信息查询请求
type UserInfoReq struct {
	request.GetById
}

// UserListReq 用户列表查询请求
type UserListReq struct {
	request.PageInfo
	Status string `json:"status" form:"status" example:"用户状态筛选"`
}

// 保持向后兼容的别名
type Login = UserLoginReq
type UpdateUser = UserUpdateReq
