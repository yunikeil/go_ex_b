package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	JWTSecret            string
	JWTTTL               time.Duration
	CardPGPKey           string
	CardHMACKey          string
	SMTPHost             string
	SMTPPort             int
	SMTPUser             string
	SMTPPassword         string
	SMTPFrom             string
	CBREndpoint          string
	CBRFallbackKeyRate   float64
	CreditSchedulerEvery time.Duration
	LogLevel             string
	MigrationsDir        string
	SeedEnabled          bool
	SeedDir              string
}

func Load() Config {
	return Config{
		Port:                 env("APP_PORT", "8080"),
		DatabaseURL:          env("DATABASE_URL", "postgres://bank:bank@localhost:5432/bank?sslmode=disable"),
		RedisAddr:            env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        env("REDIS_PASSWORD", ""),
		RedisDB:              envInt("REDIS_DB", 0),
		JWTSecret:            env("JWT_SECRET", "change-me-in-production"),
		JWTTTL:               envDuration("JWT_TTL", 24*time.Hour),
		CardPGPKey:           env("CARD_PGP_KEY", "dev-card-pgp-key"),
		CardHMACKey:          env("CARD_HMAC_KEY", "dev-card-hmac-key"),
		SMTPHost:             env("SMTP_HOST", "localhost"),
		SMTPPort:             envInt("SMTP_PORT", 1025),
		SMTPUser:             env("SMTP_USER", ""),
		SMTPPassword:         env("SMTP_PASSWORD", ""),
		SMTPFrom:             env("SMTP_FROM", "bank@example.test"),
		CBREndpoint:          env("CBR_ENDPOINT", "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"),
		CBRFallbackKeyRate:   envFloat("CBR_FALLBACK_KEY_RATE", 16.0),
		CreditSchedulerEvery: envDuration("CREDIT_SCHEDULER_EVERY", 12*time.Hour),
		LogLevel:             env("LOG_LEVEL", "info"),
		MigrationsDir:        env("MIGRATIONS_DIR", "migrations"),
		SeedEnabled:          envBool("SEED_ENABLED", false),
		SeedDir:              env("SEED_DIR", "seeds"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

func envFloat(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(env(key, ""), 64)
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := env(key, "")
	if value == "" {
		return fallback
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	if hours, err := strconv.Atoi(value); err == nil {
		return time.Duration(hours) * time.Hour
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := env(key, "")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
