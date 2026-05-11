package services

import (
	"context"
	"math"
	"time"

	"github.com/sirupsen/logrus"

	"bankapi/internal/models"
	"bankapi/internal/repository"
)

type CreditService struct {
	credits      *repository.CreditRepository
	accounts     *repository.AccountRepository
	transactions *repository.TransactionRepository
	cbr          *CBRService
	email        *EmailService
	logger       *logrus.Logger
}

func NewCreditService(credits *repository.CreditRepository, accounts *repository.AccountRepository, transactions *repository.TransactionRepository, cbr *CBRService, email *EmailService, logger *logrus.Logger) *CreditService {
	return &CreditService{credits: credits, accounts: accounts, transactions: transactions, cbr: cbr, email: email, logger: logger}
}

func (s *CreditService) Create(ctx context.Context, userID int64, req models.CreateCreditRequest) (models.Credit, error) {
	if err := models.ValidatePositiveAmount(req.Principal); err != nil {
		return models.Credit{}, err
	}
	if req.TermMonths <= 0 || req.TermMonths > 360 {
		return models.Credit{}, errText("term_months must be between 1 and 360")
	}
	keyRate := s.cbr.KeyRate(ctx)
	annualRate := keyRate + 5
	monthlyPayment := annuity(req.Principal, annualRate, req.TermMonths)
	firstDue := time.Now().AddDate(0, 1, 0)
	credit := models.Credit{
		UserID:         userID,
		AccountID:      req.AccountID,
		Principal:      req.Principal,
		AnnualRate:     annualRate,
		TermMonths:     req.TermMonths,
		MonthlyPayment: monthlyPayment,
		NextPaymentAt:  firstDue,
	}
	schedule := make([]models.PaymentSchedule, 0, req.TermMonths)
	for i := 0; i < req.TermMonths; i++ {
		schedule = append(schedule, models.PaymentSchedule{DueDate: firstDue.AddDate(0, i, 0), Amount: monthlyPayment})
	}
	created, err := s.credits.Create(ctx, credit, schedule)
	if err != nil {
		return models.Credit{}, err
	}
	if to, err := s.credits.UserEmail(ctx, userID); err == nil {
		if err := s.email.Send(to, "Credit issued", "Your credit has been issued and the payment schedule is available in Bank API."); err != nil {
			s.logger.WithError(err).Warn("send credit notification")
		}
	}
	return created, nil
}

func (s *CreditService) Schedule(ctx context.Context, creditID, userID int64) ([]models.PaymentSchedule, error) {
	return s.credits.Schedule(ctx, creditID, userID)
}

func (s *CreditService) ProcessDuePayments(ctx context.Context) error {
	due, err := s.credits.DuePayments(ctx, time.Now())
	if err != nil {
		return err
	}
	for _, item := range due {
		if err := s.credits.PaySchedule(ctx, item.ID); err != nil {
			if err != repository.ErrNotFound {
				s.logger.WithError(err).WithField("schedule_id", item.ID).Error("pay credit schedule")
			}
			continue
		}
		if to, err := s.credits.UserEmailBySchedule(ctx, item.ID); err == nil {
			if err := s.email.Send(to, "Credit payment processed", "A scheduled credit payment has been processed or updated."); err != nil {
				s.logger.WithError(err).Warn("send payment notification")
			}
		}
	}
	return nil
}

func annuity(principal, annualRate float64, months int) float64 {
	monthlyRate := annualRate / 100 / 12
	if monthlyRate == 0 {
		return principal / float64(months)
	}
	value := principal * (monthlyRate * math.Pow(1+monthlyRate, float64(months))) / (math.Pow(1+monthlyRate, float64(months)) - 1)
	return math.Round(value*100) / 100
}

type errText string

func (e errText) Error() string {
	return string(e)
}
