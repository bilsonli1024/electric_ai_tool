package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"electric_ai_tool/go_server/models"
)

func RespondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, err error, status int) {
	log.Printf("Error: %v", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
}
