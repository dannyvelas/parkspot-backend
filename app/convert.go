package app

import (
	"github.com/dannyvelas/parkspot-backend/config"
)

func getBoundedLimitAndOffset(limit, page int) (boundedLimit, offset int) {
	if limit > config.MaxLimit {
		boundedLimit = config.MaxLimit
	} else if limit <= 0 {
		boundedLimit = config.DefaultLimit
	} else {
		boundedLimit = limit
	}

	if page <= 1 {
		offset = 0
	} else {
		offset = (page - 1) * boundedLimit
	}

	return
}
