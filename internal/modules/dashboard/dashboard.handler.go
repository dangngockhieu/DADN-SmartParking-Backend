package dashboard

import (
	appErrors "backend/internal/common/errors"
	"backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetParkingFlow godoc
// @Summary Thống kê lưu lượng xe dashboard
// @Description Lấy thống kê xe vào/ra theo ngày, số xe hiện tại, tỉ lệ lấp đầy và giờ cao điểm
// @Tags dashboard
// @Security BearerAuth
// @Param date query string false "Ngày thống kê, format yyyy-mm-dd"
// @Param lotId query int false "ID bãi xe, bỏ trống để lấy toàn bộ bãi"
// @Success 200 {object} ParkingFlowResponse
// @Router /dashboard/parking-flow [get]
func (h *Handler) GetParkingFlow(c *gin.Context) {
	var query ParkingFlowQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(appErrors.NewBadRequest("Query không hợp lệ"))
		return
	}

	result, err := h.service.GetParkingFlow(query)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, 200, "Lấy thông tin lưu lượng xe thành công", result)
}
