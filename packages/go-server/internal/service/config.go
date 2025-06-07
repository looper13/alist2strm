package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"errors"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ConfigService struct {
	db     *gorm.DB
	logger *zap.Logger
}

var (
	configService *ConfigService
	configOnce    sync.Once
)

type ConfigRequest struct {
	Name  string `json:"name" binding:"required"`
	Code  string `json:"code" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type ConfigResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Code  string `json:"code"`
	Value string `json:"value"`
}

// GetConfigService 获取 ConfigService 单例
func GetConfigService() *ConfigService {
	configOnce.Do(func() {
		configService = &ConfigService{
			db:     model.DB,
			logger: utils.Logger,
		}
	})
	return configService
}

// CreateConfig 创建配置
func (s *ConfigService) CreateConfig(req *ConfigRequest) (*ConfigResponse, error) {
	s.logger.Info("开始创建配置",
		zap.String("name", req.Name),
		zap.String("code", req.Code))

	var existingConfig model.Config
	if err := s.db.Where("code = ?", req.Code).First(&existingConfig).Error; err == nil {
		s.logger.Warn("创建配置失败：配置代码已存在",
			zap.String("code", req.Code))
		return nil, errors.New("配置代码已存在")
	}

	config := model.Config{
		Name:  req.Name,
		Code:  req.Code,
		Value: req.Value,
	}

	if err := s.db.Create(&config).Error; err != nil {
		s.logger.Error("创建配置失败",
			zap.String("code", req.Code),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("配置创建成功",
		zap.String("code", req.Code),
		zap.Uint("configId", config.ID))

	return &ConfigResponse{
		ID:    config.ID,
		Name:  config.Name,
		Code:  config.Code,
		Value: config.Value,
	}, nil
}

// UpdateConfig 更新配置
func (s *ConfigService) UpdateConfig(id uint, req *ConfigRequest) (*ConfigResponse, error) {
	s.logger.Info("开始更新配置",
		zap.Uint("configId", id),
		zap.String("code", req.Code))

	var config model.Config
	if err := s.db.First(&config, id).Error; err != nil {
		s.logger.Warn("更新配置失败：配置不存在",
			zap.Uint("configId", id))
		return nil, errors.New("配置不存在")
	}

	// 检查新的 code 是否与其他配置冲突
	if config.Code != req.Code {
		var existingConfig model.Config
		if err := s.db.Where("code = ? AND id != ?", req.Code, id).First(&existingConfig).Error; err == nil {
			s.logger.Warn("更新配置失败：新的配置代码已存在",
				zap.String("code", req.Code))
			return nil, errors.New("配置代码已存在")
		}
	}

	config.Name = req.Name
	config.Code = req.Code
	config.Value = req.Value

	if err := s.db.Save(&config).Error; err != nil {
		s.logger.Error("更新配置失败",
			zap.Uint("configId", id),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("配置更新成功",
		zap.Uint("configId", id),
		zap.String("code", config.Code))

	return &ConfigResponse{
		ID:    config.ID,
		Name:  config.Name,
		Code:  config.Code,
		Value: config.Value,
	}, nil
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(id uint) error {
	s.logger.Info("开始删除配置", zap.Uint("configId", id))

	result := s.db.Delete(&model.Config{}, id)
	if result.Error != nil {
		s.logger.Error("删除配置失败",
			zap.Uint("configId", id),
			zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		s.logger.Warn("删除配置失败：配置不存在",
			zap.Uint("configId", id))
		return errors.New("配置不存在")
	}

	s.logger.Info("配置删除成功", zap.Uint("configId", id))
	return nil
}

// GetConfig 获取单个配置
func (s *ConfigService) GetConfig(id uint) (*ConfigResponse, error) {
	var config model.Config
	if err := s.db.First(&config, id).Error; err != nil {
		s.logger.Error("获取配置失败：配置不存在",
			zap.Uint("configId", id),
			zap.Error(err))
		return nil, errors.New("配置不存在")
	}

	return &ConfigResponse{
		ID:    config.ID,
		Name:  config.Name,
		Code:  config.Code,
		Value: config.Value,
	}, nil
}

// GetConfigByCode 通过代码获取配置
func (s *ConfigService) GetConfigByCode(code string) (*ConfigResponse, error) {
	var config model.Config
	if err := s.db.Where("code = ?", code).First(&config).Error; err != nil {
		s.logger.Error("获取配置失败：配置不存在",
			zap.String("code", code),
			zap.Error(err))
		return nil, errors.New("配置不存在")
	}

	return &ConfigResponse{
		ID:    config.ID,
		Name:  config.Name,
		Code:  config.Code,
		Value: config.Value,
	}, nil
}

// ListConfigs 获取配置列表
func (s *ConfigService) ListConfigs() ([]ConfigResponse, error) {
	var configs []model.Config
	if err := s.db.Find(&configs).Error; err != nil {
		s.logger.Error("获取配置列表失败", zap.Error(err))
		return nil, err
	}

	response := make([]ConfigResponse, len(configs))
	for i, config := range configs {
		response[i] = ConfigResponse{
			ID:    config.ID,
			Name:  config.Name,
			Code:  config.Code,
			Value: config.Value,
		}
	}

	return response, nil
}
