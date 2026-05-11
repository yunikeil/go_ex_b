package handlers

import (
	"net/http"

	"bankapi/internal/middleware"
	"bankapi/internal/models"
	"bankapi/internal/services"
)

type TransferHandler struct {
	accounts *services.AccountService
}

func NewTransferHandler(accounts *services.AccountService) *TransferHandler {
	return &TransferHandler{accounts: accounts}
}

func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req models.TransferRequest
	if err := Decode(r, &req); err != nil {
		Error(w, err)
		return
	}
	if err := h.accounts.Transfer(r.Context(), middleware.UserID(r.Context()), req); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
