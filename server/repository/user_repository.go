package repository

import (
	"errors"
	"time"

	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/model/user"
	userRequest "github.com/MccRay-s/alist2strm/model/user/request"
	"gorm.io/gorm"
)

type UserRepository struct{}

// 包级别的全局实例
var User = &UserRepository{}

// Create 创建用户
func (r *UserRepository) Create(user *user.User) error {
	return database.DB.Create(user).Error
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*user.User, error) {
	var u user.User
	err := database.DB.Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id uint) (*user.User, error) {
	var u user.User
	err := database.DB.Where("id = ?", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(user *user.User) error {
	return database.DB.Save(user).Error
}

// UpdateLastLoginAt 更新最后登录时间
func (r *UserRepository) UpdateLastLoginAt(id uint) error {
	now := time.Now()
	return database.DB.Model(&user.User{}).Where("id = ?", id).Update("last_login_at", &now).Error
}

// List 获取用户列表
func (r *UserRepository) List(req *userRequest.UserListReq) ([]user.User, int64, error) {
	var users []user.User
	var total int64

	query := database.DB.Model(&user.User{})

	// 添加状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 添加关键字搜索
	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Scopes(req.Paginate()).Find(&users).Error
	return users, total, err
}

// Delete 删除用户
func (r *UserRepository) Delete(id uint) error {
	return database.DB.Delete(&user.User{}, id).Error
}

// CheckUsernameExists 检查用户名是否存在
func (r *UserRepository) CheckUsernameExists(username string) (bool, error) {
	var count int64
	err := database.DB.Model(&user.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
