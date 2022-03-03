package routing

import (
	"github.com/dannyvelas/parkspot-api/routing/internal"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

type AddAdminId struct {
	Id string
}

func (addAdminId AddAdminId) HelloRouter() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", sayHello(addAdminId.Id))
	}
}

func sayHello(adminId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Say Hello Endpoint")

		internal.RespondJson(w, http.StatusOK, "hello, "+adminId)
	}
}
