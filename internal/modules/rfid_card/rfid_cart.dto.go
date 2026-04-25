package rfid_card

type CreateRfidCardRequest struct {
	UID       string   `json:"uid" binding:"required"`
	CardType  CardType `json:"card_type" binding:"required"`
	OwnerName *string  `json:"owner_name"`
	IsActive  *bool    `json:"is_active"`
}

type UpdateRfidCardRequest struct {
	CardType  *CardType `json:"card_type"`
	OwnerName *string   `json:"owner_name"`
	IsActive  *bool     `json:"is_active"`
}

type RfidCardResponse struct {
	ID        uint     `json:"id"`
	UID       string   `json:"uid"`
	CardType  CardType `json:"card_type"`
	OwnerName *string  `json:"owner_name,omitempty"`
	IsActive  bool     `json:"is_active"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}
