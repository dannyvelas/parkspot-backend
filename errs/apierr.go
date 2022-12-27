package errs

type ApiErr struct {
	StatusCode int
	message    string
}

// implements error interface
func (e ApiErr) Error() string {
	return e.message
}
