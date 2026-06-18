package service

import (
	"errors"
	"incubator-backend/internal/cache"
	"incubator-backend/internal/models"
	"incubator-backend/internal/repository"
	"incubator-backend/pkg/logger"
)

type DeviceConfigService interface {
	CreateConfig(config *models.DeviceConfig) error
	UpdateConfig(config *models.DeviceConfig) error
	DeleteConfig(id uint) error
	GetConfigByID(id uint) (*models.DeviceConfig, error)
	GetConfigByDeviceID(deviceID uint) (*models.DeviceConfig, error)
	GetConfigByDeviceCode(deviceCode string) (*models.DeviceConfig, error)
	ListConfigs(page, pageSize int, deviceType models.DeviceType) ([]*models.DeviceConfig, int64, error)
	SyncConfigToCache(deviceCode string) error
	GetCachedConfig(deviceCode string) (*cache.DeviceConfigCache, error)
}

type deviceConfigService struct {
	configRepo  repository.DeviceConfigRepository
	deviceRepo  repository.DeviceRepository
	configCache cache.DeviceConfigCacheManager
}

func NewDeviceConfigService(
	configRepo repository.DeviceConfigRepository,
	deviceRepo repository.DeviceRepository,
	configCache cache.DeviceConfigCacheManager,
) DeviceConfigService {
	return &deviceConfigService{
		configRepo:  configRepo,
		deviceRepo:  deviceRepo,
		configCache: configCache,
	}
}

func (s *deviceConfigService) CreateConfig(config *models.DeviceConfig) error {
	device, err := s.deviceRepo.GetByID(config.DeviceID)
	if err != nil {
		return errors.New("设备不存在")
	}

	existing, err := s.configRepo.GetByDeviceID(config.DeviceID)
	if err == nil && existing != nil {
		return errors.New("该设备配置已存在，请使用更新接口")
	}

	config.DeviceCode = device.DeviceCode
	config.Version = 1

	err = s.configRepo.Create(config)
	if err != nil {
		logger.Errorf("create device config failed: %v", err)
		return err
	}

	err = s.SyncConfigToCache(device.DeviceCode)
	if err != nil {
		logger.Warnf("sync config to cache failed: %v", err)
	}

	logger.Infof("device config created: device_id=%d", config.DeviceID)
	return nil
}

func (s *deviceConfigService) UpdateConfig(config *models.DeviceConfig) error {
	existing, err := s.configRepo.GetByID(config.ID)
	if err != nil {
		return errors.New("配置不存在")
	}

	device, err := s.deviceRepo.GetByID(existing.DeviceID)
	if err != nil {
		return errors.New("关联设备不存在")
	}

	config.DeviceID = existing.DeviceID
	config.DeviceCode = existing.DeviceCode
	config.Version = existing.Version

	err = s.configRepo.Update(config)
	if err != nil {
		logger.Errorf("update device config failed: %v", err)
		return err
	}

	err = s.SyncConfigToCache(device.DeviceCode)
	if err != nil {
		logger.Warnf("sync config to cache failed: %v", err)
	}

	logger.Infof("device config updated: device_id=%d, version=%d", config.DeviceID, config.Version)
	return nil
}

func (s *deviceConfigService) DeleteConfig(id uint) error {
	config, err := s.configRepo.GetByID(id)
	if err != nil {
		return errors.New("配置不存在")
	}

	err = s.configRepo.Delete(id)
	if err != nil {
		logger.Errorf("delete device config failed: %v", err)
		return err
	}

	err = s.configCache.Delete(config.DeviceCode)
	if err != nil {
		logger.Warnf("delete config cache failed: %v", err)
	}

	logger.Infof("device config deleted: %d", id)
	return nil
}

func (s *deviceConfigService) GetConfigByID(id uint) (*models.DeviceConfig, error) {
	return s.configRepo.GetByID(id)
}

func (s *deviceConfigService) GetConfigByDeviceID(deviceID uint) (*models.DeviceConfig, error) {
	return s.configRepo.GetByDeviceID(deviceID)
}

func (s *deviceConfigService) GetConfigByDeviceCode(deviceCode string) (*models.DeviceConfig, error) {
	return s.configRepo.GetByDeviceCode(deviceCode)
}

func (s *deviceConfigService) ListConfigs(page, pageSize int, deviceType models.DeviceType) ([]*models.DeviceConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.configRepo.List(page, pageSize, deviceType)
}

func (s *deviceConfigService) SyncConfigToCache(deviceCode string) error {
	config, err := s.configRepo.GetByDeviceCode(deviceCode)
	if err != nil {
		logger.Errorf("get config for cache sync failed: %v", err)
		return err
	}

	err = s.configCache.UpdateFromConfig(config)
	if err != nil {
		logger.Errorf("update config cache failed: %v", err)
		return err
	}

	logger.Debugf("config synced to cache: %s", deviceCode)
	return nil
}

func (s *deviceConfigService) GetCachedConfig(deviceCode string) (*cache.DeviceConfigCache, error) {
	cached, err := s.configCache.Get(deviceCode)
	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	err = s.SyncConfigToCache(deviceCode)
	if err != nil {
		return nil, err
	}

	return s.configCache.Get(deviceCode)
}
