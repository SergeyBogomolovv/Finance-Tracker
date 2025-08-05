package main

import (
	"FinanceTracker/gateway/internal/app"
	"FinanceTracker/gateway/internal/config"
	"FinanceTracker/gateway/internal/controller"
	"FinanceTracker/gateway/internal/middleware"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	authPb "FinanceTracker/gateway/pkg/api/auth"
	profilePb "FinanceTracker/gateway/pkg/api/profile"

	"FinanceTracker/gateway/pkg/logger"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title           FinanceTracker Gateway API
// @version         1.0
// @description     Документация HTTP API
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Используйте формат "Bearer {token}"
func main() {
	conf := config.New()
	logger := logger.New(conf.Env)

	authMiddleware := middleware.NewAuth(conf.JwtSecret)

	authConn, err := grpc.NewClient(conf.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	exitIfError(logger, err, "failed to create grpc auth client")
	authService := authPb.NewAuthServiceClient(authConn)
	authController := controller.NewAuthController(authService, conf.OAuth)

	profileConn, err := grpc.NewClient(conf.ProfileServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	exitIfError(logger, err, "failed to create grpc profile client")
	profileService := profilePb.NewProfileServiceClient(profileConn)
	profileController := controller.NewProfileController(profileService, authMiddleware)

	app := app.New(logger, conf, authController, profileController)

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
