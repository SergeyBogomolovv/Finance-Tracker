package main

import (
	"FinanceTracker/auth/internal/app"
	"FinanceTracker/auth/internal/config"
	"FinanceTracker/auth/internal/controller"
	"FinanceTracker/auth/internal/producer"
	"FinanceTracker/auth/internal/repo"
	"FinanceTracker/auth/internal/service"
	"context"
	"os/signal"
	"syscall"

	log "FinanceTracker/auth/pkg/logger"
	"FinanceTracker/auth/pkg/postgres"
	"FinanceTracker/auth/pkg/transaction"

	"github.com/joho/godotenv"
)

func main() {
	conf := config.New()
	logger := log.New(conf.Env)

	postgres := postgres.MustNew(conf.PostgresURL)
	defer postgres.Close()
	logger.Info("postgres connected")

	txManager := transaction.NewManager(postgres)
	userRepo := repo.NewUserRepo(postgres)
	producer := producer.New(conf.KafkaBrokers)
	authService := service.NewAuthService(userRepo, producer, txManager, conf.JwtTTL, conf.JwtSecret)
	authController := controller.NewAuthController(authService, conf.OAuth)

	app := app.New(logger, authController)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app.Start(conf.Host, conf.Port)
	<-ctx.Done()
	app.Stop()
	producer.Close()
}

func init() {
	godotenv.Load()
}
