package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"bankapi/internal/middleware"
	"bankapi/internal/models"
	"bankapi/internal/services"
)

type AccountHandler struct {
	accounts  *services.AccountService
	analytics *services.AnalyticsService
}

func NewAccountHandler(accounts *services.AccountService, analytics *services.AnalyticsService) *AccountHandler {
	return &AccountHandler{accounts: accounts, analytics: analytics}
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	_ = Decode(r, &req)
	account, err := h.accounts.Create(r.Context(), middleware.UserID(r.Context()), req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, account)
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.accounts.List(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, accounts)
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	h.changeBalance(w, r, true)
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	h.changeBalance(w, r, false)
}

func (h *AccountHandler) Predict(w http.ResponseWriter, r *http.Request) {
	accountID, err := pathID(r, "accountId")
	if err != nil {
		Error(w, err)
		return
	}
	days, err := strconv.Atoi(r.URL.Query().Get("days"))
	if err != nil {
		days = 30
	}
	userID := middleware.UserID(r.Context())
	account, err := h.accounts.Get(r.Context(), userID, accountID)
	if err != nil {
		Error(w, err)
		return
	}
	prediction, err := h.analytics.Predict(r.Context(), account, days)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, prediction)
}

func (h *AccountHandler) changeBalance(w http.ResponseWriter, r *http.Request, deposit bool) {
	accountID, err := pathID(r, "accountId")
	if err != nil {
		Error(w, err)
		return
	}
	var req models.AmountRequest
	if err := Decode(r, &req); err != nil {
		Error(w, err)
		return
	}
	userID := middleware.UserID(r.Context())
	var account models.Account
	if deposit {
		account, err = h.accounts.Deposit(r.Context(), userID, accountID, req.Amount)
	} else {
		account, err = h.accounts.Withdraw(r.Context(), userID, accountID, req.Amount)
	}
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, account)
}

func pathID(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)[key], 10, 64)
}
