package parking_lot

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

type parkingLotStatsRow struct {
	Status string
	Count  int64
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(lot *ParkingLot) error {
	return r.db.Create(lot).Error
}

func (r *Repository) FindAll() ([]ParkingLot, error) {
	var lots []ParkingLot
	err := r.db.
		Select("id", "name", "location").
		Order("id ASC").
		Find(&lots).Error
	if err != nil {
		return nil, err
	}
	return lots, nil
}

func (r *Repository) FindByID(id uint) (*ParkingLot, error) {
	var lot ParkingLot
	err := r.db.
		Select("id", "name", "location").
		First(&lot, id).Error
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

func (r *Repository) UpdateByID(id uint, data map[string]interface{}) error {
	return r.db.Model(&ParkingLot{}).Where("id = ?", id).Updates(data).Error
}

func (r *Repository) FindSlotsByLotID(lotID uint) ([]ParkingLotSlotResponse, error) {
	var slots []ParkingLotSlotResponse

	err := r.db.
		Table("parking_slots").
		Select("id, name, status, device_mac, port_number").
		Where("lot_id = ?", lotID).
		Order("name ASC").
		Scan(&slots).Error

	if err != nil {
		return nil, err
	}

	return slots, nil
}

func (r *Repository) FindGatesByLotID(lotID uint) ([]ParkingLotGateResponse, error) {
	var gates []ParkingLotGateResponse

	err := r.db.
		Table("gates").
		Select("id, name, type, mac_address, is_active").
		Where("lot_id = ?", lotID).
		Order("id ASC").
		Scan(&gates).Error
	if err != nil {
		return nil, err
	}

	return gates, nil
}

func (r *Repository) CountStatsByLotID(lotID uint) ([]parkingLotStatsRow, error) {
	var rows []parkingLotStatsRow

	err := r.db.
		Table("parking_slots").
		Select("status, COUNT(*) as count").
		Where("lot_id = ?", lotID).
		Group("status").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	return rows, nil
}
