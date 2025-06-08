package response

// UserLoginResponse 用户登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	ReFreshToken string `json:"refreshToken"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Status      string `json:"status"`
	LastLoginAt string `json:"lastLoginAt"`
}
