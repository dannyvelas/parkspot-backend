package api

import "strconv"

func toUint(value string) uint {
	if value == "" {
		return 0
	} else if parsed, err := strconv.ParseUint(value, 10, 32); err != nil {
		return 0
	} else {
		return uint(parsed)
	}
}

func getBoundedSizeAndOffset(size, page uint) (boundedSize, offset uint) {
	if size > maxPageSize {
		boundedSize = maxPageSize
	} else if size <= 0 {
		boundedSize = defaultPageSize
	} else {
		boundedSize = size
	}

	if page > 1 {
		offset = (page - 1) * boundedSize
	} else {
		offset = 0
	}

	return
}
