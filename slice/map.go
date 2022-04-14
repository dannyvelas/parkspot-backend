package slice

// credit: https://github.com/akrennmair/slice
func Map[T, U any](input []T, f func(T) U) (output []U) {
	output = make([]U, 0, len(input))
	for _, v := range input {
		output = append(output, f(v))
	}
	return output
}
