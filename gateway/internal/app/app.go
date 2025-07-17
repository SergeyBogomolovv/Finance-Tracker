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
)

type app struct {
	srv    *http.Server
	logger *slog.Logger
	mux    *http.ServeMux
}

func New(log *slog.Logger, conf config.Config) *app {
	mux := http.NewServeMux()

	corsMiddleware := middleware.NewCORS(middleware.CORSConfig{
		AllowedOrigins:   conf.CorsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	loggerMiddleware := logger.NewHttpMiddleware(log)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Port),
		Handler: loggerMiddleware(corsMiddleware(mux)),
	}

	return &app{
		logger: log,
		srv:    srv,
		mux:    mux,
	}
}

type Controller interface {
	Init(r *http.ServeMux)
}

func (a *app) Init(controller ...Controller) {
	for _, c := range controller {
		c.Init(a.mux)
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
