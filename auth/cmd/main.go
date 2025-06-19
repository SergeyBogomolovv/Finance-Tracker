package main

import (
	"FinanceTracker/auth/internal/config"
	"FinanceTracker/auth/internal/controller"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "FinanceTracker/auth/pkg/api/auth"

	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	conf := config.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	oauthController := controller.NewOAuthController(conf.OAuthRedirectURL, conf.GoogleClientID, conf.GoogleClientSecret, conf.YandexClientID, conf.YandexClientSecret)

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, oauthController)

	logger.Info("server started", "addr", fmt.Sprintf("0.0.0.0:%d", conf.Port))
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conf.Port))
		if err != nil {
			logger.Error("failed to listen", "err", err)
			os.Exit(1)
		}
		if err := server.Serve(lis); err != nil {
			logger.Error("failed to listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	server.GracefulStop()
	logger.Info("server stopped")
}
