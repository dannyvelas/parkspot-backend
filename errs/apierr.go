package errs

type ApiErr struct {
	StatusCode int
	message    string
}

func NewApiErr(statusCode int, message string) *ApiErr {
	return &ApiErr{
		StatusCode: statusCode,
		message:    message,
	}
}

// implements error interface
func (e *ApiErr) Error() string {
	return e.message
}
