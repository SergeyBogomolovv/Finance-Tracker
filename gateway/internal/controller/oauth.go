package controller

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	pb "FinanceTracker/gateway/pkg/api/auth"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/yandex"
)

type oauthController struct {
	googleConfig       *oauth2.Config
	yandexConfig       *oauth2.Config
	clientRedirectAddr string
	authService        pb.AuthServiceClient
}

func NewOAuthController(
	oauthRedirectAddr, clientRedirectAddr, googleClientID, yandexClientID string,
	authService pb.AuthServiceClient) *oauthController {
	return &oauthController{
		googleConfig: &oauth2.Config{
			ClientID:    googleClientID,
			RedirectURL: fmt.Sprintf("%s/auth/google/callback", oauthRedirectAddr),
			Endpoint:    google.Endpoint,
			Scopes:      []string{"email", "profile", "openid"},
		},
		yandexConfig: &oauth2.Config{
			ClientID:    yandexClientID,
			RedirectURL: fmt.Sprintf("%s/auth/yandex/callback", oauthRedirectAddr),
			Endpoint:    yandex.Endpoint,
		},
		clientRedirectAddr: clientRedirectAddr,
		authService:        authService,
	}
}

func (c *oauthController) Init(r *http.ServeMux) {
	r.HandleFunc("/auth/google/login", c.handleGoogleLogin)
	r.HandleFunc("/auth/yandex/login", c.handleYandexLogin)
	r.HandleFunc("/auth/google/callback", c.handleGoogleCallback)
	r.HandleFunc("/auth/yandex/callback", c.handleYandexCallback)
}

func (c *oauthController) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	setStateToCookie(w, state)
	url := c.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *oauthController) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !checkState(r) {
		http.Redirect(w, r, c.clientRedirectAddr+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	resp, err := c.authService.ExchangeGoogleOAuth(r.Context(), &pb.OAuthRequest{Code: code})
	if err != nil {
		http.Redirect(w, r, c.clientRedirectAddr+"?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	setTokenToCookie(w, resp.Token)
	http.Redirect(w, r, c.clientRedirectAddr, http.StatusTemporaryRedirect)
}

func (c *oauthController) handleYandexLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	setStateToCookie(w, state)
	url := c.yandexConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *oauthController) handleYandexCallback(w http.ResponseWriter, r *http.Request) {
	if !checkState(r) {
		http.Redirect(w, r, c.clientRedirectAddr+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	resp, err := c.authService.ExchangeYandexOAuth(r.Context(), &pb.OAuthRequest{Code: code})
	if err != nil {
		http.Redirect(w, r, c.clientRedirectAddr+"?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	setTokenToCookie(w, resp.Token)
	http.Redirect(w, r, c.clientRedirectAddr, http.StatusTemporaryRedirect)
}

func checkState(r *http.Request) bool {
	expectedState, err := r.Cookie("oauth_state")
	if err != nil {
		return false
	}

	actualState := r.URL.Query().Get("state")
	return actualState == expectedState.Value
}

func setTokenToCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func setStateToCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
