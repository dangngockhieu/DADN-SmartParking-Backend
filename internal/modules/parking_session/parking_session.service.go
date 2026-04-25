package parking_session

import (
	"errors"
	"strings"
	"time"

	appErrors "backend/internal/common/errors"
	"backend/internal/modules/rfid_card"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) toResponse(session *ParkingSession) *ParkingSessionResponse {
	return &ParkingSessionResponse{
		ID:          session.ID,
		LotID:       session.LotID,
		SlotID:      session.SlotID,
		CardUID:     session.CardUID,
		CardType:    string(session.CardType),
		PlateNumber: session.PlateNumber,
		EntryTime:   session.EntryTime,
		ExitTime:    session.ExitTime,
		Fee:         session.Fee,
		IsActive:    session.IsActive,
	}
}

func (s *Service) Create(input CreateParkingSessionInput) (*ParkingSession, error) {
	input.CardUID = strings.TrimSpace(input.CardUID)
	input.PlateNumber = strings.TrimSpace(input.PlateNumber)

	if input.CardUID == "" {
		return nil, appErrors.NewBadRequest("CardUID không được để trống")
	}
	if input.PlateNumber == "" {
		return nil, appErrors.NewBadRequest("PlateNumber không được để trống")
	}
	if input.CardType != string(rfid_card.CardTypeRegistered) && input.CardType != string(rfid_card.CardTypeGuest) {
		return nil, appErrors.NewBadRequest("CardType không hợp lệ")
	}

	session := &ParkingSession{
		LotID:       input.LotID,
		CardUID:     input.CardUID,
		CardType:    rfid_card.CardType(input.CardType),
		PlateNumber: input.PlateNumber,
		IsActive:    true,
	}

	if err := s.repo.Create(session); err != nil {
		return nil, appErrors.NewInternal("Tạo phiên gửi xe thất bại")
	}

	return session, nil
}

func (s *Service) AssignSlot(input AssignSlotInput) (*ParkingSession, error) {
	session, err := s.repo.FindByID(input.SessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewNotFound("Không tìm thấy phiên gửi xe")
		}
		return nil, appErrors.NewInternal("Lấy phiên gửi xe thất bại")
	}

	if !session.IsActive {
		return nil, appErrors.NewBadRequest("Phiên gửi xe đã đóng")
	}

	if err := s.repo.UpdateByID(session.ID, map[string]any{
		"slot_id": input.SlotID,
	}); err != nil {
		return nil, appErrors.NewInternal("Gán slot cho phiên gửi xe thất bại")
	}

	return s.repo.FindByID(session.ID)
}

func (s *Service) FinishSession(input FinishParkingSessionInput) (*ParkingSession, error) {
	session, err := s.repo.FindByID(input.SessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewNotFound("Không tìm thấy phiên gửi xe")
		}
		return nil, appErrors.NewInternal("Lấy phiên gửi xe thất bại")
	}

	if !session.IsActive {
		return nil, appErrors.NewBadRequest("Phiên gửi xe đã đóng")
	}

	now := time.Now()

	if err := s.repo.UpdateByID(session.ID, map[string]any{
		"exit_time": now,
		"fee":       input.Fee,
		"is_active": false,
	}); err != nil {
		return nil, appErrors.NewInternal("Kết thúc phiên gửi xe thất bại")
	}

	return s.repo.FindByID(session.ID)
}

func (s *Service) FindByID(id uint) (*ParkingSessionResponse, error) {
	session, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewNotFound("Không tìm thấy phiên gửi xe")
		}
		return nil, appErrors.NewInternal("Lấy thông tin phiên gửi xe thất bại")
	}

	return s.toResponse(session), nil
}

func (s *Service) FindAll() ([]ParkingSessionResponse, error) {
	sessions, err := s.repo.FindAll()
	if err != nil {
		return nil, appErrors.NewInternal("Lấy danh sách phiên gửi xe thất bại")
	}

	result := make([]ParkingSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, *s.toResponse(&session))
	}

	return result, nil
}

func (s *Service) FindActiveByCardUID(cardUID string) (*ParkingSession, error) {
	session, err := s.repo.FindActiveByCardUID(strings.TrimSpace(cardUID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewNotFound("Không có phiên gửi xe đang hoạt động")
		}
		return nil, appErrors.NewInternal("Lấy phiên gửi xe đang hoạt động thất bại")
	}
	return session, nil
}

func (s *Service) FindActiveByPlateNumber(plateNumber string) (*ParkingSession, error) {
	session, err := s.repo.FindActiveByPlateNumber(strings.TrimSpace(plateNumber))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NewNotFound("No active session with this plate")
		}
		return nil, appErrors.NewInternal("Failed to check plate session")
	}
	return session, nil
}

func (s *Service) DeleteByID(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.NewNotFound("Session not found")
		}
		return appErrors.NewInternal("Failed to find session")
	}
	return s.repo.DeleteByID(id)
}
