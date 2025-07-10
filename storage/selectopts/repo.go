package selectopts

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/parkspot-backend/models"
)

type Repo interface {
	SearchAsSQL(string) squirrel.Sqlizer
	StatusAsSQL(models.Status) (squirrel.Sqlizer, bool)
}
