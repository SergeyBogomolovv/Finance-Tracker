package logger

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

func UnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		return handler(WithLogger(ctx, logger), req)
	}
}
