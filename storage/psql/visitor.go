package psql

import (
	"github.com/dannyvelas/lasvistas_api/models"
	"time"
)

type visitor struct {
	ID           string `db:"id"`
	ResidentID   string `db:"resident_id"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	Relationship string `db:"relationship"`
	AccessStart  int64  `db:"access_start"`
	AccessEnd    int64  `db:"access_end"`
}

func (visitor visitor) toModels() models.Visitor {
	return models.Visitor{
		ID:           visitor.ID,
		ResidentID:   visitor.ResidentID,
		FirstName:    visitor.FirstName,
		LastName:     visitor.LastName,
		Relationship: visitor.Relationship,
		AccessStart:  time.Unix(visitor.AccessStart, 0), // time.Unix() returns time in local tz
		AccessEnd:    time.Unix(visitor.AccessEnd, 0),
	}
}

type visitorSlice []visitor

func (visitors visitorSlice) toModels() []models.Visitor {
	modelsVisitors := make([]models.Visitor, 0, len(visitors))
	for _, visitor := range visitors {
		modelsVisitors = append(modelsVisitors, visitor.toModels())
	}
	return modelsVisitors
}
