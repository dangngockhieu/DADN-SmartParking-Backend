package parking_session

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(session *ParkingSession) error {
	return r.db.Create(session).Error
}

func (r *Repository) FindByID(id uint) (*ParkingSession, error) {
	var session ParkingSession
	err := r.db.First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) FindAll() ([]ParkingSession, error) {
	var sessions []ParkingSession
	err := r.db.Order("id DESC").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *Repository) FindActiveByCardUID(cardUID string) (*ParkingSession, error) {
	var session ParkingSession
	err := r.db.
		Where("card_uid = ? AND is_active = ?", cardUID, true).
		Order("id DESC").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) FindActiveByPlateNumber(plateNumber string) (*ParkingSession, error) {
	var session ParkingSession
	err := r.db.
		Where("plate_number = ? AND is_active = ?", plateNumber, true).
		Order("id DESC").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) UpdateByID(id uint, data map[string]any) error {
	return r.db.Model(&ParkingSession{}).Where("id = ?", id).Updates(data).Error
}

func (r *Repository) DeleteByID(id uint) error {
	return r.db.Unscoped().Delete(&ParkingSession{}, id).Error
}
