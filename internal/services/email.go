package services

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"

	"bankapi/internal/config"
)

type EmailService struct {
	cfg    config.Config
	logger *logrus.Logger
}

func NewEmailService(cfg config.Config, logger *logrus.Logger) *EmailService {
	return &EmailService{cfg: cfg, logger: logger}
}

func (s *EmailService) Send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.cfg.SMTPFrom)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	dialer := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUser, s.cfg.SMTPPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
