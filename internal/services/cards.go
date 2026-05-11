package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"bankapi/internal/models"
	"bankapi/internal/repository"
)

type CardService struct {
	cards   *repository.CardRepository
	pgpKey  string
	hmacKey []byte
	logger  *logrus.Logger
}

func NewCardService(cards *repository.CardRepository, pgpKey, hmacKey string, logger *logrus.Logger) *CardService {
	return &CardService{cards: cards, pgpKey: pgpKey, hmacKey: []byte(hmacKey), logger: logger}
}

func (s *CardService) Create(ctx context.Context, userID int64, req models.CreateCardRequest) (models.Card, error) {
	number, err := generateCardNumber()
	if err != nil {
		return models.Card{}, err
	}
	cvv, err := randomDigits(3)
	if err != nil {
		return models.Card{}, err
	}
	expiry := time.Now().AddDate(3, 0, 0).Format("01/06")
	cvvHash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return models.Card{}, err
	}
	signature := s.signature(number + "|" + expiry)
	card, err := s.cards.Create(ctx, userID, req.AccountID, number, expiry, string(cvvHash), signature, number[len(number)-4:], s.pgpKey)
	if err != nil {
		return models.Card{}, err
	}
	card.Number = maskCard(number)
	card.Expiry = expiry
	return card, nil
}

func (s *CardService) List(ctx context.Context, userID int64) ([]models.Card, error) {
	cards, err := s.cards.ListByUser(ctx, userID, s.pgpKey)
	if err != nil {
		return nil, err
	}
	for i := range cards {
		expected := s.signature(cards[i].Number + "|" + cards[i].Expiry)
		if !hmac.Equal([]byte(expected), []byte(cards[i].HMAC)) {
			cards[i].Number = "integrity check failed"
			continue
		}
		cards[i].Number = maskCard(cards[i].Number)
	}
	return cards, nil
}

func (s *CardService) signature(value string) string {
	mac := hmac.New(sha256.New, s.hmacKey)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func generateCardNumber() (string, error) {
	prefix := "2202"
	body, err := randomDigits(11)
	if err != nil {
		return "", err
	}
	partial := prefix + body
	return partial + luhnCheckDigit(partial), nil
}

func luhnCheckDigit(number string) string {
	sum := 0
	double := true
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return fmt.Sprintf("%d", (10-(sum%10))%10)
}

func randomDigits(n int) (string, error) {
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		out[i] = byte('0' + value.Int64())
	}
	return string(out), nil
}

func maskCard(number string) string {
	if len(number) < 8 {
		return number
	}
	return number[:4] + "********" + number[len(number)-4:]
}
