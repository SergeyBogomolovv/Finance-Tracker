package main

import (
	"FinanceTracker/auth/internal/config"
	"FinanceTracker/auth/internal/controller"
	"FinanceTracker/auth/internal/repo"
	"FinanceTracker/auth/internal/service"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "FinanceTracker/auth/pkg/api/auth"
	"FinanceTracker/auth/pkg/postgres"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	conf := config.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	postgres := postgres.MustNew(conf.PostgresURL)
	defer postgres.Close()
	logger.Info("postgres connected")

	userRepo := repo.NewUserRepo(postgres)

	authService := service.NewAuthService(logger, userRepo, conf.JwtTTL, conf.JwtSecret)

	authController := controller.NewAuthController(
		logger,
		authService,
		conf.OAuthRedirectURL,
		conf.GoogleClientID,
		conf.GoogleClientSecret,
		conf.YandexClientID,
		conf.YandexClientSecret,
	)

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, authController)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
		if err != nil {
			logger.Error("failed to listen", "err", err)
			os.Exit(1)
		}
		logger.Info("server started", "addr", lis.Addr())

		if err := server.Serve(lis); err != nil {
			logger.Error("failed to listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	server.GracefulStop()
	logger.Info("server stopped")
}

func init() {
	godotenv.Load()
}
