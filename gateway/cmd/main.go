package main

import (
	"FinanceTracker/gateway/internal/app"
	"FinanceTracker/gateway/internal/config"
	"FinanceTracker/gateway/internal/controller"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	pb "FinanceTracker/gateway/pkg/api/auth"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conf := config.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	authConn, err := grpc.NewClient(conf.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to create grpc auth client", "err", err)
		os.Exit(1)
	}
	authService := pb.NewAuthServiceClient(authConn)
	authController := controller.NewAuthController(
		authService,
		conf.OAuthRedirectURL,
		conf.ClientRedirectURL,
		conf.GoogleClientID,
		conf.YandexClientID,
	)

	app := app.New(logger, conf)
	app.Init(authController)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app.Start()
	<-ctx.Done()
	app.Stop()
}

func init() {
	godotenv.Load()
}
