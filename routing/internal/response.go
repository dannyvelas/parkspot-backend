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
		w.WriteHeader(http.StatusInternalServerError)

		log.Error().Msgf("Error encoding response: %s", err)

		if _, err := io.WriteString(w, `{"error": "Internal Server Error"}`); err != nil {
			log.Error().Msgf("Error sending Internal Server Error response: %s", err)
		}
	}
}
