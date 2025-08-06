package main

import (
	"FinanceTracker/notification/internal/config"
	"FinanceTracker/notification/internal/consumer"
	"FinanceTracker/notification/internal/service"
	"FinanceTracker/notification/pkg/logger"

	"context"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	conf := config.New()
	log := logger.New(conf.Env)

	mailService := service.NewMailService(conf.SMTP)
	consumer := consumer.New(conf.KafkaBrokers, conf.KafkaGroupID, mailService)

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
