package gate

// ─── Request ─────────────────────────────────────────────────────────────────

type CreateGateRequest struct {
	Name       string   `json:"name"        binding:"required,max=50"`
	Type       GateType `json:"type"        binding:"required,oneof=ENTRY EXIT"`
	MacAddress string   `json:"mac_address" binding:"required,max=50"`
	LotID      uint     `json:"lot_id"      binding:"required"`
}

type UpdateGateRequest struct {
	Name       string   `json:"name"        binding:"omitempty,max=50"`
	Type       GateType `json:"type"        binding:"omitempty,oneof=ENTRY EXIT"`
	MacAddress string   `json:"mac_address" binding:"omitempty,max=50"`
	IsActive   *bool    `json:"is_active"`
}

// ─── URI params ──────────────────────────────────────────────────────────────

type GateURIParams struct {
	GateID uint `uri:"gateId" binding:"required"`
}

type LotURIParams struct {
	LotID uint `uri:"lotId" binding:"required"`
}

// ─── Response ─────────────────────────────────────────────────────────────────

type GateResponse struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	Type       GateType `json:"type"`
	MacAddress string   `json:"mac_address"`
	LotID      uint     `json:"lot_id"`
	IsActive   bool     `json:"is_active"`
}

func ToGateResponse(g *Gate) GateResponse {
	return GateResponse{
		ID:         g.ID,
		Name:       g.Name,
		Type:       g.Type,
		MacAddress: g.MacAddress,
		LotID:      g.LotID,
		IsActive:   g.IsActive,
	}
}
