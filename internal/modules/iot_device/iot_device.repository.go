package iot_device

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByMacAddress(macAddress string) (*IoTDevice, error) {
	var device IoTDevice
	err := r.db.Where("mac_address = ?", macAddress).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *Repository) CreateDevice(device *IoTDevice) error {
	return r.db.Create(device).Error
}
