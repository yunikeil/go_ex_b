package repository

import (
	"context"
	"database/sql"
	"errors"

	"bankapi/internal/models"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(ctx context.Context, userID, accountID int64, number, expiry, cvvHash, hmac, last4, pgpKey string) (models.Card, error) {
	var card models.Card
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO cards (user_id, account_id, number_encrypted, expiry_encrypted, cvv_hash, hmac, last4)
		SELECT $1, $2, pgp_sym_encrypt($3, $8), pgp_sym_encrypt($4, $8), $5, $6, $7
		WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $2 AND user_id = $1)
		RETURNING id, user_id, account_id, last4, hmac, created_at
	`, userID, accountID, number, expiry, cvvHash, hmac, last4, pgpKey).Scan(&card.ID, &card.UserID, &card.AccountID, &card.Last4, &card.HMAC, &card.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Card{}, ErrNotFound
	}
	return card, err
}

func (r *CardRepository) ListByUser(ctx context.Context, userID int64, pgpKey string) ([]models.Card, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, account_id,
		       pgp_sym_decrypt(number_encrypted, $2) AS number,
		       pgp_sym_decrypt(expiry_encrypted, $2) AS expiry,
		       last4, hmac, created_at
		FROM cards
		WHERE user_id = $1
		ORDER BY id
	`, userID, pgpKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(&card.ID, &card.UserID, &card.AccountID, &card.Number, &card.Expiry, &card.Last4, &card.HMAC, &card.CreatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}
