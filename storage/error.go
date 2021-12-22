package storage

type NotFoundError struct {
	resource string
}

func newNotFoundError(resource string) NotFoundError {
	return NotFoundError{resource: resource}
}

func (e NotFoundError) Error() string {
	return e.resource + " not found"
}
