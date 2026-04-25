package slot_history

import "time"

type GetSlotHistoryBySlotIDParams struct {
	SlotID uint `uri:"slotId" binding:"required"`
}

type SlotHistoryResponse struct {
	ID        uint              `json:"id"`
	SlotID    uint              `json:"slot_id"`
	OldDevice *string           `json:"old_device,omitempty"`
	NewDevice *string           `json:"new_device,omitempty"`
	OldPort   *int              `json:"old_port,omitempty"`
	NewPort   *int              `json:"new_port,omitempty"`
	Action    SlotHistoryAction `json:"action"`
	CreatedAt time.Time         `json:"created_at"`
	UserID    *uint             `json:"user_id,omitempty"`
	UserEmail *string           `json:"user_email,omitempty"`
}
