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

type RfidCardStatisticsResponse struct {
	TotalCards        int64 `json:"totalCards"`
	RegisteredCards   int64 `json:"registeredCards"`
	UnregisteredCards int64 `json:"unregisteredCards"`
	ActiveCards       int64 `json:"activeCards"`
}

type RfidCardListItem struct {
	ID           uint     `json:"id"`
	CardUID      string   `json:"cardUid"`
	PlateNumber  *string  `json:"plateNumber"`
	UserName     *string  `json:"userName"`
	Status       CardType `json:"status"`
	RegisteredAt *string  `json:"registeredAt"`
}

type RfidCardListMeta struct {
	TotalElements int64 `json:"totalElements"`
	TotalPages    int   `json:"totalPages"`
	CurrentPage   int   `json:"currentPage"`
	PageSize      int   `json:"pageSize"`
}

type RfidCardListResponse struct {
	Data []RfidCardListItem `json:"data"`
	Meta RfidCardListMeta   `json:"meta"`
}
