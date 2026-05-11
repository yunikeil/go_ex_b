package models

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

type AmountRequest struct {
	Amount float64 `json:"amount"`
}

type TransferRequest struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

type CreateCardRequest struct {
	AccountID int64 `json:"account_id"`
}

type CreateCreditRequest struct {
	AccountID  int64   `json:"account_id"`
	Principal  float64 `json:"principal"`
	TermMonths int     `json:"term_months"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
