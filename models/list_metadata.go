package models

type Metadata struct {
	TotalAmount int `json:"totalAmount"`
}

type ListWithMetadata[T any] struct {
	Records  []T      `json:"records"`
	Metadata Metadata `json:"metadata"`
}

func NewListWithMetadata[T any](list []T, totalAmount int) ListWithMetadata[T] {
	return ListWithMetadata[T]{
		Records: list,
		Metadata: Metadata{
			TotalAmount: totalAmount,
		},
	}
}
