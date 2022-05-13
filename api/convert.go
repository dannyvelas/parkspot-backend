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

func getBoundedSizeAndOffset(size, page uint64) (boundedSize, offset uint64) {
	if size > maxPageSize {
		boundedSize = maxPageSize
	} else if size <= 0 {
		boundedSize = defaultPageSize
	} else {
		boundedSize = size
	}

	if page <= 1 {
		offset = 0
	} else {
		offset = (page - 1) * boundedSize
	}

	return
}
