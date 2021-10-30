package utils

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func newErrorResponse(msg string) ErrorResponse {
	return ErrorResponse{Error: msg}
}

type errorType struct {
	statusCode    int
	errorResponse ErrorResponse
}

var Unauthorized = errorType{http.StatusUnauthorized, newErrorResponse("Unauthorized")}
var BadRequest = errorType{http.StatusBadRequest, newErrorResponse("Bad Request")}
var InternalServerError = errorType{http.StatusInternalServerError, newErrorResponse("Internal Server Error")}

func HandleError(w http.ResponseWriter, errorType errorType) {
	RespondJson(w, errorType.statusCode, errorType.errorResponse)
}

func HandleInternalError(w http.ResponseWriter, message string) {
	log.Error().Msgf(message)
	HandleError(w, InternalServerError)
}
