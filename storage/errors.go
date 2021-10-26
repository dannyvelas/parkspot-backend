package storage

type NotFound struct {
	msg string
}

func NewNotFound(resource string) NotFound {
	return NotFound{msg: resource}
}

func (notFound NotFound) Error() string {
	return notFound.msg
}
