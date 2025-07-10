package selectopts

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/parkspot-backend/config"
)

type limitAndOffset struct {
	limit, offset int
}

func WithLimitAndOffset(limit, offset int) limitAndOffset {
	return limitAndOffset{limit, offset}
}

func (limitAndOffset limitAndOffset) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	if limitAndOffset.limit >= 0 {
		selector = selector.Limit(uint64(getBoundedLimit(limitAndOffset.limit)))
	}
	if limitAndOffset.offset >= 0 {
		selector = selector.Offset(uint64(limitAndOffset.offset))
	}
	return selector
}

// helpers
func getBoundedLimit(limit int) int {
	if limit > config.MaxLimit {
		return config.MaxLimit
	} else {
		return limit
	}
}
