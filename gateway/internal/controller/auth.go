package controller

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	pb "FinanceTracker/gateway/pkg/api/auth"
	"FinanceTracker/gateway/pkg/utils"

	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/yandex"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authController struct {
	validate           *validator.Validate
	googleConfig       *oauth2.Config
	yandexConfig       *oauth2.Config
	clientRedirectAddr string
	authService        pb.AuthServiceClient
}

func NewAuthController(
	authService pb.AuthServiceClient,
	oauthRedirectAddr string,
	clientRedirectAddr string,
	googleClientID string,
	yandexClientID string,
) *authController {
	return &authController{
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
		validate:           validator.New(),
	}
}

func (c *authController) Init(r *http.ServeMux) {
	r.HandleFunc("/auth/google/login", c.handleGoogleLogin)
	r.HandleFunc("/auth/yandex/login", c.handleYandexLogin)
	r.HandleFunc("/auth/google/callback", c.handleGoogleCallback)
	r.HandleFunc("/auth/yandex/callback", c.handleYandexCallback)
	r.HandleFunc("POST /auth/email", c.handleEmailAuth)
	r.HandleFunc("POST /auth/email/verify", c.handleVerifyEmailOTP)
}

func (c *authController) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	setStateToCookie(w, state)
	url := c.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *authController) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !checkState(r) {
		http.Redirect(w, r, c.clientRedirectAddr+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	resp, err := c.authService.ExchangeGoogleOAuth(r.Context(), &pb.OAuthRequest{Code: code})
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, c.clientRedirectAddr+"?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	setTokenToCookie(w, resp.AccessToken)
	http.Redirect(w, r, c.clientRedirectAddr, http.StatusTemporaryRedirect)
}

func (c *authController) handleYandexLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	setStateToCookie(w, state)
	url := c.yandexConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *authController) handleYandexCallback(w http.ResponseWriter, r *http.Request) {
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

	setTokenToCookie(w, resp.AccessToken)
	http.Redirect(w, r, c.clientRedirectAddr, http.StatusTemporaryRedirect)
}

type EmailAuthRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (c *authController) handleEmailAuth(w http.ResponseWriter, r *http.Request) {
	var req EmailAuthRequest
	if err := utils.DecodeBody(r, &req); err != nil {
		utils.WriteError(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := c.validate.Struct(req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	_, err := c.authService.SendEmailOTP(r.Context(), &pb.SendEmailOTPRequest{Email: req.Email})
	if err != nil {
		utils.WriteError(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	utils.WriteMessage(w, "email sent")
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

func (c *authController) handleVerifyEmailOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := utils.DecodeBody(r, &req); err != nil {
		utils.WriteError(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := c.validate.Struct(req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	resp, err := c.authService.VerifyEmailOTP(r.Context(), &pb.VerifyEmailOTPRequest{Email: req.Email, Otp: req.OTP})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.InvalidArgument:
				utils.WriteError(w, "invalid email or otp", http.StatusBadRequest)
				return
			}
		}
		utils.WriteError(w, "failed to verify email", http.StatusInternalServerError)
		return
	}

	setTokenToCookie(w, resp.AccessToken)

	if resp.IsNewUser {
		utils.WriteMessage(w, "email verified and user created")
	} else {
		utils.WriteMessage(w, "successfully logged in")
	}
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
		Name:     "access_token",
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
