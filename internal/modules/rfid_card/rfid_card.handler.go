package rfid_card

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

func (h *Handler) Create(c *gin.Context) {
	var req CreateRfidCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewBadRequest("Dữ liệu không hợp lệ"))
		return
	}

	card, err := h.service.Create(req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 201, "Tạo thẻ RFID thành công", card)
}

func (h *Handler) Update(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewBadRequest("id không hợp lệ"))
		return
	}

	var req UpdateRfidCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewBadRequest("Dữ liệu không hợp lệ"))
		return
	}

	card, err := h.service.Update(uint(id64), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Cập nhật thẻ RFID thành công", card)
}
