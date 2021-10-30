package routing

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type errorType struct {
	statusCode  int
	errorString string
}

var Unauthorized = errorType{http.StatusUnauthorized, "Unauthorized"}
var BadRequest = errorType{http.StatusBadRequest, "Bad Request"}
var InternalServerError = errorType{http.StatusInternalServerError, "Internal Server Error"}

func HandleError(w http.ResponseWriter, errorType errorType) {
	RespondJson(w, errorType.statusCode, ErrorResponse{errorType.errorString})
}

func HandleInternalError(w http.ResponseWriter, message string) {
	log.Error().Msgf(message)
	HandleError(w, InternalServerError)
}
