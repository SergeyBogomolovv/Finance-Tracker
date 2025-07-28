package main

import (
	"FinanceTracker/notification/internal/config"
	"FinanceTracker/notification/internal/consumer"
	"FinanceTracker/notification/internal/mailer"
	"FinanceTracker/notification/internal/repo"
	"FinanceTracker/notification/internal/service"
	"FinanceTracker/notification/pkg/logger"
	"FinanceTracker/notification/pkg/postgres"

	"context"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	conf := config.New()
	log := logger.New(conf.Env)

	postgres := postgres.MustNew(conf.PostgresURL)
	defer postgres.Close()
	log.Info("postgres connected")

	mailer := mailer.New(conf.SMTP)
	userRepo := repo.NewUserRepo(postgres)
	notificationService := service.NewNotificationService(userRepo, mailer)
	consumer := consumer.New(conf.KafkaBrokers, conf.KafkaGroupID, notificationService)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	loggerCtx := logger.WithLogger(ctx, log)
	consumer.Start(loggerCtx)
	log.Info("consumer started")
	<-ctx.Done()
	consumer.Stop()
}

func init() {
	godotenv.Load()
}
