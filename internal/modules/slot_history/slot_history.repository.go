package slot_history

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindBySlotID(slotID uint) ([]SlotHistoryResponse, error) {
	var history []SlotHistoryResponse

	err := r.db.
		Table("slot_histories sh").
		Select(`
			sh.id,
			sh.slot_id,
			sh.old_device,
			sh.new_device,
			sh.old_port,
			sh.new_port,
			sh.action,
			sh.created_at,
			u.id as user_id,
			u.email as user_email
		`).
		Joins("LEFT JOIN users u ON u.id = sh.user_id").
		Where("sh.slot_id = ?", slotID).
		Order("sh.created_at DESC").
		Scan(&history).Error

	if err != nil {
		return nil, err
	}

	return history, nil
}
