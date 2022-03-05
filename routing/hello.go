package routing

import (
	"github.com/dannyvelas/parkspot-api/routing/internal"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func HelloRouter() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", sayHello())
	}
}

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("Say Hello Endpoint")
		ctx := r.Context()

		userId := ctx.Value("id")
		if userId == nil {
			internal.HandleInternalError(w, "key id not found in context")
			return
		}

		userIdString, ok := userId.(string)
		if !ok {
			internal.HandleInternalError(w, "key id not string")
			return
		}

		internal.RespondJson(w, http.StatusOK, "hello, "+userIdString)
	}
}
