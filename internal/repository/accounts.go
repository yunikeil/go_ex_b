package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"bankapi/internal/models"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, userID int64, currency string) (models.Account, error) {
	number := fmt.Sprintf("40817810%012d", time.Now().UnixNano()%1_000_000_000_000)
	var account models.Account
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO accounts (user_id, number, currency)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, number, currency, balance, created_at
	`, userID, number, currency).Scan(&account.ID, &account.UserID, &account.Number, &account.Currency, &account.Balance, &account.CreatedAt)
	return account, err
}

func (r *AccountRepository) ListByUser(ctx context.Context, userID int64) ([]models.Account, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, number, currency, balance, created_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		if err := rows.Scan(&account.ID, &account.UserID, &account.Number, &account.Currency, &account.Balance, &account.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

func (r *AccountRepository) ByIDForUser(ctx context.Context, id, userID int64) (models.Account, error) {
	var account models.Account
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, number, currency, balance, created_at
		FROM accounts
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&account.ID, &account.UserID, &account.Number, &account.Currency, &account.Balance, &account.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Account{}, ErrNotFound
	}
	return account, err
}

func (r *AccountRepository) ChangeBalance(ctx context.Context, accountID, userID int64, delta float64, txType, description string) (models.Account, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Account{}, err
	}
	defer tx.Rollback()

	var account models.Account
	err = tx.QueryRowContext(ctx, `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2 AND user_id = $3 AND balance + $1 >= 0
		RETURNING id, user_id, number, currency, balance, created_at
	`, delta, accountID, userID).Scan(&account.ID, &account.UserID, &account.Number, &account.Currency, &account.Balance, &account.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Account{}, ErrNotFound
	}
	if err != nil {
		return models.Account{}, err
	}

	var fromID *int64
	var toID *int64
	if delta < 0 {
		fromID = &accountID
	} else {
		toID = &accountID
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, from_account_id, to_account_id, type, amount, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, fromID, toID, txType, abs(delta), description); err != nil {
		return models.Account{}, err
	}

	return account, tx.Commit()
}

func (r *AccountRepository) Transfer(ctx context.Context, userID, fromID, toID int64, amount float64) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2 AND user_id = $3 AND balance >= $1
	`, amount, fromID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
	`, amount, toID)
	if err != nil {
		return err
	}
	affected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, from_account_id, to_account_id, type, amount, description)
		VALUES ($1, $2, $3, 'transfer', $4, 'account transfer')
	`, userID, fromID, toID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
