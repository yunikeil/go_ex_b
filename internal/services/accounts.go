package services

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	"bankapi/internal/models"
	"bankapi/internal/repository"
)

type AccountService struct {
	accounts     *repository.AccountRepository
	transactions *repository.TransactionRepository
	logger       *logrus.Logger
}

func NewAccountService(accounts *repository.AccountRepository, transactions *repository.TransactionRepository, logger *logrus.Logger) *AccountService {
	return &AccountService{accounts: accounts, transactions: transactions, logger: logger}
}

func (s *AccountService) Create(ctx context.Context, userID int64, req models.CreateAccountRequest) (models.Account, error) {
	currency := req.Currency
	if currency == "" {
		currency = "RUB"
	}
	if currency != "RUB" {
		return models.Account{}, errors.New("only RUB currency is supported")
	}
	return s.accounts.Create(ctx, userID, currency)
}

func (s *AccountService) List(ctx context.Context, userID int64) ([]models.Account, error) {
	return s.accounts.ListByUser(ctx, userID)
}

func (s *AccountService) Get(ctx context.Context, userID, accountID int64) (models.Account, error) {
	return s.accounts.ByIDForUser(ctx, accountID, userID)
}

func (s *AccountService) Deposit(ctx context.Context, userID, accountID int64, amount float64) (models.Account, error) {
	if err := models.ValidatePositiveAmount(amount); err != nil {
		return models.Account{}, err
	}
	return s.accounts.ChangeBalance(ctx, accountID, userID, amount, "deposit", "balance deposit")
}

func (s *AccountService) Withdraw(ctx context.Context, userID, accountID int64, amount float64) (models.Account, error) {
	if err := models.ValidatePositiveAmount(amount); err != nil {
		return models.Account{}, err
	}
	return s.accounts.ChangeBalance(ctx, accountID, userID, -amount, "withdraw", "balance withdraw")
}

func (s *AccountService) Transfer(ctx context.Context, userID int64, req models.TransferRequest) error {
	if req.FromAccountID == req.ToAccountID {
		return errors.New("source and target accounts must be different")
	}
	if err := models.ValidatePositiveAmount(req.Amount); err != nil {
		return err
	}
	return s.accounts.Transfer(ctx, userID, req.FromAccountID, req.ToAccountID, req.Amount)
}
