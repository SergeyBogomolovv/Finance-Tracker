package controller

import (
	pb "FinanceTracker/auth/pkg/api/auth"
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/yandex"
)

type oauthController struct {
	pb.UnimplementedAuthServiceServer
	googleConfig *oauth2.Config
	yandexConfig *oauth2.Config
}

func NewOAuthController(
	oauthRedirectAddr string,
	googleClientID string,
	googleClientSecret string,
	yandexClientID string,
	yandexClientSecret string,
) *oauthController {
	return &oauthController{
		googleConfig: &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/google/callback", oauthRedirectAddr),
			Endpoint:     google.Endpoint,
			Scopes:       []string{"email", "profile", "openid"},
		},
		yandexConfig: &oauth2.Config{
			ClientID:     yandexClientID,
			ClientSecret: yandexClientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/yandex/callback", oauthRedirectAddr),
			Endpoint:     yandex.Endpoint,
		},
	}
}

func (c *oauthController) ExchangeGoogleOAuth(ctx context.Context, req *pb.OAuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "test_token"}, nil
}

func (c *oauthController) ExchangeYandexOAuth(ctx context.Context, req *pb.OAuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "test_token"}, nil

}
