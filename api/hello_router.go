package api

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := ctxGetUser(ctx)
		if err != nil {
			log.Error().Msgf("hello_router.sayHello: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, "hello, "+user.Id)
	}
}
