package selectopts

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
)

type permitFilter struct {
	permitFilter models.PermitFilter
}

func WithPermitFilter(filter models.PermitFilter) permitFilter {
	return permitFilter{filter}
}

func (permitFilter permitFilter) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	filterToSQL := map[models.PermitFilter]squirrel.Sqlizer{
		models.ActivePermits: squirrel.And{
			squirrel.Expr("permit.start_ts <= extract(epoch from now())"),
			squirrel.Expr("permit.end_ts >= extract(epoch from now())"),
		},
		models.ExceptionPermits: squirrel.Expr("permit.exception_reason IS NOT NULL"),
		models.ExpiredPermits: squirrel.And{
			squirrel.Expr("permit.end_ts >= extract(epoch from (CURRENT_DATE - '1 DAY'::interval * ?))", config.DefaultExpiredWindow),
			squirrel.Expr("permit.end_ts <= extract(epoch from (CURRENT_DATE-2))"),
		},
	}

	whereSQL, ok := filterToSQL[permitFilter.permitFilter]
	if ok {
		return selector.Where(whereSQL)
	}
	return selector
}
