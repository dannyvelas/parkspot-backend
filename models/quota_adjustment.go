package models

import (
	"time"
)

type QuotaAdjustment struct {
	ID               int       `json:"id"`
	ResidentID       string    `json:"residentID"`
	Amount           int       `json:"amount"`
	Reason           string    `json:"reason"`
	CreatedByAdminID *string   `json:"createdByAdminID"`
	CreatedAt        time.Time `json:"createdAt"`
}

func NewQuotaAdjustment(
	residentID string,
	amount int,
	reason string,
	createdByAdminID *string,
) QuotaAdjustment {
	return QuotaAdjustment{
		ResidentID:       residentID,
		Amount:           amount,
		Reason:           reason,
		CreatedByAdminID: createdByAdminID,
	}
}
