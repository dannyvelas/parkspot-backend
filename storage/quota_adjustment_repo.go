package storage

import (
	"github.com/dannyvelas/parkspot-backend/models"
)

type QuotaAdjustmentRepo interface {
	Create(adjustment models.QuotaAdjustment) (models.QuotaAdjustment, error)
	SelectByResident(residentID string) ([]models.QuotaAdjustment, error)
	SelectSumByResident(residentID string) (int, error)
	Reset() error // for testing
}
