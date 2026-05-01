package dashboard

import (
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type HourlyCountRow struct {
	Hour  int   `gorm:"column:hour"`
	Count int64 `gorm:"column:count"`
}

func (r *Repository) CountTodayIn(start time.Time, end time.Time, lotID *uint) (int64, error) {
	var total int64

	db := r.db.Table("parking_sessions").
		Where("entry_time >= ? AND entry_time < ?", start, end)

	if lotID != nil {
		db = db.Where("lot_id = ?", *lotID)
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *Repository) CountTodayOut(start time.Time, end time.Time, lotID *uint) (int64, error) {
	var total int64

	db := r.db.Table("parking_sessions").
		Where("exit_time IS NOT NULL").
		Where("exit_time >= ? AND exit_time < ?", start, end)

	if lotID != nil {
		db = db.Where("lot_id = ?", *lotID)
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *Repository) CountCurrentVehicles(lotID *uint) (int64, error) {
	var total int64

	db := r.db.Table("parking_slots").
		Where("status = ?", "OCCUPIED")

	if lotID != nil {
		db = db.Where("lot_id = ?", *lotID)
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *Repository) CountCapacity(lotID *uint) (int64, error) {
	var total int64

	db := r.db.Table("parking_slots")

	if lotID != nil {
		db = db.Where("lot_id = ?", *lotID)
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *Repository) CountAvailableSlots(lotID *uint) (int64, error) {
	var total int64

	db := r.db.Table("parking_slots").
		Where("status = ?", "AVAILABLE")

	if lotID != nil {
		db = db.Where("lot_id = ?", *lotID)
	}

	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (r *Repository) GetHourlyIn(start time.Time, end time.Time, lotID *uint) ([]HourlyCountRow, error) {
	var rows []HourlyCountRow

	db := r.db.Table("parking_sessions").
		Select("HOUR(entry_time) AS hour, COUNT(*) AS count").
		Where("entry_time >= ? AND entry_time < ?", start, end)

	if lotID != nil {
		db = db.Where("lot_id = ?", *lotID)
	}

	if err := db.
		Group("HOUR(entry_time)").
		Order("hour ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *Repository) GetHourlyOut(start time.Time, end time.Time, lotID *uint) ([]HourlyCountRow, error) {
	var rows []HourlyCountRow

	db := r.db.Table("parking_sessions").
		Select("HOUR(exit_time) AS hour, COUNT(*) AS count").
		Where("exit_time IS NOT NULL").
		Where("exit_time >= ? AND exit_time < ?", start, end)

	if lotID != nil {
		db = db.Where("lot_id = ?", *lotID)
	}

	if err := db.
		Group("HOUR(exit_time)").
		Order("hour ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *Repository) GetLotName(lotID uint) (string, error) {
	var name string

	err := r.db.Table("parking_lots").
		Select("name").
		Where("id = ?", lotID).
		Scan(&name).Error

	if err != nil {
		return "", err
	}

	return name, nil
}
