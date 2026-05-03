package rfid_card

import (
	"strconv"
	"time"

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
// @Summary Tạo thẻ RFID
// @Description Tạo mới thẻ RFID trong hệ thống
// @Tags rfid_card
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRfidCardRequest true "Thông tin thẻ RFID"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /rfid-cards [post]
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

// Update godoc
// @Summary Cập nhật thẻ RFID
// @Description Cập nhật thông tin thẻ RFID theo ID
// @Tags rfid_card
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID thẻ RFID"
// @Param request body UpdateRfidCardRequest true "Thông tin cập nhật"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /rfid-cards/{id} [patch]
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

// GetStatistics godoc
// @Summary Thống kê thẻ RFID
// @Description Lấy thống kê tổng số thẻ, đã đăng ký, chưa đăng ký, đang hoạt động
// @Tags rfid_card
// @Produce json
// @Security BearerAuth
// @Param date query string false "Ngày đăng ký (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /rfid-cards/statistics [get]
func (h *Handler) GetStatistics(c *gin.Context) {

	var registeredDate *time.Time
	if dateValue := c.Query("date"); dateValue != "" {
		parsed, err := time.ParseInLocation("2006-01-02", dateValue, time.Local)
		if err != nil {
			c.Error(appErrors.NewBadRequest("date không hợp lệ"))
			return
		}
		registeredDate = &parsed
	}

	result, err := h.service.GetStatistics(registeredDate)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Lấy thống kê thẻ RFID thành công", result)
}

// FindWithFilters godoc
// @Summary Lấy danh sách thẻ RFID
// @Description Lấy danh sách thẻ RFID có filter và phân trang
// @Tags rfid_card
// @Produce json
// @Security BearerAuth
// @Param lotId query int false "ID bãi xe, bỏ trống để lấy toàn bộ"
// @Param status query string false "REGISTERED hoặc GUEST"
// @Param keyword query string false "Từ khóa tìm kiếm"
// @Param page query int false "Trang hiện tại" default(1)
// @Param pageSize query int false "Số lượng mỗi trang" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /rfid-cards [get]
func (h *Handler) FindWithFilters(c *gin.Context) {
	var lotID *uint
	if lotIDValue := c.Query("lotId"); lotIDValue != "" {
		parsed, err := strconv.ParseUint(lotIDValue, 10, 64)
		if err != nil {
			c.Error(appErrors.NewBadRequest("lotId không hợp lệ"))
			return
		}
		value := uint(parsed)
		lotID = &value
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := c.Query("keyword")

	var status *CardType
	if statusValue := c.Query("status"); statusValue != "" {
		value := CardType(statusValue)
		status = &value
	}

	result, err := h.service.FindWithFilters(lotID, status, keyword, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, 200, "Lấy danh sách thẻ RFID thành công", result)
}
