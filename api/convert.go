package api

import "strconv"

func toUint(value string) uint64 {
	if value == "" {
		return 0
	} else if parsed, err := strconv.ParseUint(value, 10, 32); err != nil {
		return 0
	} else {
		return uint64(parsed)
	}
}

func getBoundedLimitAndOffset(limit, page uint64) (boundedLimit, offset uint64) {
	if limit > maxPageLimit {
		boundedLimit = maxPageLimit
	} else if limit <= 0 {
		boundedLimit = defaultPageLimit
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
