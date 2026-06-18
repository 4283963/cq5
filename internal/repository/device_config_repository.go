package repository

import (
	"incubator-backend/internal/models"
	"incubator-backend/pkg/database"

	"gorm.io/gorm"
)

type DeviceConfigRepository interface {
	Create(config *models.DeviceConfig) error
	Update(config *models.DeviceConfig) error
	Delete(id uint) error
	GetByID(id uint) (*models.DeviceConfig, error)
	GetByDeviceID(deviceID uint) (*models.DeviceConfig, error)
	GetByDeviceCode(deviceCode string) (*models.DeviceConfig, error)
	List(page, pageSize int, deviceType models.DeviceType) ([]*models.DeviceConfig, int64, error)
}

type deviceConfigRepository struct {
	db *gorm.DB
}

func NewDeviceConfigRepository() DeviceConfigRepository {
	return &deviceConfigRepository{
		db: database.DB,
	}
}

func (r *deviceConfigRepository) Create(config *models.DeviceConfig) error {
	return r.db.Create(config).Error
}

func (r *deviceConfigRepository) Update(config *models.DeviceConfig) error {
	config.Version++
	return r.db.Save(config).Error
}

func (r *deviceConfigRepository) Delete(id uint) error {
	return r.db.Delete(&models.DeviceConfig{}, id).Error
}

func (r *deviceConfigRepository) GetByID(id uint) (*models.DeviceConfig, error) {
	var config models.DeviceConfig
	err := r.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *deviceConfigRepository) GetByDeviceID(deviceID uint) (*models.DeviceConfig, error) {
	var config models.DeviceConfig
	err := r.db.Where("device_id = ?", deviceID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *deviceConfigRepository) GetByDeviceCode(deviceCode string) (*models.DeviceConfig, error) {
	var config models.DeviceConfig
	err := r.db.Where("device_code = ?", deviceCode).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *deviceConfigRepository) List(page, pageSize int, deviceType models.DeviceType) ([]*models.DeviceConfig, int64, error) {
	var configs []*models.DeviceConfig
	var total int64

	query := r.db.Model(&models.DeviceConfig{})

	if deviceType != "" {
		query = query.Joins("JOIN devices ON devices.id = device_configs.device_id").
			Where("devices.device_type = ?", deviceType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("device_configs.id DESC").Find(&configs).Error
	if err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}
