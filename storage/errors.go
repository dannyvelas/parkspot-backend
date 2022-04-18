package storage

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
	ErrEmptyIDArg      = sentinelError{"ID argument is the zero value for its type"}
)

func (e sentinelError) Error() string {
	return e.message
}
