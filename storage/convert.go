package storage


func getBoundedLimit(limit uint) uint {
	if limit > maxLimit {
		return maxLimit
	} else {
		return limit
	}
}
