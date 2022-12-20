package api

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("hello_router.sayHello: error getting access payload: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, "hello, "+AccessPayload.ID)
	}
}
