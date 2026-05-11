package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"bankapi/internal/models"
	"bankapi/internal/repository"
	"bankapi/internal/services"
)

func JSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Decode(r *http.Request, out interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(out)
}

func Error(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, repository.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, repository.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, services.ErrInvalidCredentials):
		status = http.StatusUnauthorized
	}
	JSON(w, status, models.ErrorResponse{Error: err.Error()})
}
