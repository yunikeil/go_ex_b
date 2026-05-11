package handlers

import (
	"net/http"

	"bankapi/internal/middleware"
	"bankapi/internal/models"
	"bankapi/internal/services"
)

type CardHandler struct {
	cards *services.CardService
}

func NewCardHandler(cards *services.CardService) *CardHandler {
	return &CardHandler{cards: cards}
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCardRequest
	if err := Decode(r, &req); err != nil {
		Error(w, err)
		return
	}
	card, err := h.cards.Create(r.Context(), middleware.UserID(r.Context()), req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, card)
}

func (h *CardHandler) List(w http.ResponseWriter, r *http.Request) {
	cards, err := h.cards.List(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, cards)
}
