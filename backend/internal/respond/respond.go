package respond

import (
	"encoding/json"
	"log"
	"net/http"

	"recipe-backend/internal/models"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error: write json response: %v", err)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, models.ErrorResponse{Error: message})
}

func InternalError(w http.ResponseWriter, message string, err error) {
	if err != nil {
		log.Printf("error: %s: %v", message, err)
	}
	Error(w, http.StatusInternalServerError, message)
}
