package gate

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(gate *Gate) error {
	return r.db.Create(gate).Error
}

func (r *Repository) FindByID(id uint) (*Gate, error) {
	var gate Gate
	err := r.db.First(&gate, id).Error
	if err != nil {
		return nil, err
	}
	return &gate, nil
}

func (r *Repository) Update(gate *Gate) error {
	return r.db.Save(gate).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Gate{}, id).Error
}
