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
			respondError(w, fmt.Errorf("hello_router.sayHello: error getting access payload: %v", err))
			return
		}

		respondJSON(w, http.StatusOK, "hello, "+accessPayload.ID)
	}
}
