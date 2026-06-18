package handler

import (
	"incubator-backend/internal/models"
	"incubator-backend/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AlarmHandler struct {
	alarmService  service.AlarmService
	safetyService service.SafetyService
}

func NewAlarmHandler(alarmService service.AlarmService, safetyService service.SafetyService) *AlarmHandler {
	return &AlarmHandler{
		alarmService:  alarmService,
		safetyService: safetyService,
	}
}

type ResolveAlarmRequest struct {
	Remark string `json:"remark"`
}

type RecoverDeviceRequest struct {
	Remark string `json:"remark"`
}

func (h *AlarmHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的报警ID",
		})
		return
	}

	alarm, err := h.alarmService.GetAlarmByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "报警记录不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alarm,
	})
}

func (h *AlarmHandler) GetActiveByDevice(c *gin.Context) {
	idStr := c.Param("device_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
		})
		return
	}

	alarm, err := h.alarmService.GetActiveAlarmByDevice(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "未找到活动报警",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    alarm,
	})
}

func (h *AlarmHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	deviceID, _ := strconv.ParseUint(c.DefaultQuery("device_id", "0"), 10, 32)
	alarmStatus, _ := strconv.Atoi(c.DefaultQuery("alarm_status", "0"))
	alarmType := models.AlarmType(c.Query("alarm_type"))

	var startTime, endTime time.Time
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")
	if startStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startStr)
	}
	if endStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endStr)
	}

	alarms, total, err := h.alarmService.ListAlarms(page, pageSize, uint(deviceID), models.AlarmStatus(alarmStatus), alarmType, startTime, endTime)
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
			"list":      alarms,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *AlarmHandler) Resolve(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的报警ID",
		})
		return
	}

	var req ResolveAlarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	err = h.alarmService.ResolveAlarm(uint(id), userID.(uint), req.Remark)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "报警已解除",
	})
}

func (h *AlarmHandler) RecoverDevice(c *gin.Context) {
	idStr := c.Param("device_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
		})
		return
	}

	var req RecoverDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	err = h.safetyService.RecoverDevice(uint(id), userID.(uint), req.Remark)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设备已恢复正常",
	})
}
