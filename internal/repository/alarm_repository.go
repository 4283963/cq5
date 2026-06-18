package repository

import (
	"incubator-backend/internal/models"
	"incubator-backend/pkg/database"
	"time"

	"gorm.io/gorm"
)

type AlarmRepository interface {
	Create(alarm *models.DeviceAlarm) error
	Update(alarm *models.DeviceAlarm) error
	GetByID(id uint) (*models.DeviceAlarm, error)
	GetActiveByDeviceID(deviceID uint) (*models.DeviceAlarm, error)
	HasActiveAlarm(deviceID uint, alarmType models.AlarmType) (bool, error)
	List(page, pageSize int, deviceID uint, alarmStatus models.AlarmStatus, alarmType models.AlarmType, startTime, endTime time.Time) ([]*models.DeviceAlarm, int64, error)
	Resolve(id uint, resolvedBy uint, remark string) error
	ResolveAllByDevice(deviceID uint, resolvedBy uint, remark string) error
}

type alarmRepository struct {
	db *gorm.DB
}

func NewAlarmRepository() AlarmRepository {
	return &alarmRepository{
		db: database.DB,
	}
}

func (r *alarmRepository) Create(alarm *models.DeviceAlarm) error {
	return r.db.Create(alarm).Error
}

func (r *alarmRepository) Update(alarm *models.DeviceAlarm) error {
	return r.db.Save(alarm).Error
}

func (r *alarmRepository) GetByID(id uint) (*models.DeviceAlarm, error) {
	var alarm models.DeviceAlarm
	err := r.db.First(&alarm, id).Error
	if err != nil {
		return nil, err
	}
	return &alarm, nil
}

func (r *alarmRepository) GetActiveByDeviceID(deviceID uint) (*models.DeviceAlarm, error) {
	var alarm models.DeviceAlarm
	err := r.db.Where("device_id = ? AND alarm_status = ?", deviceID, models.AlarmStatusActive).
		Order("triggered_at DESC").First(&alarm).Error
	if err != nil {
		return nil, err
	}
	return &alarm, nil
}

func (r *alarmRepository) HasActiveAlarm(deviceID uint, alarmType models.AlarmType) (bool, error) {
	var count int64
	err := r.db.Model(&models.DeviceAlarm{}).
		Where("device_id = ? AND alarm_type = ? AND alarm_status = ?", deviceID, alarmType, models.AlarmStatusActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *alarmRepository) List(page, pageSize int, deviceID uint, alarmStatus models.AlarmStatus, alarmType models.AlarmType, startTime, endTime time.Time) ([]*models.DeviceAlarm, int64, error) {
	var alarms []*models.DeviceAlarm
	var total int64

	query := r.db.Model(&models.DeviceAlarm{})

	if deviceID > 0 {
		query = query.Where("device_id = ?", deviceID)
	}
	if alarmStatus > 0 {
		query = query.Where("alarm_status = ?", alarmStatus)
	}
	if alarmType != "" {
		query = query.Where("alarm_type = ?", alarmType)
	}
	if !startTime.IsZero() {
		query = query.Where("triggered_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("triggered_at <= ?", endTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("triggered_at DESC").Find(&alarms).Error
	if err != nil {
		return nil, 0, err
	}

	return alarms, total, nil
}

func (r *alarmRepository) Resolve(id uint, resolvedBy uint, remark string) error {
	now := time.Now()
	return r.db.Model(&models.DeviceAlarm{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"alarm_status":   models.AlarmStatusResolved,
			"resolved_at":    &now,
			"resolved_by":    &resolvedBy,
			"resolve_remark": remark,
		}).Error
}

func (r *alarmRepository) ResolveAllByDevice(deviceID uint, resolvedBy uint, remark string) error {
	now := time.Now()
	return r.db.Model(&models.DeviceAlarm{}).
		Where("device_id = ? AND alarm_status = ?", deviceID, models.AlarmStatusActive).
		Updates(map[string]interface{}{
			"alarm_status":   models.AlarmStatusResolved,
			"resolved_at":    &now,
			"resolved_by":    &resolvedBy,
			"resolve_remark": remark,
		}).Error
}
