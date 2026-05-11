package repository

import (
	"context"
	"database/sql"
	"time"

	"bankapi/internal/models"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) Monthly(ctx context.Context, userID int64, from, to time.Time) (models.MonthlyAnalytics, error) {
	var analytics models.MonthlyAnalytics
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN type IN ('deposit', 'credit_disbursement') THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type IN ('withdraw', 'transfer', 'credit_payment') THEN amount ELSE 0 END), 0) AS expense,
			COUNT(*) AS transactions
		FROM transactions
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
	`, userID, from, to).Scan(&analytics.Income, &analytics.Expense, &analytics.Transactions)
	return analytics, err
}

func (r *AnalyticsRepository) PlannedCreditPayments(ctx context.Context, accountID, userID int64, until time.Time) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(ps.amount), 0)
		FROM payment_schedules ps
		JOIN credits c ON c.id = ps.credit_id
		WHERE c.account_id = $1 AND c.user_id = $2 AND ps.paid = false AND ps.due_date <= $3
	`, accountID, userID, until).Scan(&total)
	return total, err
}

func (r *AnalyticsRepository) CreditLoad(ctx context.Context, userID int64) (float64, error) {
	var load float64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(monthly_payment), 0)
		FROM credits
		WHERE user_id = $1 AND status = 'active'
	`, userID).Scan(&load)
	return load, err
}
