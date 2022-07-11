package storage

import (
	"github.com/dannyvelas/lasvistas_api/config"
)

func getBoundedLimit(limit int) int {
	if limit > config.MaxLimit {
		return config.MaxLimit
	} else {
		return limit
	}
}
