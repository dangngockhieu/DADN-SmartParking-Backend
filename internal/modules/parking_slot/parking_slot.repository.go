package parking_slot

import (
	"backend/internal/modules/slot_history"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(id uint) (*ParkingSlot, error) {
	var slot ParkingSlot
	err := r.db.First(&slot, id).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *Repository) FindByDeviceMacAndPort(deviceMac string, port int) (*ParkingSlot, error) {
	var slot ParkingSlot
	err := r.db.
		Where("device_mac = ? AND port_number = ?", deviceMac, port).
		First(&slot).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *Repository) Create(slot *ParkingSlot) error {
	return r.db.Create(slot).Error
}

func (r *Repository) UpdateStatus(id uint, status SlotStatus) error {
	return r.db.Model(&ParkingSlot{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) FindConflictSlot(deviceMac string, portNumber int, excludeID uint) (*ParkingSlot, error) {
	var slot ParkingSlot
	err := r.db.
		Where("device_mac = ? AND port_number = ? AND id <> ?", deviceMac, portNumber, excludeID).
		First(&slot).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *Repository) ChangeDeviceAndWriteHistory(
	slotID uint,
	deviceMac string,
	portNumber int,
) (*ParkingSlot, error) {
	var updated ParkingSlot

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var slot ParkingSlot
		if err := tx.First(&slot, slotID).Error; err != nil {
			return err
		}

		if err := tx.Model(&ParkingSlot{}).
			Where("id = ?", slotID).
			Updates(map[string]any{
				"device_mac":  deviceMac,
				"port_number": portNumber,
			}).Error; err != nil {
			return err
		}

		if err := tx.Create(&slot_history.SlotHistory{
			SlotID:    slotID,
			OldDevice: &slot.DeviceMac,
			NewDevice: &deviceMac,
			OldPort:   &slot.PortNumber,
			NewPort:   &portNumber,
			Action:    slot_history.SlotHistoryActionDeviceChange,
		}).Error; err != nil {
			return err
		}

		if err := tx.First(&updated, slotID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &updated, nil
}
