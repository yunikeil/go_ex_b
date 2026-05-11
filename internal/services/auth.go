package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"bankapi/internal/models"
	"bankapi/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	users  *repository.UserRepository
	email  *EmailService
	secret []byte
	ttl    time.Duration
	logger *logrus.Logger
}

func NewAuthService(users *repository.UserRepository, email *EmailService, secret string, ttl time.Duration, logger *logrus.Logger) *AuthService {
	return &AuthService{users: users, email: email, secret: []byte(secret), ttl: ttl, logger: logger}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (models.User, error) {
	if err := req.Validate(); err != nil {
		return models.User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	user, err := s.users.Create(ctx, req.Email, req.Username, string(hash))
	if err != nil {
		return models.User{}, err
	}
	if s.email != nil {
		if err := s.email.Send(user.Email, "Welcome to Bank API", "Your Bank API account has been created successfully."); err != nil {
			s.logger.WithError(err).WithField("user_id", user.ID).Warn("send registration email")
		}
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	if err := req.Validate(); err != nil {
		return models.LoginResponse{}, err
	}
	user, err := s.users.ByEmail(ctx, req.Email)
	if err != nil {
		return models.LoginResponse{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return models.LoginResponse{}, ErrInvalidCredentials
	}
	token, err := s.issueToken(user.ID)
	if err != nil {
		return models.LoginResponse{}, err
	}
	return models.LoginResponse{Token: token, User: user}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID int64) (models.User, error) {
	return s.users.ByID(ctx, userID)
}

func (s *AuthService) ParseToken(tokenText string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	return claims, nil
}

func (s *AuthService) issueToken(userID int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "bankapi",
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}
