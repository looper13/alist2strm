package request

type Login struct {
	Username string `json:"username" example:"用户名"`
	Password string `json:"password" example:"密码"`
}

type UpdateUser struct {
	Nickname    string `json:"nickname" example:"昵称"`
	OldPassword string `json:"oldPassword" example:"旧密码"`
	NewPassword string `json:"newPassword" example:"新密码"`
}
