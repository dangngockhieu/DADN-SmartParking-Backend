package iot_device

import (
	"strings"
	"time"

	appErrors "backend/internal/common/errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDevice(req CreateIoTDeviceRequest) (*IoTDevice, error) {
	req.MacAddress = strings.TrimSpace(req.MacAddress)
	req.DeviceName = strings.TrimSpace(req.DeviceName)

	exist, err := s.repo.FindByMacAddress(req.MacAddress)
	if err != nil {
		return nil, appErrors.NewInternal("Có lỗi xảy ra khi kiểm tra thiết bị")
	}

	if exist != nil {
		return nil, appErrors.NewBadRequest("Device already exists")
	}

	now := time.Now()

	device := &IoTDevice{
		MacAddress: req.MacAddress,
		DeviceName: &req.DeviceName,
		LotID:      req.LotID,
		Status:     DeviceStatusActive,
		LastSeen:   &now,
	}

	if err := s.repo.CreateDevice(device); err != nil {
		return nil, appErrors.NewInternal("Tạo thiết bị IoT thất bại")
	}

	return device, nil
}
