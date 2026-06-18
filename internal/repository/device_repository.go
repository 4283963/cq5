package repository

import (
	"incubator-backend/internal/models"
	"incubator-backend/pkg/database"

	"gorm.io/gorm"
)

type DeviceRepository interface {
	Create(device *models.Device) error
	Update(device *models.Device) error
	Delete(id uint) error
	GetByID(id uint) (*models.Device, error)
	GetByCode(deviceCode string) (*models.Device, error)
	List(page, pageSize int, deviceType models.DeviceType, roomID uint, keyword string) ([]*models.Device, int64, error)
	ListAll() ([]*models.Device, error)
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository() DeviceRepository {
	return &deviceRepository{
		db: database.DB,
	}
}

func (r *deviceRepository) Create(device *models.Device) error {
	return r.db.Create(device).Error
}

func (r *deviceRepository) Update(device *models.Device) error {
	return r.db.Save(device).Error
}

func (r *deviceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Device{}, id).Error
}

func (r *deviceRepository) GetByID(id uint) (*models.Device, error) {
	var device models.Device
	err := r.db.First(&device, id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) GetByCode(deviceCode string) (*models.Device, error) {
	var device models.Device
	err := r.db.Where("device_code = ?", deviceCode).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) List(page, pageSize int, deviceType models.DeviceType, roomID uint, keyword string) ([]*models.Device, int64, error) {
	var devices []*models.Device
	var total int64

	query := r.db.Model(&models.Device{})

	if deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}
	if roomID > 0 {
		query = query.Where("room_id = ?", roomID)
	}
	if keyword != "" {
		query = query.Where("device_name LIKE ? OR device_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&devices).Error
	if err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

func (r *deviceRepository) ListAll() ([]*models.Device, error) {
	var devices []*models.Device
	err := r.db.Where("status = ?", 1).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}
