package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Number    string    `json:"number"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type Card struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	AccountID int64     `json:"account_id"`
	Number    string    `json:"number,omitempty"`
	Expiry    string    `json:"expiry,omitempty"`
	Last4     string    `json:"last4"`
	HMAC      string    `json:"hmac,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	FromAccountID *int64    `json:"from_account_id,omitempty"`
	ToAccountID   *int64    `json:"to_account_id,omitempty"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type Credit struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	AccountID      int64     `json:"account_id"`
	Principal      float64   `json:"principal"`
	AnnualRate     float64   `json:"annual_rate"`
	TermMonths     int       `json:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment"`
	Status         string    `json:"status"`
	NextPaymentAt  time.Time `json:"next_payment_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type PaymentSchedule struct {
	ID             int64      `json:"id"`
	CreditID       int64      `json:"credit_id"`
	DueDate        time.Time  `json:"due_date"`
	Amount         float64    `json:"amount"`
	Paid           bool       `json:"paid"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	PenaltyApplied bool       `json:"penalty_applied"`
}

type MonthlyAnalytics struct {
	Income       float64 `json:"income"`
	Expense      float64 `json:"expense"`
	CreditLoad   float64 `json:"credit_load"`
	Transactions int64   `json:"transactions"`
}

type BalancePrediction struct {
	AccountID        int64   `json:"account_id"`
	Days             int     `json:"days"`
	CurrentBalance   float64 `json:"current_balance"`
	PlannedPayments  float64 `json:"planned_payments"`
	PredictedBalance float64 `json:"predicted_balance"`
}
