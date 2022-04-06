package storage

import (
	"fmt"
)

type sentinelError struct {
	message string
}

var (
	ErrDatabaseQuery   = sentinelError{"Error querying database"}
	ErrScanningRow     = sentinelError{"Error Scanning Row"}
	ErrIterating       = sentinelError{"Error when iterating sql rows"}
	ErrQueryScanOneRow = sentinelError{"Error when querying and scanning one row"}
	ErrNoRows          = sentinelError{"No rows found"}
	ErrDatabaseExec    = sentinelError{"Error executing database command"}
	ErrGetRowsAffected = sentinelError{"Error getting rows affected"}
)

func (e sentinelError) Error() string {
	return e.message
}

func newError(sentinelErr sentinelError, err error) error {
	return fmt.Errorf("%v: %v", sentinelErr, err)
}
