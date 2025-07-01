package repository

import (
	"errors"

	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/model/configs"
	configRequest "github.com/MccRay-s/alist2strm/model/configs/request"
	"gorm.io/gorm"
)

type ConfigRepository struct{}

// 包级别的全局实例
var Config = &ConfigRepository{}

// Create 创建配置
func (r *ConfigRepository) Create(config *configs.Config) error {
	return database.DB.Create(config).Error
}

// GetByID 根据ID获取配置
func (r *ConfigRepository) GetByID(id uint) (*configs.Config, error) {
	var config configs.Config
	err := database.DB.Where("id = ?", id).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetByCode 根据代码获取配置
func (r *ConfigRepository) GetByCode(code string) (*configs.Config, error) {
	var config configs.Config
	err := database.DB.Where("code = ?", code).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// Update 更新配置
func (r *ConfigRepository) Update(config *configs.Config) error {
	return database.DB.Save(config).Error
}

// Delete 删除配置
func (r *ConfigRepository) Delete(id uint) error {
	return database.DB.Delete(&configs.Config{}, id).Error
}

// List 获取配置列表
func (r *ConfigRepository) List(req *configRequest.ConfigListReq) ([]configs.Config, error) {
	var configList []configs.Config

	query := database.DB.Model(&configs.Config{})

	// 添加名称筛选
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 添加代码筛选
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+req.Code+"%")
	}

	// 查询所有数据，按创建时间排序
	err := query.Order("created_at DESC").Find(&configList).Error
	return configList, err
}

// CheckCodeExists 检查代码是否存在
func (r *ConfigRepository) CheckCodeExists(code string) (bool, error) {
	var count int64
	err := database.DB.Model(&configs.Config{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// CheckCodeExistsExcludeID 检查代码是否存在（排除指定ID）
func (r *ConfigRepository) CheckCodeExistsExcludeID(code string, excludeID uint) (bool, error) {
	var count int64
	err := database.DB.Model(&configs.Config{}).Where("code = ? AND id != ?", code, excludeID).Count(&count).Error
	return count > 0, err
}
