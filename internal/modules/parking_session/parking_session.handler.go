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

func (h *Handler) FindAll(c *gin.Context) {
	sessions, err := h.service.FindAll()
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Lấy danh sách phiên gửi xe thành công", sessions)
}

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
