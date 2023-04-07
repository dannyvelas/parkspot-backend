package selectopts

import (
	"github.com/Masterminds/squirrel"
)

type searchFilter struct {
	search string
}

func WithSearch(search string) searchFilter {
	return searchFilter{search}
}

func (searchFilter searchFilter) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	if searchFilter.search == "" {
		return selector
	}
	return selector.Where(repo.SearchSQL(searchFilter.search))
}
