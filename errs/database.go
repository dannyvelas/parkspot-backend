package errs

import (
	"errors"
)

var (
	DBConnecting      = errors.New("Error connecting to database")
	DBQuery           = errors.New("Error querying database")
	DBQueryScanOneRow = errors.New("Error when querying and scanning one row in database")
	DBExec            = errors.New("Error executing database command")
	DBInvalidArg      = errors.New("Invalid argument to database")
	DBBuildingQuery   = errors.New("Error building database query")
	DBGetRowsAffected = errors.New("Error getting rows affected by database query")
)
