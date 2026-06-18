package service

import (
	"errors"
	"fmt"
	"incubator-backend/internal/cache"
	"incubator-backend/internal/models"
	"incubator-backend/internal/repository"
	"incubator-backend/pkg/logger"
	"time"
)

type DeviceReportService interface {
	CreateReport(report *models.DeviceReport) error
	GetReportByID(id uint) (*models.DeviceReport, error)
	GetLatestReport(deviceID uint) (*models.DeviceReport, error)
	GetLatestReportByCode(deviceCode string) (*models.DeviceReport, error)
	ListReports(deviceID uint, startTime, endTime time.Time, page, pageSize int) ([]*models.DeviceReport, int64, error)
	ReportDeviceData(deviceCode string, data *DeviceReportData) (*SafetyCheckResult, error)
	GetCachedStatus(deviceCode string) (*cache.DeviceStatusCache, error)
	SyncAllStatusFromDB() error
}

type DeviceReportData struct {
	ValveOpen       *float64  `json:"valve_open"`
	FanSpeed        *int      `json:"fan_speed"`
	CurrentTemp     *float64  `json:"current_temp"`
	CurrentHumidity *float64  `json:"current_humidity"`
	SprayStatus     *int      `json:"spray_status"`
	ReportTime      time.Time `json:"report_time"`
}

type deviceReportService struct {
	reportRepo    repository.DeviceReportRepository
	deviceRepo    repository.DeviceRepository
	statusCache   cache.DeviceStatusCacheManager
	safetyService SafetyService
}

func NewDeviceReportService(
	reportRepo repository.DeviceReportRepository,
	deviceRepo repository.DeviceRepository,
	statusCache cache.DeviceStatusCacheManager,
	safetyService SafetyService,
) DeviceReportService {
	return &deviceReportService{
		reportRepo:    reportRepo,
		deviceRepo:    deviceRepo,
		statusCache:   statusCache,
		safetyService: safetyService,
	}
}

func (s *deviceReportService) CreateReport(report *models.DeviceReport) error {
	device, err := s.deviceRepo.GetByID(report.DeviceID)
	if err != nil {
		return errors.New("设备不存在")
	}

	report.DeviceCode = device.DeviceCode
	if report.ReportTime.IsZero() {
		report.ReportTime = time.Now()
	}

	err = s.reportRepo.Create(report)
	if err != nil {
		logger.Errorf("create device report failed: %v", err)
		return err
	}

	err = s.statusCache.UpdateFromReport(report)
	if err != nil {
		logger.Warnf("update status cache failed: %v", err)
	}

	return nil
}

func (s *deviceReportService) GetReportByID(id uint) (*models.DeviceReport, error) {
	return s.reportRepo.GetByID(id)
}

func (s *deviceReportService) GetLatestReport(deviceID uint) (*models.DeviceReport, error) {
	return s.reportRepo.GetLatestByDeviceID(deviceID)
}

func (s *deviceReportService) GetLatestReportByCode(deviceCode string) (*models.DeviceReport, error) {
	return s.reportRepo.GetLatestByDeviceCode(deviceCode)
}

func (s *deviceReportService) ListReports(deviceID uint, startTime, endTime time.Time, page, pageSize int) ([]*models.DeviceReport, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.reportRepo.ListByDevice(deviceID, startTime, endTime, page, pageSize)
}

func (s *deviceReportService) ReportDeviceData(deviceCode string, data *DeviceReportData) (*SafetyCheckResult, error) {
	device, err := s.deviceRepo.GetByCode(deviceCode)
	if err != nil {
		logger.Errorf("device not found for report: %s", deviceCode)
		return nil, errors.New("设备不存在")
	}

	if device.Status == models.DeviceStatusDisabled {
		return nil, errors.New("设备已禁用")
	}

	if device.Status == models.DeviceStatusEmergency {
		return nil, fmt.Errorf("设备已处于异常停机状态，请先解除报警")
	}

	checkResult, err := s.safetyService.CheckDeviceReport(device, data)
	if err != nil {
		logger.Errorf("safety check failed for device %s: %v", deviceCode, err)
	}

	reportTime := data.ReportTime
	if reportTime.IsZero() {
		reportTime = time.Now()
	}

	report := &models.DeviceReport{
		DeviceID:        device.ID,
		DeviceCode:      deviceCode,
		ValveOpen:       data.ValveOpen,
		FanSpeed:        data.FanSpeed,
		CurrentTemp:     data.CurrentTemp,
		CurrentHumidity: data.CurrentHumidity,
		SprayStatus:     data.SprayStatus,
		ReportTime:      reportTime,
	}

	err = s.reportRepo.Create(report)
	if err != nil {
		logger.Errorf("save device report failed: %v", err)
		return nil, err
	}

	err = s.statusCache.UpdateFromReport(report)
	if err != nil {
		logger.Warnf("sync report to cache failed: %v", err)
	}

	if checkResult != nil && checkResult.Triggered {
		_, shutdownErr := s.safetyService.TriggerEmergencyShutdown(device, checkResult)
		if shutdownErr != nil {
			logger.Errorf("emergency shutdown failed for device %s: %v", deviceCode, shutdownErr)
		}
	}

	logger.Debugf("device data reported: %s, valve_open=%v, fan_speed=%v, temp=%v",
		deviceCode, data.ValveOpen, data.FanSpeed, data.CurrentTemp)

	return checkResult, nil
}

func (s *deviceReportService) GetCachedStatus(deviceCode string) (*cache.DeviceStatusCache, error) {
	cached, err := s.statusCache.Get(deviceCode)
	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	report, err := s.reportRepo.GetLatestByDeviceCode(deviceCode)
	if err != nil {
		return nil, err
	}

	if report == nil {
		return nil, nil
	}

	err = s.statusCache.UpdateFromReport(report)
	if err != nil {
		logger.Warnf("sync status to cache failed: %v", err)
	}

	return s.statusCache.Get(deviceCode)
}

func (s *deviceReportService) SyncAllStatusFromDB() error {
	devices, err := s.deviceRepo.ListAll()
	if err != nil {
		logger.Errorf("list all devices failed: %v", err)
		return err
	}

	successCount := 0
	failCount := 0

	for _, device := range devices {
		report, err := s.reportRepo.GetLatestByDeviceID(device.ID)
		if err != nil {
			failCount++
			logger.Warnf("get latest report failed for device %s: %v", device.DeviceCode, err)
			continue
		}

		if report == nil {
			continue
		}

		err = s.statusCache.UpdateFromReport(report)
		if err != nil {
			failCount++
			logger.Warnf("sync status to cache failed for device %s: %v", device.DeviceCode, err)
			continue
		}

		successCount++
	}

	logger.Infof("sync all status from db completed: success=%d, fail=%d", successCount, failCount)
	return nil
}
