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

type DeviceConfigCache struct {
	ID                uint      `json:"id"`
	DeviceID          uint      `json:"device_id"`
	DeviceCode        string    `json:"device_code"`
	TargetTemp        *float64  `json:"target_temp"`
	TempTolerance     *float64  `json:"temp_tolerance"`
	ValveMinOpen      *int      `json:"valve_min_open"`
	ValveMaxOpen      *int      `json:"valve_max_open"`
	FanMinSpeed       *int      `json:"fan_min_speed"`
	FanMaxSpeed       *int      `json:"fan_max_speed"`
	TargetHumidity    *float64  `json:"target_humidity"`
	HumidityTolerance *float64  `json:"humidity_tolerance"`
	SprayInterval     *int      `json:"spray_interval"`
	SprayDuration     *int      `json:"spray_duration"`
	Version           int       `json:"version"`
	UpdateTime        time.Time `json:"update_time"`
}

type DeviceConfigCacheManager interface {
	Set(deviceCode string, config *DeviceConfigCache) error
	Get(deviceCode string) (*DeviceConfigCache, error)
	Delete(deviceCode string) error
	UpdateFromConfig(config *models.DeviceConfig) error
	InvalidateAll() error
}

type deviceConfigCacheManager struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewDeviceConfigCacheManager() DeviceConfigCacheManager {
	cfg := config.GlobalConfig.Redis
	return &deviceConfigCacheManager{
		client: appredis.Client,
		prefix: config.GlobalConfig.Device.ConfigCachePrefix,
		ttl:    cfg.CacheTTLDuration(),
	}
}

func (m *deviceConfigCacheManager) key(deviceCode string) string {
	return fmt.Sprintf("%s%s", m.prefix, deviceCode)
}

func (m *deviceConfigCacheManager) Set(deviceCode string, config *DeviceConfigCache) error {
	ctx := context.Background()
	config.UpdateTime = time.Now()

	data, err := json.Marshal(config)
	if err != nil {
		logger.Errorf("marshal device config cache failed: %v", err)
		return err
	}

	err = m.client.Set(ctx, m.key(deviceCode), data, m.ttl).Err()
	if err != nil {
		logger.Errorf("set device config cache failed: %v", err)
		return err
	}

	logger.Debugf("device config cache set: %s", deviceCode)
	return nil
}

func (m *deviceConfigCacheManager) Get(deviceCode string) (*DeviceConfigCache, error) {
	ctx := context.Background()

	data, err := m.client.Get(ctx, m.key(deviceCode)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		logger.Errorf("get device config cache failed: %v", err)
		return nil, err
	}

	var config DeviceConfigCache
	err = json.Unmarshal(data, &config)
	if err != nil {
		logger.Errorf("unmarshal device config cache failed: %v", err)
		return nil, err
	}

	return &config, nil
}

func (m *deviceConfigCacheManager) Delete(deviceCode string) error {
	ctx := context.Background()
	err := m.client.Del(ctx, m.key(deviceCode)).Err()
	if err != nil {
		logger.Errorf("delete device config cache failed: %v", err)
		return err
	}
	return nil
}

func (m *deviceConfigCacheManager) UpdateFromConfig(config *models.DeviceConfig) error {
	cacheConfig := &DeviceConfigCache{
		ID:                config.ID,
		DeviceID:          config.DeviceID,
		DeviceCode:        config.DeviceCode,
		TargetTemp:        config.TargetTemp,
		TempTolerance:     config.TempTolerance,
		ValveMinOpen:      config.ValveMinOpen,
		ValveMaxOpen:      config.ValveMaxOpen,
		FanMinSpeed:       config.FanMinSpeed,
		FanMaxSpeed:       config.FanMaxSpeed,
		TargetHumidity:    config.TargetHumidity,
		HumidityTolerance: config.HumidityTolerance,
		SprayInterval:     config.SprayInterval,
		SprayDuration:     config.SprayDuration,
		Version:           config.Version,
	}

	return m.Set(config.DeviceCode, cacheConfig)
}

func (m *deviceConfigCacheManager) InvalidateAll() error {
	ctx := context.Background()
	pattern := fmt.Sprintf("%s*", m.prefix)

	iter := m.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		err := m.client.Del(ctx, iter.Val()).Err()
		if err != nil {
			logger.Errorf("invalidate device config cache failed: %v", err)
			return err
		}
	}

	if err := iter.Err(); err != nil {
		logger.Errorf("scan device config cache failed: %v", err)
		return err
	}

	logger.Info("all device config cache invalidated")
	return nil
}
