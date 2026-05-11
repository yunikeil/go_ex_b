package repository

import "database/sql"

type Repository struct {
	DB           *sql.DB
	Users        *UserRepository
	Accounts     *AccountRepository
	Cards        *CardRepository
	Transactions *TransactionRepository
	Credits      *CreditRepository
	Analytics    *AnalyticsRepository
}

func New(db *sql.DB) *Repository {
	return &Repository{
		DB:           db,
		Users:        NewUserRepository(db),
		Accounts:     NewAccountRepository(db),
		Cards:        NewCardRepository(db),
		Transactions: NewTransactionRepository(db),
		Credits:      NewCreditRepository(db),
		Analytics:    NewAnalyticsRepository(db),
	}
}
