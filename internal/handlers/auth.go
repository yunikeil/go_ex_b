package handlers

import (
	"net/http"

	"bankapi/internal/middleware"
	"bankapi/internal/models"
	"bankapi/internal/services"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := Decode(r, &req); err != nil {
		Error(w, err)
		return
	}
	user, err := h.auth.Register(r.Context(), req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := Decode(r, &req); err != nil {
		Error(w, err)
		return
	}
	resp, err := h.auth.Login(r.Context(), req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.CurrentUser(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, user)
}
