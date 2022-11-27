package storage

type UserRepo[Model any] interface {
	GetOne(id string) (Model, error)
	SetPassword(id string, hash string) error
}
