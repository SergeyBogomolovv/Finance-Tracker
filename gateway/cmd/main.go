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
	"FinanceTracker/gateway/pkg/logger"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title           FinanceTracker Gateway API
// @version         1.0
// @description     Документация HTTP API
func main() {
	conf := config.New()
	logger := logger.New(conf.Env)

	authConn, err := grpc.NewClient(conf.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	exitIfError(logger, err, "failed to create grpc auth client")

	authService := pb.NewAuthServiceClient(authConn)
	authController := controller.NewAuthController(authService, conf.OAuth)

	app := app.New(logger, conf, authController)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app.Start()
	<-ctx.Done()
	app.Stop()
}

func init() {
	godotenv.Load()
}

func exitIfError(logger *slog.Logger, err error, msg string) {
	if err != nil {
		logger.Error(msg, "err", err)
		os.Exit(1)
	}
}
