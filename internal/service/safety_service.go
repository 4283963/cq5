package service

import (
	"fmt"
	"incubator-backend/internal/models"
	"incubator-backend/internal/repository"
	"incubator-backend/pkg/logger"
	"strconv"
	"time"
)

const (
	MaxAllowedTemperature = 42.0
	FanIdleSpeedThreshold = 0
)

type SafetyCheckResult struct {
	Triggered    bool
	AlarmType    models.AlarmType
	AlarmLevel   models.AlarmLevel
	TriggerValue string
	Threshold    string
	Description  string
}

type SafetyService interface {
	CheckDeviceReport(device *models.Device, data *DeviceReportData) (*SafetyCheckResult, error)
	TriggerEmergencyShutdown(device *models.Device, result *SafetyCheckResult) (*models.DeviceAlarm, error)
	RecoverDevice(deviceID uint, resolvedBy uint, remark string) error
}

type safetyService struct {
	alarmRepo  repository.AlarmRepository
	deviceRepo repository.DeviceRepository
}

func NewSafetyService(
	alarmRepo repository.AlarmRepository,
	deviceRepo repository.DeviceRepository,
) SafetyService {
	return &safetyService{
		alarmRepo:  alarmRepo,
		deviceRepo: deviceRepo,
	}
}

func (s *safetyService) CheckDeviceReport(device *models.Device, data *DeviceReportData) (*SafetyCheckResult, error) {
	if device.Status == models.DeviceStatusEmergency {
		return nil, nil
	}

	if data.CurrentTemp != nil && *data.CurrentTemp > MaxAllowedTemperature {
		return &SafetyCheckResult{
			Triggered:    true,
			AlarmType:    models.AlarmTypeTempOverLimit,
			AlarmLevel:   models.AlarmLevelCritical,
			TriggerValue: fmt.Sprintf("%.2f℃", *data.CurrentTemp),
			Threshold:    fmt.Sprintf("%.2f℃", MaxAllowedTemperature),
			Description:  fmt.Sprintf("温度超限: 当前 %.2f℃ 超过阈值 %.2f℃，已触发紧急断电保护", *data.CurrentTemp, MaxAllowedTemperature),
		}, nil
	}

	if data.FanSpeed != nil && device.DeviceType == models.DeviceTypeFan && *data.FanSpeed <= FanIdleSpeedThreshold {
		return &SafetyCheckResult{
			Triggered:    true,
			AlarmType:    models.AlarmTypeFanIdle,
			AlarmLevel:   models.AlarmLevelCritical,
			TriggerValue: strconv.Itoa(*data.FanSpeed) + " rpm",
			Threshold:    fmt.Sprintf("> %d rpm", FanIdleSpeedThreshold),
			Description:  "风机空转: 转速为 0，可能已停转或损坏，已触发紧急断电保护",
		}, nil
	}

	return &SafetyCheckResult{
		Triggered: false,
	}, nil
}

func (s *safetyService) TriggerEmergencyShutdown(device *models.Device, result *SafetyCheckResult) (*models.DeviceAlarm, error) {
	hasActive, err := s.alarmRepo.HasActiveAlarm(device.ID, result.AlarmType)
	if err != nil {
		logger.Errorf("check active alarm failed: %v", err)
	}
	if hasActive {
		logger.Warnf("device %s already has active alarm %s, skip duplicate", device.DeviceCode, result.AlarmType)
		return nil, nil
	}

	alarm := &models.DeviceAlarm{
		DeviceID:     device.ID,
		DeviceCode:   device.DeviceCode,
		DeviceName:   device.DeviceName,
		RoomID:       device.RoomID,
		RoomName:     device.RoomName,
		AlarmType:    result.AlarmType,
		AlarmLevel:   result.AlarmLevel,
		AlarmStatus:  models.AlarmStatusActive,
		TriggerValue: result.TriggerValue,
		Threshold:    result.Threshold,
		Description:  result.Description,
		TriggeredAt:  time.Now(),
	}

	err = s.alarmRepo.Create(alarm)
	if err != nil {
		logger.Errorf("create alarm record failed: %v", err)
		return nil, err
	}

	device.Status = models.DeviceStatusEmergency
	err = s.deviceRepo.Update(device)
	if err != nil {
		logger.Errorf("update device status to emergency failed: %v", err)
		return alarm, err
	}

	logger.Errorf("EMERGENCY SHUTDOWN triggered: device=%s(%s), alarm_type=%s, desc=%s",
		device.DeviceCode, device.DeviceName, result.AlarmType, result.Description)

	return alarm, nil
}

func (s *safetyService) RecoverDevice(deviceID uint, resolvedBy uint, remark string) error {
	device, err := s.deviceRepo.GetByID(deviceID)
	if err != nil {
		return fmt.Errorf("设备不存在")
	}

	if device.Status != models.DeviceStatusEmergency {
		return fmt.Errorf("设备当前非异常停机状态")
	}

	err = s.alarmRepo.ResolveAllByDevice(deviceID, resolvedBy, remark)
	if err != nil {
		logger.Errorf("resolve device alarms failed: %v", err)
		return fmt.Errorf("解除报警记录失败")
	}

	device.Status = models.DeviceStatusEnabled
	err = s.deviceRepo.Update(device)
	if err != nil {
		logger.Errorf("recover device status failed: %v", err)
		return fmt.Errorf("恢复设备状态失败")
	}

	logger.Infof("device recovered: id=%d, code=%s, by=%d, remark=%s", deviceID, device.DeviceCode, resolvedBy, remark)
	return nil
}
