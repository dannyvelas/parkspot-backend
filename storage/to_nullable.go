package storage

import (
	"database/sql"
)

func toNullable(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	} else {
		return sql.NullString{String: s, Valid: true}
	}
}
