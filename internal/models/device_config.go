package models

import (
	"time"

	"gorm.io/gorm"
)

type DeviceConfig struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	DeviceID          uint           `gorm:"uniqueIndex;not null" json:"device_id"`
	DeviceCode        string         `gorm:"size:50;index" json:"device_code"`
	TargetTemp        *float64       `gorm:"type:decimal(5,2);comment:目标温度(℃)" json:"target_temp"`
	TempTolerance     *float64       `gorm:"type:decimal(4,2);comment:温度容差(℃)" json:"temp_tolerance"`
	ValveMinOpen      *int           `gorm:"comment:阀门最小开度(%)" json:"valve_min_open"`
	ValveMaxOpen      *int           `gorm:"comment:阀门最大开度(%)" json:"valve_max_open"`
	FanMinSpeed       *int           `gorm:"comment:风机最小转速(rpm)" json:"fan_min_speed"`
	FanMaxSpeed       *int           `gorm:"comment:风机最大转速(rpm)" json:"fan_max_speed"`
	TargetHumidity    *float64       `gorm:"type:decimal(5,2);comment:目标湿度(%)" json:"target_humidity"`
	HumidityTolerance *float64       `gorm:"type:decimal(4,2);comment:湿度容差(%)" json:"humidity_tolerance"`
	SprayInterval     *int           `gorm:"comment:喷雾间隔(秒)" json:"spray_interval"`
	SprayDuration     *int           `gorm:"comment:喷雾持续时间(秒)" json:"spray_duration"`
	Version           int            `gorm:"default:1;comment:配置版本号" json:"version"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DeviceConfig) TableName() string {
	return "device_configs"
}
