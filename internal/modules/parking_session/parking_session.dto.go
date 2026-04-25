package parking_session

import "time"

type ParkingSessionResponse struct {
	ID          uint       `json:"id"`
	LotID       uint       `json:"lot_id"`
	SlotID      *uint      `json:"slot_id,omitempty"`
	CardUID     string     `json:"card_uid"`
	CardType    string     `json:"card_type"`
	PlateNumber string     `json:"plate_number"`
	EntryTime   time.Time  `json:"entry_time"`
	ExitTime    *time.Time `json:"exit_time,omitempty"`
	Fee         *float64   `json:"fee,omitempty"`
	IsActive    bool       `json:"is_active"`
}

type CreateParkingSessionInput struct {
	LotID       uint   `json:"lot_id"`
	CardUID     string `json:"card_uid"`
	CardType    string `json:"card_type"`
	PlateNumber string `json:"plate_number"`
}

type AssignSlotInput struct {
	SessionID uint `json:"session_id"`
	SlotID    uint `json:"slot_id"`
}

type FinishParkingSessionInput struct {
	SessionID uint    `json:"session_id"`
	Fee       float64 `json:"fee"`
}
