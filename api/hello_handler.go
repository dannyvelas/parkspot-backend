package api

import (
	"fmt"
	"net/http"
)

func sayHello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			respondError(w, r, fmt.Errorf("hello_router.sayHello: error getting access payload: %w", err))
			return
		}

		respondJSON(w, r, http.StatusOK, "hello, "+accessPayload.ID)
	}
}
