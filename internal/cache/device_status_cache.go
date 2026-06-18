package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"incubator-backend/internal/config"
	"incubator-backend/internal/models"
	"incubator-backend/pkg/logger"
	appredis "incubator-backend/pkg/redis"
	"time"

	"github.com/redis/go-redis/v9"
)

type DeviceStatusCache struct {
	ValveOpen       *float64 `json:"valve_open,omitempty"`
	FanSpeed        *int     `json:"fan_speed,omitempty"`
	CurrentTemp     *float64 `json:"current_temp,omitempty"`
	CurrentHumidity *float64 `json:"current_humidity,omitempty"`
	SprayStatus     *int     `json:"spray_status,omitempty"`
	ReportTimestamp int64    `json:"report_timestamp,omitempty"`
}

type DeviceStatusCacheManager interface {
	Set(deviceCode string, status *DeviceStatusCache) error
	Get(deviceCode string) (*DeviceStatusCache, error)
	Delete(deviceCode string) error
	UpdateFromReport(report *models.DeviceReport) error
	GetLatest(deviceCode string) (*DeviceStatusCache, error)
}

type deviceStatusCacheManager struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewDeviceStatusCacheManager() DeviceStatusCacheManager {
	cfg := config.GlobalConfig.Redis
	return &deviceStatusCacheManager{
		client: appredis.Client,
		prefix: config.GlobalConfig.Device.StatusCachePrefix,
		ttl:    cfg.CacheTTLDuration(),
	}
}

func (m *deviceStatusCacheManager) key(deviceCode string) string {
	return fmt.Sprintf("%s%s", m.prefix, deviceCode)
}

func (m *deviceStatusCacheManager) Set(deviceCode string, status *DeviceStatusCache) error {
	ctx := context.Background()

	data, err := json.Marshal(status)
	if err != nil {
		logger.Errorf("marshal device status cache failed: %v", err)
		return err
	}

	err = m.client.Set(ctx, m.key(deviceCode), data, m.ttl).Err()
	if err != nil {
		logger.Errorf("set device status cache failed: %v", err)
		return err
	}

	logger.Debugf("device status cache set: %s, size: %d bytes", deviceCode, len(data))
	return nil
}

func (m *deviceStatusCacheManager) Get(deviceCode string) (*DeviceStatusCache, error) {
	ctx := context.Background()

	data, err := m.client.Get(ctx, m.key(deviceCode)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		logger.Errorf("get device status cache failed: %v", err)
		return nil, err
	}

	var status DeviceStatusCache
	err = json.Unmarshal(data, &status)
	if err != nil {
		logger.Errorf("unmarshal device status cache failed: %v", err)
		return nil, err
	}

	return &status, nil
}

func (m *deviceStatusCacheManager) Delete(deviceCode string) error {
	ctx := context.Background()
	err := m.client.Del(ctx, m.key(deviceCode)).Err()
	if err != nil {
		logger.Errorf("delete device status cache failed: %v", err)
		return err
	}
	return nil
}

func (m *deviceStatusCacheManager) UpdateFromReport(report *models.DeviceReport) error {
	status := &DeviceStatusCache{
		ValveOpen:       report.ValveOpen,
		FanSpeed:        report.FanSpeed,
		CurrentTemp:     report.CurrentTemp,
		CurrentHumidity: report.CurrentHumidity,
		SprayStatus:     report.SprayStatus,
		ReportTimestamp: report.ReportTime.Unix(),
	}

	return m.Set(report.DeviceCode, status)
}

func (m *deviceStatusCacheManager) GetLatest(deviceCode string) (*DeviceStatusCache, error) {
	return m.Get(deviceCode)
}
