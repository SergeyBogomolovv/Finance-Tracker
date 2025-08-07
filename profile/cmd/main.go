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
	"os"
	"os/signal"
	"syscall"

	awsConf "github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/joho/godotenv"
)

func main() {
	conf := config.New()
	logger := log.New(conf.Env)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	awsCfg, err := awsConf.LoadDefaultConfig(ctx,
		awsConf.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(conf.S3AccessKey, conf.S3SecretKey, "")),
		awsConf.WithBaseEndpoint(conf.S3Endpoint),
		awsConf.WithRegion(conf.S3Region),
	)
	if err != nil {
		logger.Error("failed to load AWS config", "error", err)
		os.Exit(1)
	}

	postgres := postgres.MustNew(conf.PostgresURL)
	defer postgres.Close()
	logger.Info("postgres connected")

	txManager := transaction.NewManager(postgres)
	userRepo := repo.NewUserRepo(postgres)
	avatarRepo := repo.NewAvatarRepo(awsCfg)

	profileService := service.NewProfileService(userRepo, avatarRepo, txManager)
	profileController := controller.NewProfileController(profileService)

	app := app.New(logger, profileController)

	consumer := controller.NewEventsController(conf.KafkaBrokers, conf.KafkaGroupID, profileService)

	app.Start(conf.Host, conf.Port)
	go consumer.Consume(log.WithLogger(ctx, logger))
	<-ctx.Done()
	app.Stop()
	consumer.Close()
}

func init() {
	godotenv.Load()
}
