package errs

import "errors"

var (
	ErrJWTGenerating = errors.New("error generating JWT")
	ErrGmailClient   = errors.New("error constructing Gmail client")
	ErrSendingMail   = errors.New("error sending mail via Gmail API")
)
