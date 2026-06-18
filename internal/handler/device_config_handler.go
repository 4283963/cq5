package handler

import (
	"incubator-backend/internal/models"
	"incubator-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeviceConfigHandler struct {
	configService service.DeviceConfigService
}

func NewDeviceConfigHandler(configService service.DeviceConfigService) *DeviceConfigHandler {
	return &DeviceConfigHandler{
		configService: configService,
	}
}

type TemperatureValveConfig struct {
	TargetTemp    *float64 `json:"target_temp"`
	TempTolerance *float64 `json:"temp_tolerance"`
	ValveMinOpen  *int     `json:"valve_min_open"`
	ValveMaxOpen  *int     `json:"valve_max_open"`
}

type FanConfig struct {
	FanMinSpeed *int `json:"fan_min_speed"`
	FanMaxSpeed *int `json:"fan_max_speed"`
}

type HumiditySprayConfig struct {
	TargetHumidity    *float64 `json:"target_humidity"`
	HumidityTolerance *float64 `json:"humidity_tolerance"`
	SprayInterval     *int     `json:"spray_interval"`
	SprayDuration     *int     `json:"spray_duration"`
}

type CreateConfigRequest struct {
	DeviceID          uint     `json:"device_id" binding:"required"`
	TargetTemp        *float64 `json:"target_temp"`
	TempTolerance     *float64 `json:"temp_tolerance"`
	ValveMinOpen      *int     `json:"valve_min_open"`
	ValveMaxOpen      *int     `json:"valve_max_open"`
	FanMinSpeed       *int     `json:"fan_min_speed"`
	FanMaxSpeed       *int     `json:"fan_max_speed"`
	TargetHumidity    *float64 `json:"target_humidity"`
	HumidityTolerance *float64 `json:"humidity_tolerance"`
	SprayInterval     *int     `json:"spray_interval"`
	SprayDuration     *int     `json:"spray_duration"`
}

type UpdateConfigRequest struct {
	TargetTemp        *float64 `json:"target_temp"`
	TempTolerance     *float64 `json:"temp_tolerance"`
	ValveMinOpen      *int     `json:"valve_min_open"`
	ValveMaxOpen      *int     `json:"valve_max_open"`
	FanMinSpeed       *int     `json:"fan_min_speed"`
	FanMaxSpeed       *int     `json:"fan_max_speed"`
	TargetHumidity    *float64 `json:"target_humidity"`
	HumidityTolerance *float64 `json:"humidity_tolerance"`
	SprayInterval     *int     `json:"spray_interval"`
	SprayDuration     *int     `json:"spray_duration"`
}

func (h *DeviceConfigHandler) Create(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	config := &models.DeviceConfig{
		DeviceID:          req.DeviceID,
		TargetTemp:        req.TargetTemp,
		TempTolerance:     req.TempTolerance,
		ValveMinOpen:      req.ValveMinOpen,
		ValveMaxOpen:      req.ValveMaxOpen,
		FanMinSpeed:       req.FanMinSpeed,
		FanMaxSpeed:       req.FanMaxSpeed,
		TargetHumidity:    req.TargetHumidity,
		HumidityTolerance: req.HumidityTolerance,
		SprayInterval:     req.SprayInterval,
		SprayDuration:     req.SprayDuration,
	}

	err := h.configService.CreateConfig(config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "配置创建成功",
		"data":    config,
	})
}

func (h *DeviceConfigHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的配置ID",
		})
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	config := &models.DeviceConfig{
		ID:                uint(id),
		TargetTemp:        req.TargetTemp,
		TempTolerance:     req.TempTolerance,
		ValveMinOpen:      req.ValveMinOpen,
		ValveMaxOpen:      req.ValveMaxOpen,
		FanMinSpeed:       req.FanMinSpeed,
		FanMaxSpeed:       req.FanMaxSpeed,
		TargetHumidity:    req.TargetHumidity,
		HumidityTolerance: req.HumidityTolerance,
		SprayInterval:     req.SprayInterval,
		SprayDuration:     req.SprayDuration,
	}

	err = h.configService.UpdateConfig(config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "配置更新成功",
		"data":    config,
	})
}

func (h *DeviceConfigHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的配置ID",
		})
		return
	}

	err = h.configService.DeleteConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "配置删除成功",
	})
}

func (h *DeviceConfigHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的配置ID",
		})
		return
	}

	config, err := h.configService.GetConfigByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    config,
	})
}

func (h *DeviceConfigHandler) GetByDeviceID(c *gin.Context) {
	idStr := c.Param("device_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
		})
		return
	}

	config, err := h.configService.GetConfigByDeviceID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    config,
	})
}

func (h *DeviceConfigHandler) GetByDeviceCode(c *gin.Context) {
	deviceCode := c.Param("device_code")

	config, err := h.configService.GetConfigByDeviceCode(deviceCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    config,
	})
}

func (h *DeviceConfigHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	deviceType := models.DeviceType(c.Query("device_type"))

	configs, total, err := h.configService.ListConfigs(page, pageSize, deviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"list":      configs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *DeviceConfigHandler) GetCachedConfig(c *gin.Context) {
	deviceCode := c.Param("device_code")

	cached, err := h.configService.GetCachedConfig(deviceCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取缓存失败",
		})
		return
	}

	if cached == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "缓存不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    cached,
	})
}

func (h *DeviceConfigHandler) SyncToCache(c *gin.Context) {
	deviceCode := c.Param("device_code")

	err := h.configService.SyncConfigToCache(deviceCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "同步缓存失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "同步成功",
	})
}
