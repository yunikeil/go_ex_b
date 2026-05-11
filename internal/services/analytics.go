package services

import (
	"context"
	"time"

	"bankapi/internal/models"
	"bankapi/internal/repository"
)

type AnalyticsService struct {
	analytics *repository.AnalyticsRepository
	credits   *repository.CreditRepository
}

func NewAnalyticsService(analytics *repository.AnalyticsRepository, credits *repository.CreditRepository) *AnalyticsService {
	return &AnalyticsService{analytics: analytics, credits: credits}
}

func (s *AnalyticsService) Monthly(ctx context.Context, userID int64) (models.MonthlyAnalytics, error) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, 0)
	result, err := s.analytics.Monthly(ctx, userID, from, to)
	if err != nil {
		return models.MonthlyAnalytics{}, err
	}
	load, err := s.analytics.CreditLoad(ctx, userID)
	if err != nil {
		return models.MonthlyAnalytics{}, err
	}
	result.CreditLoad = load
	return result, nil
}

func (s *AnalyticsService) Predict(ctx context.Context, account models.Account, days int) (models.BalancePrediction, error) {
	if days <= 0 || days > 365 {
		return models.BalancePrediction{}, errText("days must be between 1 and 365")
	}
	planned, err := s.analytics.PlannedCreditPayments(ctx, account.ID, account.UserID, time.Now().AddDate(0, 0, days))
	if err != nil {
		return models.BalancePrediction{}, err
	}
	return models.BalancePrediction{
		AccountID:        account.ID,
		Days:             days,
		CurrentBalance:   account.Balance,
		PlannedPayments:  planned,
		PredictedBalance: account.Balance - planned,
	}, nil
}
