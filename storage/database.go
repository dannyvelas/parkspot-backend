package storage

import (
	"github.com/jmoiron/sqlx"
)

type Database interface {
	Driver() *sqlx.DB
}
