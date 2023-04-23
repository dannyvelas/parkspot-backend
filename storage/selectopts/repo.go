package selectopts

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
)

type Repo interface {
	SearchAsSQL(string) squirrel.Sqlizer
	StatusAsSQL(models.Status) (squirrel.Sqlizer, bool)
}
