package internal

import (
	"encoding/json"
	"log"
	"net/http"
)

func checkMissingFields(payload map[string]string, fields ...string) []string {
	var missing []string
	for _, field := range fields {
		if _, ok := payload[field]; !ok {
			missing = append(missing, field)
		}
	}
	return missing
}

func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] Failed to encode body into resposne: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)

	type errorrResponse struct {
		Message string `json:"message"`
	}

	if err := json.NewEncoder(w).Encode(errorrResponse{message}); err != nil {
		log.Printf("[ERROR] Failed to encode error message into resposne: %v", err)
	}
}
