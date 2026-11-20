package app

import (
	"net/http"

	"github.com/dannyvelas/parkspot-backend/errs"
)

var ResidentTwoActivePermits = errs.NewAPIErr(
	http.StatusBadRequest,
	"Cannot create a permit during these dates"+
		" because this resident has at least two active permits during that time.")
