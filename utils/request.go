package utils

import (
	"strconv"
)

func ToUint(value string) uint {
	if value == "" {
		return 0
	}
	res, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}

	return uint(res)
}

func PagingToLimitOffset(page, size uint) (limit, offset int) {
	if 0 < size && size < maxPageSize {
		limit = int(size)
	} else if size > maxPageSize {
		limit = maxPageSize
	} else {
		limit = defaultPageSize
	}

	if page > 1 {
		offset = (int(page) - 1) * limit
	} else {
		offset = 0
	}

	return
}
