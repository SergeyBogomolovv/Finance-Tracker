package controller

import (
	"FinanceTracker/auth/internal/config"
	"FinanceTracker/auth/internal/dto"
	pb "FinanceTracker/auth/pkg/api/auth"
	"FinanceTracker/auth/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/yandex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	OAuth(ctx context.Context, payload dto.OAuthPayload) (string, error)
}

type authController struct {
	pb.UnimplementedAuthServiceServer
	googleConfig *oauth2.Config
	yandexConfig *oauth2.Config
	authService  AuthService
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

func (c *authController) SendEmailOTP(ctx context.Context, req *pb.SendEmailOTPRequest) (*pb.SendEmailOTPResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendEmailOTP not implemented")
}

func (c *authController) VerifyEmailOTP(ctx context.Context, req *pb.VerifyEmailOTPRequest) (*pb.AuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyEmailOTP not implemented")
}
