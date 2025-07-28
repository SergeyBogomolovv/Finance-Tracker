package service

import (
	"FinanceTracker/auth/internal/domain"
	"FinanceTracker/auth/internal/dto"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserRepo interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
}

type Producer interface {
	PublishUserRegistered(ctx context.Context, userID int) error
	PublishOTPGenerated(ctx context.Context, userID int, code string) error
}

type authService struct {
	users    UserRepo
	producer Producer
	jwtTTL   time.Duration
	jwtKey   []byte
}

func NewAuthService(users UserRepo, producer Producer, jwtTTL time.Duration, jwtKey []byte) *authService {
	return &authService{
		producer: producer,
		users:    users,
		jwtTTL:   jwtTTL,
		jwtKey:   jwtKey,
	}
}

func (s *authService) OAuth(ctx context.Context, payload dto.OAuthPayload) (string, error) {
	user, err := s.users.GetByEmail(ctx, payload.Email)

	if errors.Is(err, domain.ErrUserNotFound) {
		// TODO: ОБЕРНУТЬ В ТРАНЗАКЦИЮ И ВЫНЕСТИ В ОТДЕЛЬНЫЙ МЕТОД РЕГИСТРАЦИИ
		user, err = s.users.Create(ctx, domain.User{
			Email:           payload.Email,
			FullName:        payload.FullName,
			AvatarUrl:       payload.AvatarUrl,
			Provider:        domain.UserProvider(payload.Provider),
			IsEmailVerified: true,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
		token, err := signToken(user.ID, []byte("secret"), 24*time.Hour)
		if err != nil {
			return "", fmt.Errorf("failed to sign token: %w", err)
		}
		if err := s.producer.PublishUserRegistered(ctx, user.ID); err != nil {
			return "", fmt.Errorf("failed to publish user registered event: %w", err)
		}
		return token, nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user.Provider != domain.UserProvider(payload.Provider) {
		return "", domain.ErrProviderMismatch
	}

	token, err := signToken(user.ID, s.jwtKey, s.jwtTTL)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
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
