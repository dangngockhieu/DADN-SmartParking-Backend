package slot_history

import appErrors "backend/internal/common/errors"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) FindBySlotID(slotID uint) ([]SlotHistoryResponse, error) {
	history, err := s.repo.FindBySlotID(slotID)
	if err != nil {
		return nil, appErrors.NewInternal("Không thể lấy lịch sử slot")
	}
	return history, nil
}
