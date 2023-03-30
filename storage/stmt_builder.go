package storage

import (
	"github.com/Masterminds/squirrel"
)

var (
	stmtBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)
