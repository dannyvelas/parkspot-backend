package storage

func getBoundedLimit(limit uint64) uint64 {
	if limit > maxLimit {
		return maxLimit
	} else {
		return limit
	}
}
