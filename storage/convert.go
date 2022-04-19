package storage

import (
	"fmt"
	"time"
)

func getBoundedLimit(limit uint) uint {
	if limit > maxLimit {
		return maxLimit
	} else {
		return limit
	}
}

func parseStartEndDate(format, startDate, endDate string) (time.Time, time.Time, error) {
	startTime, err := time.ParseInLocation(format, startDate, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Error parsing startDate: %v", err)
	}

	endTime, err := time.ParseInLocation(format, endDate, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Error parsing endDate: %v", err)
	}

	return startTime, endTime, nil
}
