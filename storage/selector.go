package storage

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
)

type Selector struct {
	selector    squirrel.SelectBuilder
	countSelect squirrel.SelectBuilder
	searchFn    func(string) squirrel.Sqlizer
}

func newSelector(selector, countSelect squirrel.SelectBuilder) Selector {
	return Selector{selector: selector, countSelect: countSelect}
}

func (selector Selector) withOpts(opts ...func(*Selector)) Selector {
	for _, opt := range opts {
		opt(&selector)
	}
	return selector
}

func WithPermitFilter(filter models.PermitFilter) func(*Selector) {
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

	whereSQL, ok := filterToSQL[filter]
	return func(opts *Selector) {
		if ok {
			opts.selector = opts.selector.Where(whereSQL)
			opts.countSelect = opts.countSelect.Where(whereSQL)
		}
	}
}

func withSearchFn(searchFn func(string) squirrel.Sqlizer) func(*Selector) {
	return func(opts *Selector) {
		opts.searchFn = searchFn
	}
}

func WithSearch(search string) func(*Selector) {
	return func(opts *Selector) {
		if search != "" && opts.searchFn != nil {
			opts.selector = opts.selector.Where(opts.searchFn(search))
			opts.countSelect = opts.countSelect.Where(opts.searchFn(search))
		}
	}
}

func WithLimitAndOffset(limit, offset int) func(*Selector) {
	return func(opts *Selector) {
		if limit >= 0 && offset >= 0 {
			opts.selector = opts.selector.
				Limit(uint64(getBoundedLimit(limit))).
				Offset(uint64(offset))
		}
	}
}

func WithReversed(reversed bool) func(*Selector) {
	return func(opts *Selector) {
		if !reversed {
			opts.selector = opts.selector.OrderBy("permit.id ASC")
		} else {
			opts.selector = opts.selector.OrderBy("permit.id DESC")
		}
	}
}
