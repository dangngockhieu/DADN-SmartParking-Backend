package gate

import (
	appErrors "backend/internal/common/errors"

	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(req *CreateGateRequest) (*Gate, error) {
	gate := &Gate{
		Name:       req.Name,
		Type:       req.Type,
		MacAddress: req.MacAddress,
		LotID:      req.LotID,
	}
	if err := s.repo.Create(gate); err != nil {
		return nil, appErrors.NewInternal("Không thể tạo cổng")
	}
	return gate, nil
}

func (s *Service) FindByID(id uint) (*Gate, error) {
	gate, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, appErrors.NewNotFound("Không tìm thấy cổng")
		}
		return nil, appErrors.NewInternal("Không thể lấy thông tin cổng")
	}
	return gate, nil
}

func (s *Service) Update(id uint, req *UpdateGateRequest) (*Gate, error) {
	gate, err := s.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		gate.Name = req.Name
	}
	if req.Type != "" {
		gate.Type = req.Type
	}
	if req.MacAddress != "" {
		gate.MacAddress = req.MacAddress
	}
	if req.IsActive != nil {
		gate.IsActive = *req.IsActive
	}

	if err := s.repo.Update(gate); err != nil {
		return nil, appErrors.NewInternal("Không thể cập nhật cổng")
	}
	return gate, nil
}

func (s *Service) Delete(id uint) error {
	if _, err := s.FindByID(id); err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return appErrors.NewInternal("Không thể xoá cổng")
	}
	return nil
}
