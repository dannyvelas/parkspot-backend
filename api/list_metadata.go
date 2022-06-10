package api

import (
	"github.com/dannyvelas/lasvistas_api/models"
)

type listType interface {
	models.Car | models.Permit | models.Resident | models.Visitor
}

type metadata struct {
	TotalAmount int `json:"totalAmount"`
}

type listWithMetadata[T listType] struct {
	Records  []T      `json:"records"`
	Metadata metadata `json:"metadata"`
}

func newListWithMetadata[T listType](list []T, totalAmount int) listWithMetadata[T] {
	return listWithMetadata[T]{
		Records: list,
		Metadata: metadata{
			TotalAmount: totalAmount,
		},
	}
}
