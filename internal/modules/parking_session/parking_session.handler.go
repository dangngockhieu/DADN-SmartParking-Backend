package parking_session

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

// FindAll godoc
// @Summary Lấy danh sách phiên gửi xe
// @Description Trả về danh sách tất cả phiên gửi xe
// @Tags parking_session
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /parking-sessions [get]
func (h *Handler) FindAll(c *gin.Context) {
	sessions, err := h.service.FindAll()
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Lấy danh sách phiên gửi xe thành công", sessions)
}

// FindByID godoc
// @Summary Lấy chi tiết phiên gửi xe
// @Description Lấy thông tin phiên gửi xe theo ID
// @Tags parking_session
// @Produce json
// @Param id path int true "ID phiên gửi xe"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /parking-sessions/{id} [get]
func (h *Handler) FindByID(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewBadRequest("id không hợp lệ"))
		return
	}

	session, err := h.service.FindByID(uint(id64))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Session found", session)
}

// ForceClose closes a session by ID (for testing)
// @Summary Đóng phiên gửi xe theo ID (test)
// @Description Đóng phiên gửi xe theo ID, dùng cho testing
// @Tags parking_session
// @Produce json
// @Param id path int true "ID phiên gửi xe"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /parking-sessions/{id} [delete]
func (h *Handler) ForceClose(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewBadRequest("Invalid ID"))
		return
	}

	result, err := h.service.FinishSession(FinishParkingSessionInput{
		SessionID: uint(id64),
		Fee:       0,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Session closed", h.service.toResponse(result))
}

// CloseByCard closes active session by card UID (for testing)
// @Summary Đóng phiên gửi xe theo UID thẻ (test)
// @Description Đóng phiên gửi xe đang active theo UID thẻ, dùng cho testing
// @Tags parking_session
// @Produce json
// @Param uid path string true "UID thẻ"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /parking-sessions/card/{uid} [delete]
func (h *Handler) CloseByCard(c *gin.Context) {
	cardUID := c.Param("uid")
	if cardUID == "" {
		c.Error(appErrors.NewBadRequest("Card UID required"))
		return
	}

	session, err := h.service.FindActiveByCardUID(cardUID)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := h.service.FinishSession(FinishParkingSessionInput{
		SessionID: session.ID,
		Fee:       0,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Session closed", h.service.toResponse(result))
}

// HardDelete permanently removes a session from DB
// @Summary Xóa vĩnh viễn phiên gửi xe
// @Description Xóa vĩnh viễn phiên gửi xe khỏi DB
// @Tags parking_session
// @Produce json
// @Param id path int true "ID phiên gửi xe"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /parking-sessions/purge/{id} [delete]
func (h *Handler) HardDelete(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewBadRequest("Invalid ID"))
		return
	}

	if err := h.service.DeleteByID(uint(id64)); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Session deleted from DB", nil)
}
