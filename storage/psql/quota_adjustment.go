package psql

import (
	"database/sql"
	"time"

	"github.com/dannyvelas/parkspot-backend/models"
)

type quotaAdjustment struct {
	ID               int            `db:"id"`
	ResidentID       string         `db:"resident_id"`
	Amount           int            `db:"amount"`
	Reason           string         `db:"reason"`
	CreatedByAdminID sql.NullString `db:"created_by_admin_id"`
	CreatedAt        time.Time      `db:"created_at"`
}

func (quotaAdjustment quotaAdjustment) toModels() models.QuotaAdjustment {
	var createdByAdminID *string
	if quotaAdjustment.CreatedByAdminID.Valid {
		createdByAdminID = &quotaAdjustment.CreatedByAdminID.String
	}

	return models.QuotaAdjustment{
		ID:               quotaAdjustment.ID,
		ResidentID:       quotaAdjustment.ResidentID,
		Amount:           quotaAdjustment.Amount,
		Reason:           quotaAdjustment.Reason,
		CreatedByAdminID: createdByAdminID,
		CreatedAt:        quotaAdjustment.CreatedAt,
	}
}

type quotaAdjustmentSlice []quotaAdjustment

func (quotaAdjustments quotaAdjustmentSlice) toModels() []models.QuotaAdjustment {
	modelsQuotaAdjustments := make([]models.QuotaAdjustment, 0, len(quotaAdjustments))
	for _, quotaAdjustment := range quotaAdjustments {
		modelsQuotaAdjustments = append(modelsQuotaAdjustments, quotaAdjustment.toModels())
	}
	return modelsQuotaAdjustments
}
