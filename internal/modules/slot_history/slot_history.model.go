package slot_history

import "time"

type SlotHistoryAction string

const (
	SlotHistoryActionDeviceChange SlotHistoryAction = "DEVICE_CHANGE"
	SlotHistoryActionStatusChange SlotHistoryAction = "STATUS_CHANGE"
	SlotHistoryActionSystemFix    SlotHistoryAction = "SYSTEM_FIX"
	SlotHistoryActionMaintainMode SlotHistoryAction = "MAINTAIN_MODE"
)

type SlotHistory struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	SlotID    uint    `gorm:"not null;index;index:idx_slot_created_at"`
	OldDevice *string `gorm:"type:varchar(50)"`
	NewDevice *string `gorm:"type:varchar(50)"`
	OldPort   *int
	NewPort   *int
	Action    SlotHistoryAction `gorm:"type:enum('DEVICE_CHANGE','STATUS_CHANGE','SYSTEM_FIX','MAINTAIN_MODE');default:'DEVICE_CHANGE';not null"`
	UserID    *uint             `gorm:"index"`
	CreatedAt time.Time         `gorm:"autoCreateTime;index;index:idx_slot_created_at"`
}

func (SlotHistory) TableName() string {
	return "slot_histories"
}
