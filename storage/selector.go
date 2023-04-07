package storage

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
)

type selectableRepo interface {
	searchSQL(string) squirrel.Sqlizer
}

type SelectOpt interface {
	dispatch(selectableRepo, squirrel.SelectBuilder) squirrel.SelectBuilder
}

type permitFilter struct {
	permitFilter models.PermitFilter
}

func WithPermitFilter(filter models.PermitFilter) permitFilter {
	return permitFilter{filter}
}

func (permitFilter permitFilter) dispatch(repo selectableRepo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
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

type searchFilter struct {
	search string
}

func WithSearch(search string) searchFilter {
	return searchFilter{search}
}
func (searchFilter searchFilter) dispatch(repo selectableRepo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	if searchFilter.search == "" {
		return selector
	}
	return selector.Where(repo.searchSQL(searchFilter.search))
}

type limitAndOffset struct {
	limit, offset int
}

func WithLimitAndOffset(limit, offset int) limitAndOffset {
	return limitAndOffset{limit, offset}
}
func (limitAndOffset limitAndOffset) dispatch(repo selectableRepo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	if limitAndOffset.limit >= 0 {
		selector = selector.Limit(uint64(getBoundedLimit(limitAndOffset.limit)))
	}
	if limitAndOffset.offset >= 0 {
		selector = selector.Offset(uint64(limitAndOffset.offset))
	}
	return selector
}

type reverseOp struct {
	reversed bool
}

func WithReversed(reversed bool) reverseOp {
	return reverseOp{reversed}
}
func (reverseOp reverseOp) dispatch(repo selectableRepo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	if !reverseOp.reversed {
		return selector.OrderBy("permit.id ASC")
	} else {
		return selector.OrderBy("permit.id DESC")
	}
}
