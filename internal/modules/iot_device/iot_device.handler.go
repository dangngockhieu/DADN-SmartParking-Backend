package iot_device

import (
	"github.com/gin-gonic/gin"

	appErrors "backend/internal/common/errors"
	"backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateDevice godoc
// @Summary Tạo thiết bị IoT
// @Description Tạo mới một thiết bị IoT trong hệ thống
// @Tags iot_device
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateIoTDeviceRequest true "Thông tin thiết bị IoT"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /iot-devices [post]
func (h *Handler) CreateDevice(c *gin.Context) {
	var req CreateIoTDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewBadRequest("Dữ liệu không hợp lệ"))
		return
	}

	device, err := h.service.CreateDevice(req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 201, "Tạo thiết bị IoT thành công", device)
}
