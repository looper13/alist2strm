package response

import "time"

// UserLoginResp 用户登录响应
type UserLoginResp struct {
	User  UserInfo `json:"user"`
	Token string   `json:"token"`
}

// UserInfo 用户信息响应
type UserInfo struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
}

// UserListResp 用户列表响应
type UserListResp struct {
	List     []UserInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// 保持向后兼容的别名
type LoginResponse = UserLoginResp
type UserInfoResponse = UserInfo
