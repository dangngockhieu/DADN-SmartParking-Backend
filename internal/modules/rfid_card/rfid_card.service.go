package rfid_card

import (
	"errors"
	"strings"
	"time"

	appErrors "backend/internal/common/errors"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) validateCardType(cardType CardType) bool {
	return cardType == CardTypeRegistered || cardType == CardTypeGuest
}

func (s *Service) toResponse(card *RfidCard) *RfidCardResponse {
	return &RfidCardResponse{
		ID:        card.ID,
		UID:       card.UID,
		CardType:  card.CardType,
		OwnerName: card.OwnerName,
		IsActive:  card.IsActive,
		CreatedAt: card.CreatedAt.Format(time.RFC3339),
		UpdatedAt: card.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *Service) Create(req CreateRfidCardRequest) (*RfidCardResponse, error) {
	req.UID = strings.TrimSpace(req.UID)

	if req.UID == "" {
		return nil, appErrors.NewBadRequest("UID không được để trống")
	}
	if !s.validateCardType(req.CardType) {
		return nil, appErrors.NewBadRequest("CardType phải là REGISTERED hoặc GUEST")
	}

	var ownerName *string
	if req.OwnerName != nil {
		name := strings.TrimSpace(*req.OwnerName)
		ownerName = &name
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	card := &RfidCard{
		UID:       req.UID,
		CardType:  req.CardType,
		OwnerName: ownerName,
		IsActive:  isActive,
	}

	if err := s.repo.Create(card); err != nil {
		return nil, appErrors.NewInternal("Tạo thẻ RFID thất bại")
	}

	return s.toResponse(card), nil
}

func (s *Service) Update(id uint, req UpdateRfidCardRequest) (*RfidCardResponse, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewNotFound("Không tìm thấy thẻ RFID")
		}
		return nil, appErrors.NewInternal("Lấy thông tin thẻ RFID thất bại")
	}

	data := map[string]any{}

	if req.CardType != nil {
		if !s.validateCardType(*req.CardType) {
			return nil, appErrors.NewBadRequest("CardType phải là REGISTERED hoặc GUEST")
		}
		data["card_type"] = *req.CardType
	}

	if req.OwnerName != nil {
		name := strings.TrimSpace(*req.OwnerName)
		data["owner_name"] = name
	}

	if req.IsActive != nil {
		data["is_active"] = *req.IsActive
	}

	if len(data) == 0 {
		return nil, appErrors.NewBadRequest("Không có dữ liệu để cập nhật")
	}

	if err := s.repo.UpdateByID(id, data); err != nil {
		return nil, appErrors.NewInternal("Cập nhật thẻ RFID thất bại")
	}

	card, err := s.repo.FindByID(id)
	if err != nil {
		return nil, appErrors.NewInternal("Lấy thông tin thẻ RFID thất bại")
	}

	return s.toResponse(card), nil
}

func (s *Service) FindByUID(uid string) (*RfidCard, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return nil, appErrors.NewBadRequest("UID không được để trống")
	}

	card, err := s.repo.FindByUID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewNotFound("Không tìm thấy thẻ RFID")
		}
		return nil, appErrors.NewInternal("Lấy thông tin thẻ RFID thất bại")
	}

	return card, nil
}

func (s *Service) GetStatistics(registeredDate *time.Time) (*RfidCardStatisticsResponse, error) {
	total, registered, unregistered, active, registeredOnDate, err := s.repo.CountStatistics(registeredDate)
	if err != nil {
		return nil, appErrors.NewInternal("Thống kê thẻ RFID thất bại")
	}

	return &RfidCardStatisticsResponse{
		TotalCards:        total,
		RegisteredCards:   registered,
		UnregisteredCards: unregistered,
		ActiveCards:       active,
		RegisteredOnDate:  registeredOnDate,
	}, nil
}

func (s *Service) FindWithFilters(
	lotID *uint,
	status *CardType,
	keyword string,
	page int,
	pageSize int,
) (*RfidCardListResponse, error) {
	if status != nil && !s.validateCardType(*status) {
		return nil, appErrors.NewBadRequest("Status phải là REGISTERED hoặc GUEST")
	}

	rows, total, err := s.repo.FindWithFilters(lotID, status, keyword, page, pageSize)
	if err != nil {
		return nil, appErrors.NewInternal("Lấy danh sách thẻ RFID thất bại")
	}

	items := make([]RfidCardListItem, 0, len(rows))
	for _, row := range rows {
		var registeredAt *string
		if row.CardType == CardTypeRegistered {
			value := row.CreatedAt.Format("2006-01-02")
			registeredAt = &value
		}

		items = append(items, RfidCardListItem{
			ID:           row.ID,
			CardUID:      row.UID,
			PlateNumber:  row.PlateNumber,
			UserName:     row.OwnerName,
			Status:       row.CardType,
			RegisteredAt: registeredAt,
		})
	}

	meta := buildRfidCardListMeta(total, page, pageSize)

	return &RfidCardListResponse{
		Data: items,
		Meta: meta,
	}, nil
}

func buildRfidCardListMeta(totalElements int64, page int, pageSize int) RfidCardListMeta {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := 0
	if totalElements > 0 {
		totalPages = int((totalElements + int64(pageSize) - 1) / int64(pageSize))
	}

	return RfidCardListMeta{
		TotalElements: totalElements,
		TotalPages:    totalPages,
		CurrentPage:   page,
		PageSize:      pageSize,
	}
}
