package gate

import (
	"strconv"

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

// Create godoc
// @Summary Tạo cổng mới
// @Tags gate
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateGateRequest true "Thông tin cổng"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /gates [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateGateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewBadRequest(err.Error()))
		return
	}

	gate, err := h.service.Create(&req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 201, "Tạo cổng thành công", ToGateResponse(gate))
}

// GetByID godoc
// @Summary Lấy thông tin cổng theo ID
// @Tags gate
// @Produce json
// @Security BearerAuth
// @Param gateId path int true "ID cổng"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /gates/{gateId} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "gateId")
	if err != nil {
		c.Error(appErrors.NewBadRequest("gateId không hợp lệ"))
		return
	}

	gate, err := h.service.FindByID(id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Lấy thông tin cổng thành công", ToGateResponse(gate))
}

// Update godoc
// @Summary Cập nhật thông tin cổng
// @Tags gate
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param gateId path int true "ID cổng"
// @Param body body UpdateGateRequest true "Thông tin cập nhật"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /gates/{gateId} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "gateId")
	if err != nil {
		c.Error(appErrors.NewBadRequest("gateId không hợp lệ"))
		return
	}

	var req UpdateGateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewBadRequest(err.Error()))
		return
	}

	gate, err := h.service.Update(id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Cập nhật cổng thành công", ToGateResponse(gate))
}

// Delete godoc
// @Summary Xoá cổng
// @Tags gate
// @Produce json
// @Security BearerAuth
// @Param gateId path int true "ID cổng"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /gates/{gateId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "gateId")
	if err != nil {
		c.Error(appErrors.NewBadRequest("gateId không hợp lệ"))
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Xoá cổng thành công", nil)
}

// ─── helper ──────────────────────────────────────────────────────────────────

func parseUintParam(c *gin.Context, key string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
