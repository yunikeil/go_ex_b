package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"bankapi/internal/config"
	"bankapi/internal/handlers"
	"bankapi/internal/middleware"
	"bankapi/internal/services"
)

type Dependencies struct {
	Config    config.Config
	Logger    *logrus.Logger
	Auth      *services.AuthService
	Accounts  *services.AccountService
	Cards     *services.CardService
	Credits   *services.CreditService
	Analytics *services.AnalyticsService
}

func New(deps Dependencies) http.Handler {
	r := mux.NewRouter()
	r.Use(middleware.CORS)
	r.Use(middleware.Metrics)
	r.Use(middleware.Logging(deps.Logger))
	r.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	authHandler := handlers.NewAuthHandler(deps.Auth)
	accountHandler := handlers.NewAccountHandler(deps.Accounts, deps.Analytics)
	cardHandler := handlers.NewCardHandler(deps.Cards)
	transferHandler := handlers.NewTransferHandler(deps.Accounts)
	creditHandler := handlers.NewCreditHandler(deps.Credits)
	analyticsHandler := handlers.NewAnalyticsHandler(deps.Analytics)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.JSON(w, http.StatusOK, map[string]string{
			"name":    "Bank API",
			"health":  "/health",
			"swagger": "http://localhost:8081",
		})
	}).Methods(http.MethodGet)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	r.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	api := r.PathPrefix("").Subrouter()
	api.Use(middleware.Auth(deps.Auth))
	api.HandleFunc("/me", authHandler.Me).Methods(http.MethodGet)
	api.HandleFunc("/accounts", accountHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/accounts", accountHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/accounts/{accountId:[0-9]+}/deposit", accountHandler.Deposit).Methods(http.MethodPost)
	api.HandleFunc("/accounts/{accountId:[0-9]+}/withdraw", accountHandler.Withdraw).Methods(http.MethodPost)
	api.HandleFunc("/accounts/{accountId:[0-9]+}/predict", accountHandler.Predict).Methods(http.MethodGet)
	api.HandleFunc("/cards", cardHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/cards", cardHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/transfer", transferHandler.Transfer).Methods(http.MethodPost)
	api.HandleFunc("/credits", creditHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/credits/{creditId:[0-9]+}/schedule", creditHandler.Schedule).Methods(http.MethodGet)
	api.HandleFunc("/analytics", analyticsHandler.Monthly).Methods(http.MethodGet)

	return r
}
