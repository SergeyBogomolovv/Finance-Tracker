package service

import (
	"FinanceTracker/auth/internal/domain"
	"FinanceTracker/auth/internal/dto"
	"FinanceTracker/auth/pkg/events"
	"FinanceTracker/auth/pkg/logger"
	"FinanceTracker/auth/pkg/transaction"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, email, provider string) (domain.User, error)
}

type OTPRepo interface {
	Generate(ctx context.Context, email string, duration time.Duration) (domain.OTP, error)
	Verify(ctx context.Context, email, code string) (bool, error)
	MarkUsed(ctx context.Context, email, code string) error
}

type Producer interface {
	PublishUserRegistered(ctx context.Context, event events.EventUserRegistered) error
	PublishOTPGenerated(ctx context.Context, event events.EventOTPGenerated) error
}

type authService struct {
	otps      OTPRepo
	users     UserRepo
	txManager transaction.Manager
	producer  Producer
	jwtTTL    time.Duration
	jwtKey    []byte
}

func NewAuthService(users UserRepo, otps OTPRepo, producer Producer, txManager transaction.Manager, jwtTTL time.Duration, jwtKey []byte) *authService {
	return &authService{
		otps:      otps,
		producer:  producer,
		users:     users,
		jwtTTL:    jwtTTL,
		jwtKey:    jwtKey,
		txManager: txManager,
	}
}

func (s *authService) OAuth(ctx context.Context, payload dto.OAuthPayload) (string, error) {
	var token string

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// check if user already exists
		user, err := s.users.GetByEmail(ctx, payload.Email)

		// if user not found, register new user
		if errors.Is(err, domain.ErrUserNotFound) {
			// create user
			user, err = s.users.Create(ctx, payload.Email, payload.Provider)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			// publish event
			event := events.EventUserRegistered{
				UserID:    user.ID,
				Email:     user.Email,
				Provider:  user.Provider,
				AvatarURL: payload.AvatarUrl,
				FullName:  payload.FullName,
			}
			if err := s.producer.PublishUserRegistered(ctx, event); err != nil {
				return fmt.Errorf("failed to publish user registered event: %w", err)
			}
			logger.Debug(ctx, "user registered", "id", user.ID, "email", user.Email, "provider", user.Provider)
			// unknown get user error
		} else if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
			// check provider matches if user was found
		} else if user.Provider != payload.Provider {
			return domain.ErrProviderMismatch
		}

		// sign JWT token
		token, err = signToken(user.ID, s.jwtKey, s.jwtTTL)
		if err != nil {
			return fmt.Errorf("failed to sign token: %w", err)
		}
		logger.Debug(ctx, "user logined", "id", user.ID, "email", user.Email, "provider", user.Provider)
		return nil
	})

	return token, err
}

func (c *authService) GenerateOTP(ctx context.Context, email string) error {
	const duration = 5 * time.Minute
	return c.txManager.Do(ctx, func(ctx context.Context) error {
		// check user provider
		user, err := c.users.GetByEmail(ctx, email)
		if err != nil {
			if !errors.Is(err, domain.ErrUserNotFound) {
				return fmt.Errorf("failed to check user provider: %w", err)
			}
		} else if user.Provider != domain.UserProviderEmail {
			return domain.ErrProviderMismatch
		}

		// generate otp
		otp, err := c.otps.Generate(ctx, email, duration)
		if err != nil {
			return fmt.Errorf("failed to generate otp: %w", err)
		}

		// send otp
		event := events.EventOTPGenerated{
			Email:     otp.Email,
			Code:      otp.Code,
			ExpiresAt: otp.ExpiresAt,
			CreatedAt: otp.CreatedAt,
		}
		if err := c.producer.PublishOTPGenerated(ctx, event); err != nil {
			return fmt.Errorf("failed to publish OTP generated event: %w", err)
		}

		logger.Debug(ctx, "OTP generated", "email", email)
		return nil
	})
}

func (s *authService) VerifyOTP(ctx context.Context, email, code string) (string, error) {
	var token string
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// check is otp valid
		valid, err := s.otps.Verify(ctx, email, code)
		if err != nil {
			return fmt.Errorf("failed to verify otp: %w", err)
		}
		if !valid {
			return domain.ErrInvalidOTP
		}

		// mark otp used
		if err := s.otps.MarkUsed(ctx, email, code); err != nil {
			return fmt.Errorf("failed to mark otp used: %w", err)
		}

		// get user info
		user, err := s.users.GetByEmail(ctx, email)

		// if user not registered
		if errors.Is(err, domain.ErrUserNotFound) {
			// create user
			user, err = s.users.Create(ctx, email, domain.UserProviderEmail)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			// send message to kafka
			event := events.EventUserRegistered{
				UserID:   user.ID,
				Email:    user.Email,
				Provider: user.Provider,
			}
			if err := s.producer.PublishUserRegistered(ctx, event); err != nil {
				return fmt.Errorf("failed to publish user registered event: %w", err)
			}
			logger.Debug(ctx, "user registered", "id", user.ID, "email", user.Email, "provider", user.Provider)
		} else if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		// check is user registered by email
		if user.Provider != domain.UserProviderEmail {
			return domain.ErrProviderMismatch
		}
		// sign jwt token
		token, err = signToken(user.ID, s.jwtKey, s.jwtTTL)
		if err != nil {
			return fmt.Errorf("failed to sign token: %w", err)
		}
		logger.Debug(ctx, "user logined", "id", user.ID, "email", user.Email, "provider", user.Provider)
		return nil
	})

	return token, err
}

func signToken(userID int, secret []byte, ttl time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
