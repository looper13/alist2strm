package service

import (
	"errors"
	"time"

	"github.com/MccRay-s/alist2strm/config"
	"github.com/MccRay-s/alist2strm/model/user"
	userRequest "github.com/MccRay-s/alist2strm/model/user/request"
	userResponse "github.com/MccRay-s/alist2strm/model/user/response"
	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/utils"
)

type UserService struct{}

// 包级别的全局实例
var User = &UserService{}

// Login 用户登录
func (s *UserService) Login(req *userRequest.UserLoginReq) (*userResponse.UserLoginResp, error) {
	// 根据用户名查找用户
	user, err := repository.User.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户已被禁用")
	}

	// 更新最后登录时间
	if err := repository.User.UpdateLastLoginAt(user.ID); err != nil {
		// 记录错误但不影响登录流程
		utils.Error("更新用户最后登录时间失败", "user_id", user.ID, "error", err.Error())
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	// 构建响应
	resp := &userResponse.UserLoginResp{
		User: userResponse.UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Nickname:    user.Nickname,
			Status:      user.Status,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			LastLoginAt: user.LastLoginAt,
		},
		Token: token,
	}

	return resp, nil
}

// Register 用户注册
func (s *UserService) Register(req *userRequest.UserRegisterReq) error {
	// 检查用户名是否已存在
	exists, err := repository.User.CheckUsernameExists(req.Username)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 设置默认昵称
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	// 创建用户
	newUser := &user.User{
		Username: req.Username,
		Password: hashedPassword,
		Nickname: nickname,
		Status:   "active",
	}

	return repository.User.Create(newUser)
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(req *userRequest.UserInfoReq) (*userResponse.UserInfo, error) {
	user, err := repository.User.GetByID(uint(req.ID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	resp := &userResponse.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}

	return resp, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(req *userRequest.UserUpdateReq) error {
	// 获取用户信息
	user, err := repository.User.GetByID(req.ID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 判断是修改昵称还是密码
	if req.Nickname != "" {
		// 修改昵称
		user.Nickname = req.Nickname
		user.UpdatedAt = time.Now()
	}

	if req.OldPassword != "" && req.NewPassword != "" {
		// 修改密码
		// 验证旧密码
		if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
			return errors.New("原密码错误")
		}

		// 加密新密码
		hashedPassword, err := utils.HashPassword(req.NewPassword)
		if err != nil {
			return errors.New("新密码加密失败")
		}

		user.Password = hashedPassword
		user.UpdatedAt = time.Now()
	}

	// 如果既没有昵称也没有密码更新，返回错误
	if req.Nickname == "" && (req.OldPassword == "" || req.NewPassword == "") {
		return errors.New("请提供要更新的信息")
	}

	return repository.User.Update(user)
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(req *userRequest.UserListReq) (*userResponse.UserListResp, error) {
	users, total, err := repository.User.List(req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	userInfos := make([]userResponse.UserInfo, len(users))
	for i, u := range users {
		userInfos[i] = userResponse.UserInfo{
			ID:          u.ID,
			Username:    u.Username,
			Nickname:    u.Nickname,
			Status:      u.Status,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			LastLoginAt: u.LastLoginAt,
		}
	}

	resp := &userResponse.UserListResp{
		List:     userInfos,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	return resp, nil
}

// InitializeDefaultUser 初始化默认用户
func (s *UserService) InitializeDefaultUser() error {
	// 检查是否已有用户
	count, err := repository.User.CountUsers()
	if err != nil {
		return err
	}

	// 如果已有用户，则不需要创建
	if count > 0 {
		return nil
	}

	// 获取配置中的用户信息
	cfg := config.GlobalConfig
	if cfg == nil {
		return errors.New("配置未初始化")
	}

	username := cfg.User.Name
	password := cfg.User.Password

	// 如果用户名为空，使用默认值
	if username == "" {
		username = "admin"
	}

	// 如果密码为空，生成随机密码
	if password == "" {
		password = utils.GenerateRandomPassword(12)
		utils.Info("==============================================")
		utils.Info("🔐 系统已自动创建默认管理员账户")
		utils.Info("👤 用户名: " + username)
		utils.Info("🔑 密码: " + password)
		utils.Info("⚠️  请妥善保存密码，首次登录后建议修改密码")
		utils.Info("==============================================")
	} else {
		utils.Info("使用配置文件中的密码创建默认管理员账户", "username", username)
	}

	// 创建默认用户
	req := &userRequest.UserRegisterReq{
		Username: username,
		Password: password,
		Nickname: "系统管理员",
	}

	return s.Register(req)
}
