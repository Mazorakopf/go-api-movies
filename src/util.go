package main

import (
	"encoding/json"
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

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func errorResponse(message string) map[string]string {
	return map[string]string{"message": message}
}
