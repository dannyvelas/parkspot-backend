package internal

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func RespondJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if data == nil {
		_, _ = w.Write([]byte(""))
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Msg("Error parsing response")

		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error": "Internal Server Error"}`)
		return
	}
}
