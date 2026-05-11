package handlers

import (
	"net/http"

	"bankapi/internal/middleware"
	"bankapi/internal/services"
)

type AnalyticsHandler struct {
	analytics *services.AnalyticsService
}

func NewAnalyticsHandler(analytics *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) Monthly(w http.ResponseWriter, r *http.Request) {
	result, err := h.analytics.Monthly(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, result)
}
