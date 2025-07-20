package app

import (
	"FinanceTracker/gateway/internal/config"
	"FinanceTracker/gateway/internal/middleware"
	"FinanceTracker/gateway/pkg/logger"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "FinanceTracker/gateway/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type app struct {
	srv    *http.Server
	logger *slog.Logger
}

type Controller interface {
	Init(r *http.ServeMux)
}

func New(log *slog.Logger, conf config.Config, controllers ...Controller) *app {
	mux := http.NewServeMux()
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	for _, c := range controllers {
		c.Init(mux)
	}

	corsMiddleware := middleware.NewCORS(middleware.CORSConfig{
		AllowedOrigins:   conf.CorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	loggerMiddleware := logger.NewHttpMiddleware(log)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Handler: loggerMiddleware(corsMiddleware(mux)),
	}

	return &app{
		logger: log,
		srv:    srv,
	}
}

func (a *app) Start() {
	a.logger.Info("starting server", "addr", a.srv.Addr)
	go func() {
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("failed to start server", "err", err)
			os.Exit(1)
		}
	}()
}

func (a *app) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	a.srv.Shutdown(ctx)
	a.logger.Info("server stopped")
}
