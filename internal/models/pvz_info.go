package models

type PVZInfo struct {
	PVZ        PVZ             `json:"pvz"`
	Receptions []ReceptionInfo `json:"receptions"`
}

type ReceptionInfo struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}
