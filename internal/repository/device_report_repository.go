package repository

import (
	"incubator-backend/internal/models"
	"incubator-backend/pkg/database"
	"time"

	"gorm.io/gorm"
)

type DeviceReportRepository interface {
	Create(report *models.DeviceReport) error
	GetByID(id uint) (*models.DeviceReport, error)
	GetLatestByDeviceID(deviceID uint) (*models.DeviceReport, error)
	GetLatestByDeviceCode(deviceCode string) (*models.DeviceReport, error)
	ListByDevice(deviceID uint, startTime, endTime time.Time, page, pageSize int) ([]*models.DeviceReport, int64, error)
}

type deviceReportRepository struct {
	db *gorm.DB
}

func NewDeviceReportRepository() DeviceReportRepository {
	return &deviceReportRepository{
		db: database.DB,
	}
}

func (r *deviceReportRepository) Create(report *models.DeviceReport) error {
	return r.db.Create(report).Error
}

func (r *deviceReportRepository) GetByID(id uint) (*models.DeviceReport, error) {
	var report models.DeviceReport
	err := r.db.First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *deviceReportRepository) GetLatestByDeviceID(deviceID uint) (*models.DeviceReport, error) {
	var report models.DeviceReport
	err := r.db.Where("device_id = ?", deviceID).Order("report_time DESC").First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *deviceReportRepository) GetLatestByDeviceCode(deviceCode string) (*models.DeviceReport, error) {
	var report models.DeviceReport
	err := r.db.Where("device_code = ?", deviceCode).Order("report_time DESC").First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *deviceReportRepository) ListByDevice(deviceID uint, startTime, endTime time.Time, page, pageSize int) ([]*models.DeviceReport, int64, error) {
	var reports []*models.DeviceReport
	var total int64

	query := r.db.Model(&models.DeviceReport{}).Where("device_id = ?", deviceID)

	if !startTime.IsZero() {
		query = query.Where("report_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("report_time <= ?", endTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("report_time DESC").Find(&reports).Error
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}
