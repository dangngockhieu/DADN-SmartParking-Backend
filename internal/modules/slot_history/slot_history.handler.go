package slot_history

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

// GetBySlotID godoc
// @Summary Lấy lịch sử theo vị trí đỗ
// @Description Trả về lịch sử hoạt động của một vị trí đỗ xe theo slot ID
// @Tags slot_history
// @Produce json
// @Security BearerAuth
// @Param slotId path int true "ID vị trí đỗ xe"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /slot-histories/{slotId} [get]
func (h *Handler) GetBySlotID(c *gin.Context) {
	var params GetSlotHistoryBySlotIDParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.Error(appErrors.NewBadRequest("slotId không hợp lệ"))
		return
	}

	history, err := h.service.FindBySlotID(params.SlotID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Lấy lịch sử thành công", history)
}
