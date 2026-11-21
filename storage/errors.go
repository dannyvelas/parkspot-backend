package storage

import (
	"fmt"
	"net/http"

	"github.com/dannyvelas/parkspot-backend/errs"
)

var (
	NotFound = errs.NewAPIErr(http.StatusNotFound, "not found")

	ErrDBConnecting      = errs.NewSentinel("ErrDBConnecting", "error connecting to database")
	ErrDBQuery           = errs.NewSentinel("ErrDBQuery", "error querying database")
	ErrDBQueryScanOneRow = errs.NewSentinel("ErrDBQueryScanOneRow", "error when querying and scanning one row in database")
	ErrDBExec            = errs.NewSentinel("ErrDBExec", "error executing database command")
	ErrDBInvalidArg      = errs.NewSentinel("ErrDBInvalidArg", "invalid argument to database")
	ErrDBBuildingQuery   = errs.NewSentinel("ErrDBBuildingQuery", "error building database query")
	ErrDBGetRowsAffected = errs.NewSentinel("ErrDBGetRowsAffected", "error getting rows affected by database query")
	ErrDBPinging         = errs.NewSentinel("ErrDBPinging", "error pinging database")
)

func NewNotFound(resource string) *errs.APIErr {
	return errs.NewAPIErrFromError(http.StatusNotFound, fmt.Errorf("%s %w", resource, NotFound))
}
