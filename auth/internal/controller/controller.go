package controller

import (
	"FinanceTracker/auth/internal/config"
	"FinanceTracker/auth/internal/domain"
	"FinanceTracker/auth/internal/dto"
	pb "FinanceTracker/auth/pkg/api/auth"
	"FinanceTracker/auth/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/yandex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	OAuth(ctx context.Context, payload dto.OAuthPayload) (string, error)
	GenerateOTP(ctx context.Context, email string) error
	VerifyOTP(ctx context.Context, email, otp string) (string, error)
}

type authController struct {
	pb.UnimplementedAuthServiceServer
	googleConfig *oauth2.Config
	yandexConfig *oauth2.Config
	authService  AuthService
	validate     *validator.Validate
}

func NewAuthController(authService AuthService, oauthConf config.OAuth) *authController {
	return &authController{
		googleConfig: &oauth2.Config{
			ClientID:     oauthConf.GoogleClientID,
			ClientSecret: oauthConf.GoogleClientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/google/callback", oauthConf.RedirectURL),
			Endpoint:     google.Endpoint,
			Scopes:       []string{"email", "profile", "openid"},
		},
		yandexConfig: &oauth2.Config{
			ClientID:     oauthConf.YandexClientID,
			ClientSecret: oauthConf.YandexClientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/yandex/callback", oauthConf.RedirectURL),
			Endpoint:     yandex.Endpoint,
		},
		authService: authService,
		validate:    validator.New(),
	}
}

func (c *authController) Register(server *grpc.Server) {
	pb.RegisterAuthServiceServer(server, c)
}

type GooglePayload struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (c *authController) ExchangeGoogleOAuth(ctx context.Context, req *pb.OAuthRequest) (*pb.AuthResponse, error) {
	token, err := c.googleConfig.Exchange(ctx, req.Code)
	if err != nil {
		logger.Error(ctx, "failed to exchange token", "err", err)
		return nil, status.Error(codes.Unauthenticated, "failed to exchange token")
	}

	client := c.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "failed to get user info", "err", err)
		return nil, status.Error(codes.Unauthenticated, "failed to get user info")
	}
	defer resp.Body.Close()

	var data GooglePayload
	json.NewDecoder(resp.Body).Decode(&data)

	accessToken, err := c.authService.OAuth(ctx, dto.OAuthPayload{
		Email:     data.Email,
		FullName:  data.Name,
		AvatarUrl: data.Picture,
		Provider:  dto.OAuthProviderGoogle,
	})
	if err != nil {
		logger.Error(ctx, "failed to oauth user", "err", err)
		return nil, status.Error(codes.Unauthenticated, "failed to oauth user")
	}

	return &pb.AuthResponse{AccessToken: accessToken}, nil
}

type YandexPayload struct {
	Email         string `json:"default_email"`
	Name          string `json:"real_name"`
	AvatarID      string `json:"default_avatar_id"`
	IsAvatarEmpty bool   `json:"is_avatar_empty"`
}

func (c *authController) ExchangeYandexOAuth(ctx context.Context, req *pb.OAuthRequest) (*pb.AuthResponse, error) {
	token, err := c.yandexConfig.Exchange(ctx, req.Code)
	if err != nil {
		logger.Error(ctx, "failed to exchange token", "err", err)
		return nil, status.Error(codes.Unauthenticated, "failed to exchange token")
	}

	client := c.yandexConfig.Client(ctx, token)
	resp, err := client.Get("https://login.yandex.ru/info")
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "failed to get user info", "err", err)
		return nil, status.Error(codes.Unauthenticated, "failed to get user info")
	}
	defer resp.Body.Close()

	var data YandexPayload
	json.NewDecoder(resp.Body).Decode(&data)

	accessToken, err := c.authService.OAuth(ctx, dto.OAuthPayload{
		Email:     data.Email,
		FullName:  data.Name,
		AvatarUrl: fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", data.AvatarID),
		Provider:  dto.OAuthProviderYandex,
	})
	if err != nil {
		logger.Error(ctx, "failed to oauth user", "err", err)
		return nil, status.Error(codes.Unauthenticated, "failed to oauth user")
	}

	return &pb.AuthResponse{AccessToken: accessToken}, nil
}

func (c *authController) GenerateOTP(ctx context.Context, req *pb.GenerateOTPRequest) (*pb.GenerateOTPResponse, error) {
	if err := c.validate.Var(req.Email, "email"); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	err := c.authService.GenerateOTP(ctx, req.Email)
	if errors.Is(err, domain.ErrProviderMismatch) {
		return nil, status.Error(codes.InvalidArgument, "invalid provider")
	}
	if err != nil {
		logger.Error(ctx, "failed to generate email OTP", "err", err)
		return nil, status.Error(codes.Internal, "failed to generate email OTP")
	}
	return &pb.GenerateOTPResponse{}, nil
}

func (c *authController) VerifyOTP(ctx context.Context, req *pb.VerifyOTPRequest) (*pb.AuthResponse, error) {
	if err := c.validate.Var(req.Email, "email"); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}
	if err := c.validate.Var(req.Otp, "required"); err != nil {
		return nil, status.Error(codes.InvalidArgument, "OTP is required")
	}

	accessToken, err := c.authService.VerifyOTP(ctx, req.Email, req.Otp)
	if errors.Is(err, domain.ErrInvalidOTP) {
		return nil, status.Error(codes.Unauthenticated, "invalid OTP")
	}

	if err != nil {
		logger.Error(ctx, "failed to verify email OTP", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify email OTP")
	}

	return &pb.AuthResponse{AccessToken: accessToken}, nil
}
