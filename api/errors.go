package api

import (
	"net/http"

	"github.com/dannyvelas/parkspot-backend/errs"
)

func BadRequest(message string) *errs.APIErr {
	return errs.NewAPIErr(http.StatusBadRequest, message)
}

func Malformed(payload string) *errs.APIErr {
	return errs.NewAPIErr(http.StatusBadRequest, payload+" malformed")
}
