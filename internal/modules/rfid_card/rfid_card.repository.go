package rfid_card

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(card *RfidCard) error {
	return r.db.Create(card).Error
}

func (r *Repository) FindByID(id uint) (*RfidCard, error) {
	var card RfidCard
	err := r.db.First(&card, id).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *Repository) UpdateByID(id uint, data map[string]any) error {
	return r.db.Model(&RfidCard{}).Where("id = ?", id).Session(&gorm.Session{}).UpdateColumns(data).Error
}

func (r *Repository) FindByUID(uid string) (*RfidCard, error) {
	var card RfidCard
	err := r.db.Where("uid = ?", uid).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}
