package models

import (
	"errors"
	"net/mail"
	"regexp"
)

var usernameRE = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func (r RegisterRequest) Validate() error {
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("invalid email")
	}
	if !usernameRE.MatchString(r.Username) {
		return errors.New("username must contain 3-32 letters, digits or underscores")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func (r LoginRequest) Validate() error {
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("invalid email")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func ValidatePositiveAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}
