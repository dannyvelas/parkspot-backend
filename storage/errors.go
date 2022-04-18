package storage

import (
	"fmt"
)

type sentinelError struct {
	message string
}

var (
	ErrConnecting      = sentinelError{"Error connecting to database"}
	ErrDatabaseQuery   = sentinelError{"Error querying database"}
	ErrQueryScanOneRow = sentinelError{"Error when querying and scanning one row"}
	ErrNoRows          = sentinelError{"No rows found"}
	ErrDatabaseExec    = sentinelError{"Error executing database command"}
	ErrGetRowsAffected = sentinelError{"Error getting rows affected"}
	ErrMissingField    = sentinelError{"Error: missing field"}
	ErrInvalidField    = sentinelError{"Error: invalid field"}
)

func (e sentinelError) Error() string {
	return e.message
}

func newError(sentinelErr sentinelError, err error) error {
	return fmt.Errorf("%w: %v", sentinelErr, err)
}
