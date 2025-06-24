package main

import (
	"FinanceTracker/notification/internal/config"
	"FinanceTracker/notification/internal/consumer"
	"FinanceTracker/notification/internal/mailer"
	"FinanceTracker/notification/internal/repo"
	"FinanceTracker/notification/internal/service"
	"FinanceTracker/notification/pkg/events"
	"FinanceTracker/notification/pkg/postgres"

	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	conf := config.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	postgres := postgres.MustNew(conf.PostgresURL)
	defer postgres.Close()
	logger.Info("postgres connected")

	mailer := mailer.New(conf.SMTP)
	userRepo := repo.NewUserRepo(postgres)

	notificationService := service.NewNotificationService(logger, userRepo, mailer)

	consumer := consumer.MustNew(logger, conf.KafkaBrokers, notificationService)
	consumer.ConsumeTopic(ctx, events.AuthOTPGeneratedTopic)

	logger.Info("service started")
	<-ctx.Done()

	if err := consumer.Close(); err != nil {
		logger.Error("failed to close consumer", "err", err)
		os.Exit(1)
	}
	logger.Info("service stopped")
}

func init() {
	godotenv.Load()
}
