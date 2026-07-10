package errs

import (
	"errors"
)

var (
	ErrDBConnecting      = errors.New("error connecting to database")
	ErrDBQuery           = errors.New("error querying database")
	ErrDBQueryScanOneRow = errors.New("error when querying and scanning one row in database")
	ErrDBExec            = errors.New("error executing database command")
	ErrDBInvalidArg      = errors.New("invalid argument to database")
	ErrDBBuildingQuery   = errors.New("error building database query")
	ErrDBGetRowsAffected = errors.New("error getting rows affected by database query")
	ErrDBPinging         = errors.New("error pinging database")
)
