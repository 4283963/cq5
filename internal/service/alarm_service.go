package service

import (
	"errors"
	"incubator-backend/internal/models"
	"incubator-backend/internal/repository"
	"incubator-backend/pkg/logger"
	"time"
)

type AlarmService interface {
	GetAlarmByID(id uint) (*models.DeviceAlarm, error)
	GetActiveAlarmByDevice(deviceID uint) (*models.DeviceAlarm, error)
	ListAlarms(page, pageSize int, deviceID uint, alarmStatus models.AlarmStatus, alarmType models.AlarmType, startTime, endTime time.Time) ([]*models.DeviceAlarm, int64, error)
	ResolveAlarm(id uint, resolvedBy uint, remark string) error
	ResolveAllByDevice(deviceID uint, resolvedBy uint, remark string) error
}

type alarmService struct {
	alarmRepo repository.AlarmRepository
}

func NewAlarmService(alarmRepo repository.AlarmRepository) AlarmService {
	return &alarmService{
		alarmRepo: alarmRepo,
	}
}

func (s *alarmService) GetAlarmByID(id uint) (*models.DeviceAlarm, error) {
	return s.alarmRepo.GetByID(id)
}

func (s *alarmService) GetActiveAlarmByDevice(deviceID uint) (*models.DeviceAlarm, error) {
	return s.alarmRepo.GetActiveByDeviceID(deviceID)
}

func (s *alarmService) ListAlarms(page, pageSize int, deviceID uint, alarmStatus models.AlarmStatus, alarmType models.AlarmType, startTime, endTime time.Time) ([]*models.DeviceAlarm, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.alarmRepo.List(page, pageSize, deviceID, alarmStatus, alarmType, startTime, endTime)
}

func (s *alarmService) ResolveAlarm(id uint, resolvedBy uint, remark string) error {
	alarm, err := s.alarmRepo.GetByID(id)
	if err != nil {
		return errors.New("报警记录不存在")
	}
	if alarm.AlarmStatus == models.AlarmStatusResolved {
		return errors.New("该报警已解除")
	}

	err = s.alarmRepo.Resolve(id, resolvedBy, remark)
	if err != nil {
		logger.Errorf("resolve alarm failed: %v", err)
		return errors.New("解除报警失败")
	}

	logger.Infof("alarm resolved: id=%d, by=%d, remark=%s", id, resolvedBy, remark)
	return nil
}

func (s *alarmService) ResolveAllByDevice(deviceID uint, resolvedBy uint, remark string) error {
	err := s.alarmRepo.ResolveAllByDevice(deviceID, resolvedBy, remark)
	if err != nil {
		logger.Errorf("resolve all alarms by device failed: %v", err)
		return errors.New("批量解除报警失败")
	}
	return nil
}
