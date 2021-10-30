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

type errorType int8

const (
	Unauthorized errorType = iota
	BadRequest
	InternalServerError
)

func (errorType errorType) statusAndResponse() (int, ErrorResponse) {
	switch errorType {
	case Unauthorized:
		return http.StatusUnauthorized, newErrorResponse("Unauthorized")
	case BadRequest:
		return http.StatusBadRequest, newErrorResponse("Bad Request")
	default:
		return http.StatusInternalServerError, newErrorResponse("Internal Server Error.")
	}
}

func HandleError(w http.ResponseWriter, errorType errorType) {
	status, errorResponse := errorType.statusAndResponse()
	RespondJson(w, status, errorResponse)
}

func HandleInternalError(w http.ResponseWriter, message string) {
	log.Error().Msgf(message)
	HandleError(w, InternalServerError)
}
