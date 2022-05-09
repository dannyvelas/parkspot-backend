package api

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userId := ctx.Value("id")
		if userId == nil {
			log.Error().Msg("hello_router: key id not found in context")
			respondError(w, errInternalServerError)
			return
		}

		userIdString, ok := userId.(string)
		if !ok {
			log.Error().Msg("hello_router: context key `id` is not string")
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, "hello, "+userIdString)
	}
}
