package service

import (
	"FinanceTracker/notification/pkg/logger"
	"context"
	"fmt"
)

type UserRepo interface {
	GetEmailByID(ctx context.Context, userID int) (string, error)
}

type Mailer interface {
	SendOTPEmail(to string, otp string) error
	SendRegistrationEmail(to string) error
}

type notificationService struct {
	userRepo UserRepo
	mailer   Mailer
}

func NewNotificationService(userRepo UserRepo, mailer Mailer) *notificationService {
	return &notificationService{
		userRepo: userRepo,
		mailer:   mailer,
	}
}

func (s *notificationService) SendOTP(ctx context.Context, userID int, code string) error {
	email, err := s.userRepo.GetEmailByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	logger.Debug(ctx, "sending otp email", "email", email)
	return s.mailer.SendOTPEmail(email, code)
}

func (s *notificationService) SendRegistered(ctx context.Context, userID int) error {
	email, err := s.userRepo.GetEmailByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	return s.mailer.SendRegistrationEmail(email)
}
