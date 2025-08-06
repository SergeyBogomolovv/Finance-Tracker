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
	Create(ctx context.Context, data dto.CreateUserDto) (domain.User, error)
	MarkEmailVerified(ctx context.Context, userID int) error
}

type OTPRepo interface {
	Generate(ctx context.Context, userID int) (domain.OTP, error)
	Validate(ctx context.Context, userID int, code string) (bool, error)
	DeleteAll(ctx context.Context, userID int) error
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
	// check if user already exists
	user, err := s.users.GetByEmail(ctx, payload.Email)

	// if user not found, register new user
	if errors.Is(err, domain.ErrUserNotFound) {
		return s.oauthRegister(ctx, payload)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// check provider matches
	if user.Provider != domain.UserProvider(payload.Provider) {
		return "", domain.ErrProviderMismatch
	}

	// sign JWT token
	token, err := signToken(user.ID, s.jwtKey, s.jwtTTL)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	logger.Debug(ctx, "user logged in", "email", payload.Email, "provider", payload.Provider)

	return token, nil
}

func (s *authService) oauthRegister(ctx context.Context, payload dto.OAuthPayload) (string, error) {
	var token string
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// create new user
		user, err := s.users.Create(ctx, dto.CreateUserDto{
			Email:           payload.Email,
			Provider:        domain.UserProvider(payload.Provider),
			IsEmailVerified: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// send user registered event
		event := events.EventUserRegistered{
			UserID:    user.ID,
			Email:     user.Email,
			Provider:  string(user.Provider),
			AvatarURL: payload.AvatarUrl,
			FullName:  payload.FullName,
		}
		if err := s.producer.PublishUserRegistered(ctx, event); err != nil {
			return fmt.Errorf("failed to publish user registered event: %w", err)
		}

		// sign JWT token
		token, err = signToken(user.ID, s.jwtKey, s.jwtTTL)
		if err != nil {
			return fmt.Errorf("failed to sign token: %w", err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	logger.Debug(ctx, "user registered", "email", payload.Email, "provider", payload.Provider)

	return token, nil
}

func (c *authService) GenerateOTP(ctx context.Context, email string) error {
	// check if user exists
	user, err := c.users.GetByEmail(ctx, email)

	// if user not found, register new user
	if errors.Is(err, domain.ErrUserNotFound) {
		return c.emailRegister(ctx, email)
	}

	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// check provider matches
	if user.Provider != domain.UserProviderEmail {
		return domain.ErrProviderMismatch
	}

	return c.txManager.Do(ctx, func(ctx context.Context) error {
		// generate otp
		otp, err := c.otps.Generate(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to generate OTP: %w", err)
		}

		// send otp
		event := events.EventOTPGenerated{
			UserID:    user.ID,
			Email:     user.Email,
			Code:      otp.Code,
			ExpiresAt: otp.ExpiresAt,
		}
		if err := c.producer.PublishOTPGenerated(ctx, event); err != nil {
			return fmt.Errorf("failed to publish OTP generated event: %w", err)
		}

		logger.Debug(ctx, "OTP generated", "email", email)
		return nil
	})
}

func (c *authService) emailRegister(ctx context.Context, email string) error {
	return c.txManager.Do(ctx, func(ctx context.Context) error {
		// add user
		user, err := c.users.Create(ctx, dto.CreateUserDto{
			Email:           email,
			Provider:        domain.UserProviderEmail,
			IsEmailVerified: false,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// generate otp
		otp, err := c.otps.Generate(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("failed to generate OTP: %w", err)
		}

		// send otp
		event := events.EventOTPGenerated{
			UserID:    user.ID,
			Email:     user.Email,
			Code:      otp.Code,
			ExpiresAt: otp.ExpiresAt,
		}
		if err := c.producer.PublishOTPGenerated(ctx, event); err != nil {
			return fmt.Errorf("failed to publish OTP generated event: %w", err)
		}

		logger.Debug(ctx, "OTP generated", "email", email)
		return nil
	})
}

func (s *authService) VerifyOTP(ctx context.Context, email, code string) (string, error) {
	// get user by email
	user, err := s.users.GetByEmail(ctx, email)
	// check if email exists
	if errors.Is(err, domain.ErrUserNotFound) {
		return "", domain.ErrInvalidOTP
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// validate OTP
	valid, err := s.otps.Validate(ctx, user.ID, code)
	if err != nil {
		return "", fmt.Errorf("failed to validate OTP: %w", err)
	}
	if !valid {
		return "", domain.ErrInvalidOTP
	}

	var token string
	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		// delete used OTPs
		if err := s.otps.DeleteAll(ctx, user.ID); err != nil {
			return fmt.Errorf("failed to delete OTP: %w", err)
		}

		// sign JWT token
		token, err = signToken(user.ID, s.jwtKey, s.jwtTTL)
		if err != nil {
			return fmt.Errorf("failed to sign token: %w", err)
		}

		// notify user and add profile info if email is not verified
		if !user.IsEmailVerified {
			if err := s.users.MarkEmailVerified(ctx, user.ID); err != nil {
				return fmt.Errorf("failed to mark email as verified: %w", err)
			}
			event := events.EventUserRegistered{
				UserID:   user.ID,
				Email:    user.Email,
				Provider: string(user.Provider),
			}
			if err := s.producer.PublishUserRegistered(ctx, event); err != nil {
				return fmt.Errorf("failed to publish user registered event: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	logger.Debug(ctx, "OTP verified", "email", email)
	return token, nil
}

func signToken(userID int, secret []byte, ttl time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
