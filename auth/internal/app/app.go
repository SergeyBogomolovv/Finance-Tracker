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

type Controller interface {
	Register(server *grpc.Server)
}

func New(logger *slog.Logger, controllers ...Controller) *app {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(log.UnaryInterceptor(logger)),
	)

	for _, c := range controllers {
		c.Register(server)
	}

	return &app{logger: logger, srv: server}
}

func (a *app) Start(host string, port int) {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		exitIfErr(a.logger, err, "failed to listen")
		a.logger.Info("server started", "addr", lis.Addr())
		err = a.srv.Serve(lis)
		exitIfErr(a.logger, err, "failed to serve")
	}()
}

func (a *app) Stop() {
	a.srv.GracefulStop()
	a.logger.Info("server stopped")
}

func exitIfErr(logger *slog.Logger, err error, msg string) {
	if err != nil {
		logger.Error(msg, "err", err)
		os.Exit(1)
	}
}
