package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"bankapi/internal/models"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) Create(ctx context.Context, credit models.Credit, schedule []models.PaymentSchedule) (models.Credit, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Credit{}, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO credits (user_id, account_id, principal, annual_rate, term_months, monthly_payment, status, next_payment_at)
		SELECT $1, $2, $3, $4, $5, $6, 'active', $7
		WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $2 AND user_id = $1)
		RETURNING id, user_id, account_id, principal, annual_rate, term_months, monthly_payment, status, next_payment_at, created_at
	`, credit.UserID, credit.AccountID, credit.Principal, credit.AnnualRate, credit.TermMonths, credit.MonthlyPayment, credit.NextPaymentAt).
		Scan(&credit.ID, &credit.UserID, &credit.AccountID, &credit.Principal, &credit.AnnualRate, &credit.TermMonths, &credit.MonthlyPayment, &credit.Status, &credit.NextPaymentAt, &credit.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Credit{}, ErrNotFound
	}
	if err != nil {
		return models.Credit{}, err
	}

	for _, item := range schedule {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO payment_schedules (credit_id, due_date, amount)
			VALUES ($1, $2, $3)
		`, credit.ID, item.DueDate, item.Amount); err != nil {
			return models.Credit{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2 AND user_id = $3
	`, credit.Principal, credit.AccountID, credit.UserID); err != nil {
		return models.Credit{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, to_account_id, type, amount, description)
		VALUES ($1, $2, 'credit_disbursement', $3, 'credit issued')
	`, credit.UserID, credit.AccountID, credit.Principal); err != nil {
		return models.Credit{}, err
	}

	return credit, tx.Commit()
}

func (r *CreditRepository) UserEmail(ctx context.Context, userID int64) (string, error) {
	var email string
	err := r.db.QueryRowContext(ctx, `SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return email, err
}

func (r *CreditRepository) UserEmailBySchedule(ctx context.Context, scheduleID int64) (string, error) {
	var email string
	err := r.db.QueryRowContext(ctx, `
		SELECT u.email
		FROM payment_schedules ps
		JOIN credits c ON c.id = ps.credit_id
		JOIN users u ON u.id = c.user_id
		WHERE ps.id = $1
	`, scheduleID).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return email, err
}

func (r *CreditRepository) Schedule(ctx context.Context, creditID, userID int64) ([]models.PaymentSchedule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ps.id, ps.credit_id, ps.due_date, ps.amount, ps.paid, ps.paid_at, ps.penalty_applied
		FROM payment_schedules ps
		JOIN credits c ON c.id = ps.credit_id
		WHERE ps.credit_id = $1 AND c.user_id = $2
		ORDER BY ps.due_date
	`, creditID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.PaymentSchedule
	for rows.Next() {
		var item models.PaymentSchedule
		if err := rows.Scan(&item.ID, &item.CreditID, &item.DueDate, &item.Amount, &item.Paid, &item.PaidAt, &item.PenaltyApplied); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *CreditRepository) DuePayments(ctx context.Context, now time.Time) ([]models.PaymentSchedule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ps.id, ps.credit_id, ps.due_date, ps.amount, ps.paid, ps.paid_at, ps.penalty_applied
		FROM payment_schedules ps
		JOIN credits c ON c.id = ps.credit_id
		WHERE ps.paid = false AND ps.due_date <= $1 AND c.status = 'active'
		ORDER BY ps.due_date
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.PaymentSchedule
	for rows.Next() {
		var item models.PaymentSchedule
		if err := rows.Scan(&item.ID, &item.CreditID, &item.DueDate, &item.Amount, &item.Paid, &item.PaidAt, &item.PenaltyApplied); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *CreditRepository) PaySchedule(ctx context.Context, scheduleID int64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var creditID, userID, accountID int64
	var amount float64
	err = tx.QueryRowContext(ctx, `
		SELECT ps.credit_id, c.user_id, c.account_id, ps.amount
		FROM payment_schedules ps
		JOIN credits c ON c.id = ps.credit_id
		WHERE ps.id = $1 AND ps.paid = false
		FOR UPDATE
	`, scheduleID).Scan(&creditID, &userID, &accountID, &amount)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2 AND user_id = $3 AND balance >= $1
	`, amount, accountID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		amount = amount * 1.10
		if _, err := tx.ExecContext(ctx, `
			UPDATE payment_schedules
			SET amount = $1, penalty_applied = true
			WHERE id = $2 AND penalty_applied = false
		`, amount, scheduleID); err != nil {
			return err
		}
		return tx.Commit()
	}

	now := time.Now()
	if _, err := tx.ExecContext(ctx, `
		UPDATE payment_schedules
		SET paid = true, paid_at = $1
		WHERE id = $2
	`, now, scheduleID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, from_account_id, type, amount, description)
		VALUES ($1, $2, 'credit_payment', $3, 'scheduled credit payment')
	`, userID, accountID, amount); err != nil {
		return err
	}

	_ = creditID
	return tx.Commit()
}
