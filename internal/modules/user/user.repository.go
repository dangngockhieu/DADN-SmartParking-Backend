package user

import (
	"strings"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", strings.TrimSpace(strings.ToLower(email))).Take(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *Repository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (r *Repository) UpdateRole(id uint, role Role) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("role", role).Error
}

func (r *Repository) FindWithPagination(page, limit int, search string) ([]User, int64, error) {
	var users []User
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&User{}).Where("is_verified = ?", true)

	search = strings.TrimSpace(search)
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			r.db.Where("first_name LIKE ?", like).
				Or("last_name LIKE ?", like).
				Or("email LIKE ?", like),
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Select("id", "first_name", "last_name", "email", "role").
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
