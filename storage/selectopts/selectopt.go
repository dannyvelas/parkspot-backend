package selectopts

import (
	"github.com/Masterminds/squirrel"
)

type SelectOpt interface {
	Dispatch(Repo, squirrel.SelectBuilder) squirrel.SelectBuilder
}
