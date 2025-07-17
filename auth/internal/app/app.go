package app

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	log "FinanceTracker/auth/pkg/logger"

	"google.golang.org/grpc"
)

type app struct {
	logger *slog.Logger
	srv    *grpc.Server
}

func New(logger *slog.Logger) *app {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(log.UnaryInterceptor(logger)),
	)
	return &app{logger: logger, srv: server}
}

type Controller interface {
	Register(server *grpc.Server)
}

func (a *app) Register(controller ...Controller) {
	for _, c := range controller {
		c.Register(a.srv)
	}
	a.logger.Info("controllers registered")
}

func (a *app) Start(port int) {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			a.logger.Error("failed to listen", "err", err)
			os.Exit(1)
		}
		a.logger.Info("server started", "addr", lis.Addr())

		if err := a.srv.Serve(lis); err != nil {
			a.logger.Error("failed to listen", "err", err)
			os.Exit(1)
		}
	}()
}

func (a *app) Stop() {
	a.srv.GracefulStop()
	a.logger.Info("server stopped")
}
