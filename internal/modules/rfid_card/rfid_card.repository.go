package rfid_card

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

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

type RfidCardListRow struct {
	ID          uint
	UID         string
	CardType    CardType
	OwnerName   *string
	IsActive    bool
	CreatedAt   time.Time
	PlateNumber *string
}

func (r *Repository) CountStatistics(lotID *uint) (int64, int64, int64, int64, error) {
	base := r.db.Model(&RfidCard{})
	if lotID != nil {
		base = base.Where(
			"EXISTS (SELECT 1 FROM parking_sessions ps WHERE ps.card_uid = rfid_cards.uid AND ps.lot_id = ?)",
			*lotID,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, 0, 0, 0, err
	}

	var registered int64
	if err := base.Session(&gorm.Session{}).
		Where("card_type = ?", CardTypeRegistered).
		Count(&registered).Error; err != nil {
		return 0, 0, 0, 0, err
	}

	var unregistered int64
	if err := base.Session(&gorm.Session{}).
		Where("card_type = ?", CardTypeGuest).
		Count(&unregistered).Error; err != nil {
		return 0, 0, 0, 0, err
	}

	var active int64
	if err := base.Session(&gorm.Session{}).
		Where("is_active = ?", true).
		Count(&active).Error; err != nil {
		return 0, 0, 0, 0, err
	}

	return total, registered, unregistered, active, nil
}

func (r *Repository) FindWithFilters(
	lotID *uint,
	status *CardType,
	keyword string,
	page int,
	pageSize int,
) ([]RfidCardListRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	latestBase := r.db.Table("parking_sessions AS ps1").
		Select("ps1.card_uid, ps1.plate_number, ps1.entry_time")
	maxEntry := r.db.Table("parking_sessions").
		Select("card_uid, MAX(entry_time) AS max_entry")
	if lotID != nil {
		latestBase = latestBase.Where("ps1.lot_id = ?", *lotID)
		maxEntry = maxEntry.Where("lot_id = ?", *lotID)
	}
	maxEntry = maxEntry.Group("card_uid")
	latest := latestBase.
		Joins("JOIN (?) AS ps2 ON ps1.card_uid = ps2.card_uid AND ps1.entry_time = ps2.max_entry", maxEntry)

	base := r.db.Table("rfid_cards AS rc").
		Joins("LEFT JOIN (?) AS ps ON ps.card_uid = rc.uid", latest)

	if lotID != nil {
		base = base.Where(
			"EXISTS (SELECT 1 FROM parking_sessions ps3 WHERE ps3.card_uid = rc.uid AND ps3.lot_id = ?)",
			*lotID,
		)
	}

	if status != nil {
		base = base.Where("rc.card_type = ?", *status)
	}

	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where(
			"rc.uid LIKE ? OR rc.owner_name LIKE ? OR ps.plate_number LIKE ?",
			like,
			like,
			like,
		)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).
		Distinct("rc.id").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]RfidCardListRow, 0)
	err := base.
		Select("rc.id, rc.uid, rc.card_type, rc.owner_name, rc.is_active, rc.created_at, ps.plate_number").
		Order("rc.id ASC").
		Offset(offset).
		Limit(pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
