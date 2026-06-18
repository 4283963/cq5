package handler

import (
	"incubator-backend/internal/models"
	"incubator-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	deviceService service.DeviceService
}

func NewDeviceHandler(deviceService service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

type CreateDeviceRequest struct {
	DeviceCode  string            `json:"device_code" binding:"required"`
	DeviceName  string            `json:"device_name" binding:"required"`
	DeviceType  models.DeviceType `json:"device_type" binding:"required"`
	RoomID      uint              `json:"room_id"`
	RoomName    string            `json:"room_name"`
	Status      int               `json:"status"`
	Description string            `json:"description"`
}

type UpdateDeviceRequest struct {
	DeviceName  string            `json:"device_name"`
	DeviceType  models.DeviceType `json:"device_type"`
	RoomID      uint              `json:"room_id"`
	RoomName    string            `json:"room_name"`
	Status      int               `json:"status"`
	Description string            `json:"description"`
}

func (h *DeviceHandler) Create(c *gin.Context) {
	var req CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	device := &models.Device{
		DeviceCode:  req.DeviceCode,
		DeviceName:  req.DeviceName,
		DeviceType:  req.DeviceType,
		RoomID:      req.RoomID,
		RoomName:    req.RoomName,
		Status:      req.Status,
		Description: req.Description,
	}
	if device.Status == 0 {
		device.Status = 1
	}

	err := h.deviceService.CreateDevice(device)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    device,
	})
}

func (h *DeviceHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
		})
		return
	}

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	device := &models.Device{
		ID:          uint(id),
		DeviceName:  req.DeviceName,
		DeviceType:  req.DeviceType,
		RoomID:      req.RoomID,
		RoomName:    req.RoomName,
		Status:      req.Status,
		Description: req.Description,
	}

	err = h.deviceService.UpdateDevice(device)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    device,
	})
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
		})
		return
	}

	err = h.deviceService.DeleteDevice(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

func (h *DeviceHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的设备ID",
		})
		return
	}

	device, err := h.deviceService.GetDeviceByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "设备不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    device,
	})
}

func (h *DeviceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	deviceType := models.DeviceType(c.Query("device_type"))
	roomID, _ := strconv.ParseUint(c.Query("room_id"), 10, 32)
	keyword := c.Query("keyword")

	devices, total, err := h.deviceService.ListDevices(page, pageSize, deviceType, uint(roomID), keyword)
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
			"list":      devices,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
