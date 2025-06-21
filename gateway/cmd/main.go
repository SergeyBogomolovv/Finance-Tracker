package main

import (
	"FinanceTracker/gateway/internal/config"
	"FinanceTracker/gateway/internal/controller"
	"FinanceTracker/gateway/internal/middleware"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "FinanceTracker/gateway/pkg/api/auth"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	conf := config.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	authConn, err := grpc.NewClient(conf.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to create grpc auth client: %v", err)
	}
	authService := pb.NewAuthServiceClient(authConn)
	authController := controller.NewAuthController(
		authService,
		conf.OAuthRedirectURL,
		conf.ClientRedirectURL,
		conf.GoogleClientID,
		conf.YandexClientID,
	)

	cors := middleware.NewCORS(middleware.CORSConfig{
		AllowedOrigins:   conf.CorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	mux := http.NewServeMux()
	authController.Init(mux)

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", conf.Port),
		Handler: cors(mux),
	}

	logger.Info("server started", "addr", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	logger.Info("server shut down")
}

func init() {
	godotenv.Load()
}
