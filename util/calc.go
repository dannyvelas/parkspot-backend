package util

import (
	"time"
)

func GetAmtDays(startDate, endDate time.Time) int {
	return int(endDate.Sub(startDate).Hours() / 24)
}
