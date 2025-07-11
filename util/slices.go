package util

func MapSlice[T, U any](slice []T, fn func(T) U) []U {
	resultSlice := make([]U, len(slice))
	for i, e := range slice {
		resultSlice[i] = fn(e)
	}
	return resultSlice
}

func Contains[T comparable](slice []T, needle T) bool {
	for _, e := range slice {
		if needle == e {
			return true
		}
	}

	return false
}

func Find[T any](slice []T, fn func(T) bool) int {
	for i, e := range slice {
		if fn(e) {
			return i
		}
	}

	return -1
}
