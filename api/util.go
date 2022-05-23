package api

import (
	"regexp"
	"strconv"
)

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

func isUUIDV4(id string) bool {
	re := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return re.MatchString(id)
}

func toBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return parsed
}
