package selectopts

import (
	"github.com/Masterminds/squirrel"
)

type search struct {
	search string
}

func WithSearch(s string) search {
	return search{s}
}

func (search search) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	if search.search == "" {
		return selector
	}

	searchRepo, ok := repo.(SearchRepo)
	if !ok {
		return selector
	}

	return selector.Where(searchRepo.SearchAsSQL(search.search))
}
