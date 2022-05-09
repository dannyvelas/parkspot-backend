package api

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user := ctx.Value("user")
		if user == nil {
			log.Error().Msg("hello_router: key `user` not found in context")
			respondError(w, errInternalServerError)
			return
		}

		parsedUser, ok := user.(jwtUser)
		if !ok {
			log.Error().Msg("hello_router: context key `user` is not string")
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, "hello, "+parsedUser.Id)
	}
}
