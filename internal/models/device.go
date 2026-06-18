package models

import (
	"time"

	"gorm.io/gorm"
)

type DeviceType string

const (
	DeviceTypeTemperatureValve DeviceType = "temperature_valve"
	DeviceTypeFan              DeviceType = "fan"
	DeviceTypeHumiditySpray    DeviceType = "humidity_spray"
)

const (
	DeviceStatusDisabled  = 0
	DeviceStatusEnabled   = 1
	DeviceStatusEmergency = 2
)

type Device struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	DeviceCode  string         `gorm:"size:50;uniqueIndex;not null" json:"device_code"`
	DeviceName  string         `gorm:"size:100;not null" json:"device_name"`
	DeviceType  DeviceType     `gorm:"size:30;not null;index" json:"device_type"`
	RoomID      uint           `gorm:"index;not null" json:"room_id"`
	RoomName    string         `gorm:"size:100" json:"room_name"`
	Status      int            `gorm:"default:1;comment:0-禁用 1-启用 2-异常停机" json:"status"`
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Device) TableName() string {
	return "devices"
}
