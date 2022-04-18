package models

type sentinelError struct {
	message string
}

var (
	ErrEmptyFields   = sentinelError{"One or more missing fields"}
	ErrInvalidFields = sentinelError{"One or more invalid fields"}
)

func (e sentinelError) Error() string {
	return e.message
}
