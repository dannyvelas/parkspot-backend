package psql

import (
	"reflect"

	"github.com/Masterminds/squirrel"
)

var stmtBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func rmEmptyVals(whereClause squirrel.Eq) squirrel.Eq {
	newClause := make(squirrel.Eq)
	for key, value := range whereClause {
		if !reflect.ValueOf(value).IsZero() {
			newClause[key] = value
		}
	}

	return newClause
}
