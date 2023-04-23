package selectopts

import (
	"github.com/Masterminds/squirrel"
	"time"
)

type dateIntersect struct {
	startDate, endDate time.Time
}

func WithDateIntersect(startDate, endDate time.Time) dateIntersect {
	return dateIntersect{startDate, endDate}
}

func (dateIntersect dateIntersect) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	return selector.
		Where("start_ts <= ?", dateIntersect.endDate.Unix()).
		Where("end_ts >= ?", dateIntersect.startDate.Unix())
}
