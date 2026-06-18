package models

import (
	"time"

	"gorm.io/gorm"
)

type DeviceReport struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	DeviceID        uint           `gorm:"index;not null" json:"device_id"`
	DeviceCode      string         `gorm:"size:50;index" json:"device_code"`
	ValveOpen       *float64       `gorm:"type:decimal(5,2);comment:执行器开度(%)" json:"valve_open"`
	FanSpeed        *int           `gorm:"comment:当前转速(rpm)" json:"fan_speed"`
	CurrentTemp     *float64       `gorm:"type:decimal(5,2);comment:当前温度(℃)" json:"current_temp"`
	CurrentHumidity *float64       `gorm:"type:decimal(5,2);comment:当前湿度(%)" json:"current_humidity"`
	SprayStatus     *int           `gorm:"comment:喷雾状态 0-关闭 1-开启" json:"spray_status"`
	ReportTime      time.Time      `gorm:"index;not null" json:"report_time"`
	CreatedAt       time.Time      `json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DeviceReport) TableName() string {
	return "device_reports"
}
