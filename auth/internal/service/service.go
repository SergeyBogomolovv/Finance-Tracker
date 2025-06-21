package service

import (
	"FinanceTracker/auth/internal/dto"
	"context"
	"fmt"
	"log/slog"
)

type authService struct {
	logger *slog.Logger
}

func NewAuthService(logger *slog.Logger) *authService {
	return &authService{
		logger: logger.With(slog.String("layer", "service")),
	}
}

func (s *authService) OAuth(ctx context.Context, payload dto.OAuthPayload) (string, error) {
	fmt.Println("registering user", payload)
	return "test_token", nil
}
