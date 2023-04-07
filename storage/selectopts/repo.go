package selectopts

import (
	"github.com/Masterminds/squirrel"
)

type Repo interface {
	SearchSQL(string) squirrel.Sqlizer
}
