package models

import (
	"time"

	"gorm.io/gorm"
)

type AlarmType string

const (
	AlarmTypeTempOverLimit AlarmType = "temp_over_limit"
	AlarmTypeFanIdle       AlarmType = "fan_idle"
)

type AlarmLevel string

const (
	AlarmLevelWarning  AlarmLevel = "warning"
	AlarmLevelCritical AlarmLevel = "critical"
)

type AlarmStatus int

const (
	AlarmStatusActive   = 1
	AlarmStatusResolved = 2
	AlarmStatusIgnored  = 3
)

type DeviceAlarm struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	DeviceID      uint           `gorm:"index;not null" json:"device_id"`
	DeviceCode    string         `gorm:"size:50;index;not null" json:"device_code"`
	DeviceName    string         `gorm:"size:100" json:"device_name"`
	RoomID        uint           `gorm:"index" json:"room_id"`
	RoomName      string         `gorm:"size:100" json:"room_name"`
	AlarmType     AlarmType      `gorm:"size:30;index;not null" json:"alarm_type"`
	AlarmLevel    AlarmLevel     `gorm:"size:20;not null" json:"alarm_level"`
	AlarmStatus   AlarmStatus    `gorm:"default:1;index;comment:1-未处理 2-已解除 3-已忽略" json:"alarm_status"`
	TriggerValue  string         `gorm:"size:100" json:"trigger_value"`
	Threshold     string         `gorm:"size:100" json:"threshold"`
	Description   string         `gorm:"size:500" json:"description"`
	TriggeredAt   time.Time      `gorm:"index;not null" json:"triggered_at"`
	ResolvedAt    *time.Time     `json:"resolved_at"`
	ResolvedBy    *uint          `json:"resolved_by"`
	ResolveRemark string         `gorm:"size:500" json:"resolve_remark"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DeviceAlarm) TableName() string {
	return "device_alarms"
}
