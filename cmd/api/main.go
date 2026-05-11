package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"bankapi/internal/config"
	"bankapi/internal/db"
	"bankapi/internal/logging"
	"bankapi/internal/repository"
	apirouter "bankapi/internal/router"
	"bankapi/internal/services"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)

	pg, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("open database")
	}
	defer pg.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := pg.PingContext(ctx); err != nil {
		logger.WithError(err).Fatal("ping database")
	}
	if err := db.Migrate(ctx, pg, cfg.MigrationsDir); err != nil {
		logger.WithError(err).Fatal("run migrations")
	}
	if cfg.SeedEnabled {
		if err := db.Seed(ctx, pg, cfg.SeedDir); err != nil {
			logger.WithError(err).Fatal("run seed data")
		}
		logger.WithField("dir", cfg.SeedDir).Info("seed data applied")
	}

	var cache *redis.Client
	if cfg.RedisAddr != "" {
		cache = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
		if err := cache.Ping(context.Background()).Err(); err != nil {
			logger.WithError(err).Warn("redis unavailable, continuing without cache")
			_ = cache.Close()
			cache = nil
		}
	}

	repos := repository.New(pg)
	email := services.NewEmailService(cfg, logger)
	cbr := services.NewCBRService(cfg, cache, logger)
	auth := services.NewAuthService(repos.Users, email, cfg.JWTSecret, cfg.JWTTTL, logger)
	accounts := services.NewAccountService(repos.Accounts, repos.Transactions, logger)
	cards := services.NewCardService(repos.Cards, cfg.CardPGPKey, cfg.CardHMACKey, logger)
	credits := services.NewCreditService(repos.Credits, repos.Accounts, repos.Transactions, cbr, email, logger)
	analytics := services.NewAnalyticsService(repos.Analytics, repos.Credits)

	go startCreditScheduler(context.Background(), credits, cfg.CreditSchedulerEvery, logger)

	handler := apirouter.New(apirouter.Dependencies{
		Config:    cfg,
		Logger:    logger,
		Auth:      auth,
		Accounts:  accounts,
		Cards:     cards,
		Credits:   credits,
		Analytics: analytics,
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.WithField("port", cfg.Port).Info("api listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("http server")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("shutdown server")
	}
}

func startCreditScheduler(ctx context.Context, svc *services.CreditService, every time.Duration, logger *logrus.Logger) {
	if every <= 0 {
		every = 12 * time.Hour
	}
	run := func() {
		if err := svc.ProcessDuePayments(ctx); err != nil {
			logger.WithError(err).Error("process due credit payments")
		}
	}
	run()
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
