package api

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type listType interface {
	models.Car | models.Permit | models.Resident
}

type metadata struct {
	TotalAmount int `json:"totalAmount"`
}

type listAndMetadata[T any] struct {
	Records  []T      `json:"records"`
	Metadata metadata `json:"metadata"`
}

func newListWithMetadata[T any](list []T, totalAmount int) listAndMetadata[T] {
	return listAndMetadata[T]{
		Records: list,
		Metadata: metadata{
			TotalAmount: totalAmount,
		},
	}
}
