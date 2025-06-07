package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

// UpdateUserInfoRequest 更新用户信息请求
type UpdateUserInfoRequest struct {
	Nickname    string `json:"nickname,omitempty"`    // 新昵称
	Password    string `json:"password,omitempty"`    // 新密码
	OldPassword string `json:"oldPassword,omitempty"` // 旧密码
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserInfo struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname,omitempty"`
	Email       string    `json:"email,omitempty"`
	Status      string    `json:"status"`
	LastLoginAt time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserResponse UserInfo // 为了保持兼容性

type UserService struct {
	db     *gorm.DB
	logger *zap.Logger
}

var (
	userService *UserService
	userOnce    sync.Once
)

// GetUserService 获取 UserService 单例
func GetUserService() *UserService {
	userOnce.Do(func() {
		userService = &UserService{
			db:     model.DB,
			logger: utils.Logger,
		}
	})
	return userService
}

// Register 用户注册
func (s *UserService) Register(req *RegisterRequest) error {
	utils.Info("开始用户注册",
		zap.String("username", req.Username),
		zap.String("nickname", req.Nickname))

	var existingUser model.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		utils.Warn("注册失败：用户名已存在",
			zap.String("username", req.Username))
		return errors.New("用户名已存在")
	}

	user := model.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Status:   "active",
	}

	if err := user.SetPassword(req.Password); err != nil {
		utils.Error("密码加密失败",
			zap.String("username", req.Username),
			zap.Error(err))
		return err
	}

	if err := s.db.Create(&user).Error; err != nil {
		utils.Error("创建用户失败",
			zap.String("username", req.Username),
			zap.Error(err))
		return err
	}

	utils.Info("用户注册成功",
		zap.String("username", req.Username),
		zap.Uint("userId", user.ID))
	return nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	utils.Info("用户登录尝试",
		zap.String("username", req.Username))

	var user model.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.Warn("登录失败：用户不存在",
			zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	if !user.CheckPassword(req.Password) {
		utils.Warn("登录失败：密码错误",
			zap.String("username", req.Username))
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	user.LastLoginAt = time.Now()
	if err := s.db.Save(&user).Error; err != nil {
		utils.Error("更新最后登录时间失败",
			zap.String("username", req.Username),
			zap.Error(err))
		return nil, err
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Nickname)
	if err != nil {
		utils.Error("生成Token失败",
			zap.String("username", req.Username),
			zap.Error(err))
		return nil, err
	}

	utils.Info("用户登录成功",
		zap.String("username", req.Username),
		zap.Uint("userId", user.ID))

	return &LoginResponse{
		Token: token,
		User: UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Nickname:    user.Nickname,
			Status:      user.Status,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(userID uint) (*UserInfo, error) {
	utils.Info("获取用户信息", zap.Uint("userId", userID))

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		utils.Error("获取用户信息失败：用户不存在",
			zap.Uint("userId", userID),
			zap.Error(err))
		return nil, errors.New("用户不存在")
	}

	utils.Debug("用户信息获取成功",
		zap.Uint("userId", userID),
		zap.String("username", user.Username))

	return &UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// HasAnyUser 检查是否存在任何用户
func (s *UserService) HasAnyUser() bool {
	var count int64
	s.db.Model(&model.User{}).Count(&count)
	return count > 0
}

// CreateDefaultUser 创建默认用户如果不存在任何用户
func (s *UserService) CreateDefaultUser(username, password string) error {
	// 首先检查是否有任何用户
	if s.HasAnyUser() {
		s.logger.Info("已存在用户，跳过创建默认用户")
		return nil
	}

	s.logger.Info("系统中没有用户，准备创建默认用户")

	// 如果密码为空，生成随机密码
	if password == "" {
		password = utils.GenerateRandomPassword()
		s.logger.Info("已为默认用户生成随机密码，请及时修改",
			zap.String("username", username),
			zap.String("password", password))
	}

	// 创建默认用户
	user := &model.User{
		Username: username,
		Status:   "active",
		Nickname: "Admin",
	}

	if err := user.SetPassword(password); err != nil {
		s.logger.Error("设置用户密码失败",
			zap.String("username", username),
			zap.Error(err))
		return err
	}

	if err := s.db.Create(user).Error; err != nil {
		s.logger.Error("创建默认用户失败",
			zap.String("username", username),
			zap.Error(err))
		return err
	}

	s.logger.Info("成功创建默认用户",
		zap.String("username", username),
		zap.String("password", password))
	return nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID uint, req *UpdateUserInfoRequest) (*UserInfo, error) {
	s.logger.Info("更新用户信息",
		zap.Uint("userId", userID))

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		s.logger.Error("更新失败：用户不存在",
			zap.Uint("userId", userID),
			zap.Error(err))
		return nil, errors.New("用户不存在")
	}

	// 更新昵称
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	// 处理密码更新
	if req.Password != "" && req.OldPassword != "" {
		// 验证旧密码
		if !user.CheckPassword(req.OldPassword) {
			s.logger.Error("更新失败：旧密码验证失败",
				zap.Uint("userId", userID))
			return nil, errors.New("旧密码验证失败")
		}

		// 更新新密码
		if err := user.SetPassword(req.Password); err != nil {
			s.logger.Error("更新失败：密码设置失败",
				zap.Uint("userId", userID),
				zap.Error(err))
			return nil, errors.New("密码设置失败")
		}
		s.logger.Info("用户密码更新成功",
			zap.Uint("userId", userID))
	}

	if err := s.db.Save(&user).Error; err != nil {
		s.logger.Error("更新用户信息失败",
			zap.Uint("userId", userID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("用户信息更新成功",
		zap.Uint("userId", userID))

	return &UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
	}, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(userID uint, req *UpdateUserInfoRequest) error {
	utils.Info("开始更新用户信息", zap.Uint("userId", userID))

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		utils.Error("更新失败：用户不存在",
			zap.Uint("userId", userID),
			zap.Error(err))
		return errors.New("用户不存在")
	}

	// 如果要更新密码，验证旧密码
	if req.Password != "" {
		if req.OldPassword == "" {
			return errors.New("更新密码时必须提供原密码")
		}
		if !user.CheckPassword(req.OldPassword) {
			return errors.New("原密码错误")
		}
		if err := user.SetPassword(req.Password); err != nil {
			utils.Error("密码加密失败",
				zap.Uint("userId", userID),
				zap.Error(err))
			return err
		}
	}

	// 更新昵称
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	if err := s.db.Save(&user).Error; err != nil {
		utils.Error("更新用户信息失败",
			zap.Uint("userId", userID),
			zap.Error(err))
		return err
	}

	utils.Info("用户信息更新成功",
		zap.Uint("userId", userID))
	return nil
}

// UpdateUserInfo 更新用户信息（静态方法）
func UpdateUserInfo(userID uint, req *UpdateUserInfoRequest) (*UserInfo, error) {
	utils.Info("更新用户信息",
		zap.Uint("userId", userID))

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		utils.Error("更新失败：用户不存在",
			zap.Uint("userId", userID),
			zap.Error(err))
		return nil, errors.New("用户不存在")
	}

	// 更新昵称
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	// 处理密码更新
	if req.Password != "" && req.OldPassword != "" {
		// 验证旧密码
		if !user.CheckPassword(req.OldPassword) {
			utils.Error("更新失败：旧密码验证失败",
				zap.Uint("userId", userID))
			return nil, errors.New("旧密码验证失败")
		}

		// 更新新密码
		if err := user.SetPassword(req.Password); err != nil {
			utils.Error("更新失败：密码设置失败",
				zap.Uint("userId", userID),
				zap.Error(err))
			return nil, errors.New("密码设置失败")
		}
		utils.Info("用户密码更新成功",
			zap.Uint("userId", userID))
	}

	if err := model.DB.Save(&user).Error; err != nil {
		utils.Error("更新用户信息失败",
			zap.Uint("userId", userID),
			zap.Error(err))
		return nil, err
	}

	utils.Info("用户信息更新成功",
		zap.Uint("userId", userID))

	return &UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
	}, nil
}
