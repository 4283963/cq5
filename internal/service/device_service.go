package service

import (
	"errors"
	"incubator-backend/internal/models"
	"incubator-backend/internal/repository"
	"incubator-backend/pkg/logger"
)

type DeviceService interface {
	CreateDevice(device *models.Device) error
	UpdateDevice(device *models.Device) error
	DeleteDevice(id uint) error
	GetDeviceByID(id uint) (*models.Device, error)
	GetDeviceByCode(deviceCode string) (*models.Device, error)
	ListDevices(page, pageSize int, deviceType models.DeviceType, roomID uint, keyword string) ([]*models.Device, int64, error)
	ListAllDevices() ([]*models.Device, error)
}

type deviceService struct {
	deviceRepo repository.DeviceRepository
}

func NewDeviceService(deviceRepo repository.DeviceRepository) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
	}
}

func (s *deviceService) CreateDevice(device *models.Device) error {
	existing, err := s.deviceRepo.GetByCode(device.DeviceCode)
	if err == nil && existing != nil {
		return errors.New("设备编号已存在")
	}

	err = s.deviceRepo.Create(device)
	if err != nil {
		logger.Errorf("create device failed: %v", err)
		return err
	}

	logger.Infof("device created: %s", device.DeviceCode)
	return nil
}

func (s *deviceService) UpdateDevice(device *models.Device) error {
	existing, err := s.deviceRepo.GetByID(device.ID)
	if err != nil {
		return errors.New("设备不存在")
	}

	if existing.DeviceCode != device.DeviceCode {
		_, err := s.deviceRepo.GetByCode(device.DeviceCode)
		if err == nil {
			return errors.New("设备编号已存在")
		}
	}

	err = s.deviceRepo.Update(device)
	if err != nil {
		logger.Errorf("update device failed: %v", err)
		return err
	}

	logger.Infof("device updated: %s", device.DeviceCode)
	return nil
}

func (s *deviceService) DeleteDevice(id uint) error {
	_, err := s.deviceRepo.GetByID(id)
	if err != nil {
		return errors.New("设备不存在")
	}

	err = s.deviceRepo.Delete(id)
	if err != nil {
		logger.Errorf("delete device failed: %v", err)
		return err
	}

	logger.Infof("device deleted: %d", id)
	return nil
}

func (s *deviceService) GetDeviceByID(id uint) (*models.Device, error) {
	return s.deviceRepo.GetByID(id)
}

func (s *deviceService) GetDeviceByCode(deviceCode string) (*models.Device, error) {
	return s.deviceRepo.GetByCode(deviceCode)
}

func (s *deviceService) ListDevices(page, pageSize int, deviceType models.DeviceType, roomID uint, keyword string) ([]*models.Device, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.deviceRepo.List(page, pageSize, deviceType, roomID, keyword)
}

func (s *deviceService) ListAllDevices() ([]*models.Device, error) {
	return s.deviceRepo.ListAll()
}
