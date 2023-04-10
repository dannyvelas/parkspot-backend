package selectopts

import (
	"github.com/Masterminds/squirrel"
)

type reverseOp struct {
	reversed bool
}

func WithReversed(reversed bool) reverseOp {
	return reverseOp{reversed}
}

func (reverseOp reverseOp) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	if !reverseOp.reversed {
		return selector.OrderBy("id ASC")
	} else {
		return selector.OrderBy("id DESC")
	}
}
