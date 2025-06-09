package request

// UserLoginReq 用户登录请求
type UserLoginReq struct {
	Username string `json:"username" binding:"required" validate:"required,min=3,max=20"`
	Password string `json:"password" binding:"required" validate:"required,min=6,max=32"`
}

// UserRegisterReq 用户注册请求
type UserRegisterReq struct {
	Username string `json:"username" binding:"required" validate:"required,min=3,max=20"`
	Password string `json:"password" binding:"required" validate:"required,min=6,max=32"`
	Nickname string `json:"nickname" validate:"max=50"`
}

// UserUpdateReq 用户更新请求
type UserUpdateReq struct {
	ID          uint   `json:"id" binding:"required"`
	Nickname    string `json:"nickname,omitempty" validate:"omitempty,max=50"`
	OldPassword string `json:"oldPassword,omitempty" validate:"omitempty,min=6,max=32"`
	NewPassword string `json:"newPassword,omitempty" validate:"omitempty,min=6,max=32"`
}

// UserInfoReq 用户信息查询请求
type UserInfoReq struct {
	GetById
}

// UserListReq 用户列表查询请求
type UserListReq struct {
	PageInfo
	Status string `json:"status" form:"status"` // 用户状态筛选
}
