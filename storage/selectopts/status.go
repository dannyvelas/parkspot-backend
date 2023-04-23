package selectopts

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/lasvistas_api/models"
)

type status struct {
	status models.Status
}

func WithStatus(s models.Status) status {
	return status{s}
}

func (status status) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	statusAsSQL, ok := repo.StatusAsSQL(status.status)
	if !ok {
		return selector
	}
	return selector.Where(statusAsSQL)
}
