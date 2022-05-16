package storage

func getBoundedLimit(limit int) int {
	if limit > maxLimit {
		return maxLimit
	} else {
		return limit
	}
}
