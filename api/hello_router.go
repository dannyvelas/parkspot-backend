package api

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func HelloRouter() func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", sayHello())
	}
}

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userId := ctx.Value("id")
		if userId == nil {
			err := errors.New("hello_router: key id not found in context")
			respondError(w, err, errInternalServerError)
			return
		}

		userIdString, ok := userId.(string)
		if !ok {
			err := errors.New("hello_router: key id is not string")
			respondError(w, err, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, "hello, "+userIdString)
	}
}
