package service

import (
	"errors"

	"github.com/MccRay-s/alist2strm/model/configs"
	configRequest "github.com/MccRay-s/alist2strm/model/configs/request"
	configResponse "github.com/MccRay-s/alist2strm/model/configs/response"
	"github.com/MccRay-s/alist2strm/repository"
)

type ConfigService struct{}

// 包级别的全局实例
var Config = &ConfigService{}

// Create 创建配置
func (s *ConfigService) Create(req *configRequest.ConfigCreateReq) error {
	// 检查代码是否已存在
	exists, err := repository.Config.CheckCodeExists(req.Code)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("配置代码已存在")
	}

	// 创建配置
	newConfig := &configs.Config{
		Name:  req.Name,
		Code:  req.Code,
		Value: req.Value,
	}

	return repository.Config.Create(newConfig)
}

// GetConfigInfo 获取配置信息
func (s *ConfigService) GetConfigInfo(req *configRequest.ConfigInfoReq) (*configResponse.ConfigInfo, error) {
	config, err := repository.Config.GetByID(uint(req.ID))
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("配置不存在")
	}

	resp := &configResponse.ConfigInfo{
		ID:        config.ID,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
		Name:      config.Name,
		Code:      config.Code,
		Value:     config.Value,
	}

	return resp, nil
}

// GetConfigByCode 根据代码获取配置
func (s *ConfigService) GetConfigByCode(req *configRequest.ConfigByCodeReq) (*configResponse.ConfigInfo, error) {
	config, err := repository.Config.GetByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("配置不存在")
	}

	resp := &configResponse.ConfigInfo{
		ID:        config.ID,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
		Name:      config.Name,
		Code:      config.Code,
		Value:     config.Value,
	}

	return resp, nil
}

// UpdateConfig 更新配置
func (s *ConfigService) UpdateConfig(req *configRequest.ConfigUpdateReq) error {
	// 获取配置信息
	config, err := repository.Config.GetByID(req.ID)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("配置不存在")
	}

	// 更新名称
	if req.Name != "" {
		config.Name = req.Name
	}

	// 更新值
	if req.Value != "" {
		config.Value = req.Value
	}

	// 如果没有任何更新，返回错误
	if req.Name == "" && req.Value == "" {
		return errors.New("请提供要更新的信息")
	}

	return repository.Config.Update(config)
}

// DeleteConfig 删除配置
func (s *ConfigService) DeleteConfig(id uint) error {
	// 检查配置是否存在
	config, err := repository.Config.GetByID(id)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("配置不存在")
	}

	return repository.Config.Delete(id)
}

// GetConfigList 获取配置列表
func (s *ConfigService) GetConfigList(req *configRequest.ConfigListReq) ([]configResponse.ConfigInfo, error) {
	configList, err := repository.Config.List(req)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	configInfos := make([]configResponse.ConfigInfo, len(configList))
	for i, c := range configList {
		configInfos[i] = configResponse.ConfigInfo{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Name:      c.Name,
			Code:      c.Code,
			Value:     c.Value,
		}
	}

	return configInfos, nil
}
