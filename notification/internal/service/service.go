package service

import (
	"context"
	"fmt"
	"log/slog"
)

type UserRepo interface {
	GetEmailByID(ctx context.Context, userID int) (string, error)
}

type Mailer interface {
	SendEmail(to, subject, body string) error
}

type notificationService struct {
	logger   *slog.Logger
	userRepo UserRepo
	mailer   Mailer
}

func NewNotificationService(logger *slog.Logger, userRepo UserRepo, mailer Mailer) *notificationService {
	return &notificationService{
		logger:   logger,
		userRepo: userRepo,
		mailer:   mailer,
	}
}

func (s *notificationService) SendOTP(ctx context.Context, userID int, code string) error {
	email, err := s.userRepo.GetEmailByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	return s.mailer.SendEmail(email, "OTP", code) // change
}
