package internal

import "strconv"

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

func GetBoundedSizeAndOffset(size, page uint) (boundedSize, offset uint) {
	if 0 < size && size < maxPageSize {
		boundedSize = size
	} else if size > maxPageSize {
		boundedSize = maxPageSize
	} else {
		boundedSize = defaultPageSize
	}

	if page > 1 {
		offset = (page - 1) * boundedSize
	} else {
		offset = 0
	}

	return
}
