package handlers

import (
	"net/http"

	"bankapi/internal/middleware"
	"bankapi/internal/models"
	"bankapi/internal/services"
)

type CreditHandler struct {
	credits *services.CreditService
}

func NewCreditHandler(credits *services.CreditService) *CreditHandler {
	return &CreditHandler{credits: credits}
}

func (h *CreditHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCreditRequest
	if err := Decode(r, &req); err != nil {
		Error(w, err)
		return
	}
	credit, err := h.credits.Create(r.Context(), middleware.UserID(r.Context()), req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, credit)
}

func (h *CreditHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	creditID, err := pathID(r, "creditId")
	if err != nil {
		Error(w, err)
		return
	}
	schedule, err := h.credits.Schedule(r.Context(), creditID, middleware.UserID(r.Context()))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, schedule)
}
