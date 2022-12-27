package api

import (
	"github.com/dannyvelas/lasvistas_api/errs"
	"net/http"
)

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, *errs.Internalf("hello_router.sayHello: error getting access payload: %v", err))
			return
		}

		respondJSON(w, http.StatusOK, "hello, "+AccessPayload.ID)
	}
}
