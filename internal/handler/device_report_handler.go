package handler

import (
	"incubator-backend/internal/service"
	"incubator-backend/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DeviceReportHandler struct {
	reportService service.DeviceReportService
}

func NewDeviceReportHandler(reportService service.DeviceReportService) *DeviceReportHandler {
	return &DeviceReportHandler{
		reportService: reportService,
	}
}

type DeviceReportRequest struct {
	ValveOpen       *float64 `json:"valve_open"`
	FanSpeed        *int     `json:"fan_speed"`
	CurrentTemp     *float64 `json:"current_temp"`
	CurrentHumidity *float64 `json:"current_humidity"`
	SprayStatus     *int     `json:"spray_status"`
	ReportTime      string   `json:"report_time"`
}

func (h *DeviceReportHandler) Report(c *gin.Context) {
	deviceCode := c.Param("device_code")
	if deviceCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "设备编号不能为空",
		})
		return
	}

	var req DeviceReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("report data bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	var reportTime time.Time
	if req.ReportTime != "" {
		parsed, err := time.Parse(time.RFC3339, req.ReportTime)
		if err == nil {
			reportTime = parsed
		}
	}

	reportData := &service.DeviceReportData{
		ValveOpen:       req.ValveOpen,
		FanSpeed:        req.FanSpeed,
		CurrentTemp:     req.CurrentTemp,
		CurrentHumidity: req.CurrentHumidity,
		SprayStatus:     req.SprayStatus,
		ReportTime:      reportTime,
	}

	checkResult, err := h.reportService.ReportDeviceData(deviceCode, reportData)
	if err != nil {
		logger.Warnf("report device data failed: %s, %v", deviceCode, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	response := gin.H{
		"code":    200,
		"message": "上报成功",
	}

	if checkResult != nil && checkResult.Triggered {
		response["warning"] = gin.H{
			"alarm_type":    checkResult.AlarmType,
			"alarm_level":   checkResult.AlarmLevel,
			"trigger_value": checkResult.TriggerValue,
			"threshold":     checkResult.Threshold,
			"description":   checkResult.Description,
		}
		response["message"] = checkResult.Description
	}

	c.JSON(http.StatusOK, response)
}

func (h *DeviceReportHandler) GetLatest(c *gin.Context) {
	deviceCode := c.Param("device_code")

	report, err := h.reportService.GetLatestReportByCode(deviceCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "未找到上报数据",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    report,
	})
}

func (h *DeviceReportHandler) ListByDevice(c *gin.Context) {
	idStr := c.Param("device_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var startTime, endTime time.Time
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")

	if startStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startStr)
	}
	if endStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endStr)
	}

	reports, total, err := h.reportService.ListReports(uint(id), startTime, endTime, page, pageSize)
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
			"list":      reports,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *DeviceReportHandler) GetCachedStatus(c *gin.Context) {
	deviceCode := c.Param("device_code")

	status, err := h.reportService.GetCachedStatus(deviceCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取缓存状态失败",
		})
		return
	}

	if status == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "暂无状态数据",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    status,
	})
}

func (h *DeviceReportHandler) SyncAllFromDB(c *gin.Context) {
	err := h.reportService.SyncAllStatusFromDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "同步失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "全量同步成功",
	})
}
