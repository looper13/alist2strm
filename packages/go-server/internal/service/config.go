package service

import (
	"alist2strm/internal/model"
	"alist2strm/internal/utils"
	"errors"

	"go.uber.org/zap"
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

// CreateConfig 创建配置
func CreateConfig(req *ConfigRequest) (*ConfigResponse, error) {
	utils.Info("开始创建配置",
		zap.String("name", req.Name),
		zap.String("code", req.Code))

	var existingConfig model.Config
	if err := model.DB.Where("code = ?", req.Code).First(&existingConfig).Error; err == nil {
		utils.Warn("创建配置失败：配置代码已存在",
			zap.String("code", req.Code))
		return nil, errors.New("配置代码已存在")
	}

	config := model.Config{
		Name:  req.Name,
		Code:  req.Code,
		Value: req.Value,
	}

	if err := model.DB.Create(&config).Error; err != nil {
		utils.Error("创建配置失败",
			zap.String("code", req.Code),
			zap.Error(err))
		return nil, err
	}

	utils.Info("配置创建成功",
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
func UpdateConfig(id uint, req *ConfigRequest) (*ConfigResponse, error) {
	utils.Info("开始更新配置",
		zap.Uint("configId", id),
		zap.String("code", req.Code))

	var config model.Config
	if err := model.DB.First(&config, id).Error; err != nil {
		utils.Warn("更新配置失败：配置不存在",
			zap.Uint("configId", id))
		return nil, errors.New("配置不存在")
	}

	// 检查新的 code 是否与其他配置冲突
	if config.Code != req.Code {
		var existingConfig model.Config
		if err := model.DB.Where("code = ? AND id != ?", req.Code, id).First(&existingConfig).Error; err == nil {
			utils.Warn("更新配置失败：新的配置代码已存在",
				zap.String("code", req.Code))
			return nil, errors.New("配置代码已存在")
		}
	}

	config.Name = req.Name
	config.Code = req.Code
	config.Value = req.Value

	if err := model.DB.Save(&config).Error; err != nil {
		utils.Error("更新配置失败",
			zap.Uint("configId", id),
			zap.Error(err))
		return nil, err
	}

	utils.Info("配置更新成功",
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
func DeleteConfig(id uint) error {
	utils.Info("开始删除配置", zap.Uint("configId", id))

	result := model.DB.Delete(&model.Config{}, id)
	if result.Error != nil {
		utils.Error("删除配置失败",
			zap.Uint("configId", id),
			zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		utils.Warn("删除配置失败：配置不存在",
			zap.Uint("configId", id))
		return errors.New("配置不存在")
	}

	utils.Info("配置删除成功", zap.Uint("configId", id))
	return nil
}

// GetConfig 获取单个配置
func GetConfig(id uint) (*ConfigResponse, error) {
	var config model.Config
	if err := model.DB.First(&config, id).Error; err != nil {
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
func GetConfigByCode(code string) (*ConfigResponse, error) {
	var config model.Config
	if err := model.DB.Where("code = ?", code).First(&config).Error; err != nil {
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
func ListConfigs() ([]ConfigResponse, error) {
	var configs []model.Config
	if err := model.DB.Find(&configs).Error; err != nil {
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
