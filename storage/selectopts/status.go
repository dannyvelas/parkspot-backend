package selectopts

import (
	"github.com/Masterminds/squirrel"
	"github.com/dannyvelas/parkspot-backend/models"
)

type status struct {
	status models.Status
}

func WithStatus(s models.Status) status {
	return status{s}
}

func (status status) Dispatch(repo Repo, selector squirrel.SelectBuilder) squirrel.SelectBuilder {
	statusRepo, ok := repo.(StatusRepo)
	if !ok {
		return selector
	}

	statusAsSQL, ok := statusRepo.StatusAsSQL(status.status)
	if !ok {
		return selector
	}

	return selector.Where(statusAsSQL)
}
