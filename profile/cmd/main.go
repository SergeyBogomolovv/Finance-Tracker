package main

import (
	"FinanceTracker/profile/internal/app"
	"FinanceTracker/profile/internal/config"
	"FinanceTracker/profile/internal/controller"
	"FinanceTracker/profile/internal/repo"
	"FinanceTracker/profile/internal/service"
	log "FinanceTracker/profile/pkg/logger"
	"FinanceTracker/profile/pkg/postgres"
	"FinanceTracker/profile/pkg/transaction"
	"context"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	conf := config.New()
	logger := log.New(conf.Env)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	ctx = log.WithLogger(ctx, logger)

	postgres := postgres.MustNew(conf.PostgresURL)
	defer postgres.Close()
	logger.Info("postgres connected")

	txManager := transaction.NewManager(postgres)
	userRepo := repo.NewUserRepo(postgres)
	avatarRepo := repo.MustAvatarRepo(ctx, conf.S3)

	profileService := service.NewProfileService(userRepo, avatarRepo, txManager)
	profileController := controller.NewProfileController(profileService)

	app := app.New(logger, profileController)

	consumer := controller.NewEventsController(conf.KafkaBrokers, conf.KafkaGroupID, profileService)

	app.Start(conf.Host, conf.Port)
	go consumer.Consume(ctx)
	<-ctx.Done()
	consumer.Close()
	app.Stop()
}

func init() {
	godotenv.Load()
}
