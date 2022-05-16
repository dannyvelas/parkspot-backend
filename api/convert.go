package api

import "strconv"

func toPosInt(value string) int {
	if len(value) > 0 && value[0] == '-' {
		value = value[1:]
	}

	if parsed, err := strconv.Atoi(value); err != nil {
		return 0
	} else {
		return parsed
	}
}

func getBoundedLimitAndOffset(limit, page int) (boundedLimit, offset int) {
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
