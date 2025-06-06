package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Status      string    `json:"status"`
	LastLoginAt time.Time `json:"lastLoginAt"`
	Token       string    `json:"token,omitempty"`
}

type UserService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserService(db *gorm.DB, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// GetUserByUsername 通过用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建新用户
func (s *UserService) Create(user *model.User) error {
	return s.db.Create(user).Error
}

func Register(req *RegisterRequest) error {
	utils.Info("开始用户注册",
		zap.String("username", req.Username),
		zap.String("nickname", req.Nickname))

	var existingUser model.User
	if err := model.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
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

	if err := model.DB.Create(&user).Error; err != nil {
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

func Login(req *LoginRequest) (*UserResponse, error) {
	utils.Info("用户登录尝试",
		zap.String("username", req.Username))

	var user model.User
	if err := model.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
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
	if err := model.DB.Save(&user).Error; err != nil {
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

	return &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		Token:       token,
	}, nil
}

func GetUserInfo(userID uint) (*UserResponse, error) {
	utils.Info("获取用户信息", zap.Uint("userId", userID))

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		utils.Error("获取用户信息失败：用户不存在",
			zap.Uint("userId", userID),
			zap.Error(err))
		return nil, errors.New("用户不存在")
	}

	utils.Debug("用户信息获取成功",
		zap.Uint("userId", userID),
		zap.String("username", user.Username))

	return &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
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

	if err := s.Create(user); err != nil {
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
