package controller

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"FinanceTracker/gateway/internal/config"
	pb "FinanceTracker/gateway/pkg/api/auth"
	"FinanceTracker/gateway/pkg/logger"
	"FinanceTracker/gateway/pkg/utils"

	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/yandex"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authController struct {
	validate          *validator.Validate
	googleConfig      *oauth2.Config
	yandexConfig      *oauth2.Config
	clientRedirectURL string
	authService       pb.AuthServiceClient
}

func NewAuthController(authService pb.AuthServiceClient, oauthConf config.OAuth) *authController {
	return &authController{
		googleConfig: &oauth2.Config{
			ClientID:    oauthConf.GoogleClientID,
			RedirectURL: fmt.Sprintf("%s/auth/google/callback", oauthConf.RedirectURL),
			Endpoint:    google.Endpoint,
			Scopes:      []string{"email", "profile", "openid"},
		},
		yandexConfig: &oauth2.Config{
			ClientID:    oauthConf.YandexClientID,
			RedirectURL: fmt.Sprintf("%s/auth/yandex/callback", oauthConf.RedirectURL),
			Endpoint:    yandex.Endpoint,
		},
		clientRedirectURL: oauthConf.ClientRedirectURL,
		authService:       authService,
		validate:          validator.New(),
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

// @Summary		Google OAuth вход
// @Description	Перенаправляет пользователя на Google OAuth страницу
// @Tags			auth
// @Produce		json
// @Success		307	{string}	string	"Redirect to Google OAuth"
// @Router			/auth/google/login [get]
func (c *authController) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateOAuthState()
	setOAuthStateToCookie(w, state)
	url := c.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// @Summary		Callback Google OAuth
// @Description	Обрабатывает redirect от Google и выдает access token
// @Tags			auth
// @Produce		json
// @Param			code	query		string	true	"Authorization Code"
// @Success		307		{string}	string	"Redirect to frontend"
// @Failure		307		{string}	string	"Redirect with error"
// @Router			/auth/google/callback [get]
func (c *authController) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !checkOAuthState(r) {
		logger.Debug(ctx, "invalid state")
		http.Redirect(w, r, c.clientRedirectURL+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	resp, err := c.authService.ExchangeGoogleOAuth(r.Context(), &pb.OAuthRequest{Code: code})
	if err != nil {
		logger.Error(ctx, "failed to exchange google oauth", "err", err)
		http.Redirect(w, r, c.clientRedirectURL+"?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	setAuthTokenToCookie(w, resp.AccessToken)
	http.Redirect(w, r, c.clientRedirectURL, http.StatusTemporaryRedirect)
}

// @Summary		Yandex OAuth вход
// @Description	Перенаправляет пользователя на Yandex OAuth страницу
// @Tags			auth
// @Produce		json
// @Success		307	{string}	string	"Redirect to Yandex OAuth"
// @Router			/auth/yandex/login [get]
func (c *authController) handleYandexLogin(w http.ResponseWriter, r *http.Request) {
	state := generateOAuthState()
	setOAuthStateToCookie(w, state)
	url := c.yandexConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// @Summary		Callback Yandex OAuth
// @Description	Обрабатывает redirect от Yandex и выдает access token
// @Tags			auth
// @Produce		json
// @Param			code	query		string	true	"Authorization Code"
// @Success		307		{string}	string	"Redirect to frontend"
// @Failure		307		{string}	string	"Redirect with error"
// @Router			/auth/yandex/callback [get]
func (c *authController) handleYandexCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !checkOAuthState(r) {
		logger.Debug(ctx, "invalid state")
		http.Redirect(w, r, c.clientRedirectURL+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	resp, err := c.authService.ExchangeYandexOAuth(r.Context(), &pb.OAuthRequest{Code: code})
	if err != nil {
		logger.Error(ctx, "failed to exchange yandex oauth", "err", err)
		http.Redirect(w, r, c.clientRedirectURL+"?error=oauth_failed", http.StatusTemporaryRedirect)
		return
	}

	setAuthTokenToCookie(w, resp.AccessToken)
	http.Redirect(w, r, c.clientRedirectURL, http.StatusTemporaryRedirect)
}

type EmailAuthRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// @Summary		Запросить код на email
// @Description	Отправляет одноразовый код на email
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		EmailAuthRequest				true	"Email для отправки OTP"
// @Success		200		{object}	utils.MessageResponse			"Email sent"
// @Failure		400		{object}	utils.ValidationErrorResponse	"Некорректные данные"
// @Failure		500		{object}	utils.ErrorResponse				"Сбой при отправке"
// @Router			/auth/email [post]
func (c *authController) handleEmailAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req EmailAuthRequest
	if err := utils.DecodeBody(r, &req); err != nil {
		logger.Error(ctx, "failed to decode body", "err", err)
		utils.WriteError(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := c.validate.Struct(req); err != nil {
		logger.Debug(ctx, "invalid request", "err", err)
		utils.WriteValidationError(w, err)
		return
	}

	_, err := c.authService.SendEmailOTP(r.Context(), &pb.SendEmailOTPRequest{Email: req.Email})
	if err != nil {
		logger.Error(ctx, "failed to send email", "err", err)
		utils.WriteError(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	utils.WriteMessage(w, "email sent")
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

// @Summary		Подтверждение email-кода
// @Description	Подтверждает OTP-код и возвращает access token
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		VerifyEmailRequest		true	"Email и OTP-код"
// @Success		200		{object}	utils.MessageResponse	"Email verified or login successful"
// @Failure		400		{object}	utils.ErrorResponse		"Неверные данные или код"
// @Failure		500		{object}	utils.ErrorResponse		"Внутренняя ошибка"
// @Router			/auth/email/verify [post]
func (c *authController) handleVerifyEmailOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req VerifyEmailRequest
	if err := utils.DecodeBody(r, &req); err != nil {
		logger.Error(ctx, "failed to decode body", "err", err)
		utils.WriteError(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := c.validate.Struct(req); err != nil {
		logger.Debug(ctx, "invalid request", "err", err)
		utils.WriteValidationError(w, err)
		return
	}

	resp, err := c.authService.VerifyEmailOTP(r.Context(), &pb.VerifyEmailOTPRequest{Email: req.Email, Otp: req.OTP})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.InvalidArgument:
				logger.Debug(ctx, "invalid email or otp", "err", err)
				utils.WriteError(w, "invalid email or otp", http.StatusBadRequest)
				return
			}
		}

		logger.Error(ctx, "failed to verify email", "err", err)
		utils.WriteError(w, "failed to verify email", http.StatusInternalServerError)
		return
	}

	setAuthTokenToCookie(w, resp.AccessToken)

	if resp.IsNewUser {
		utils.WriteMessage(w, "email verified and user created")
	} else {
		utils.WriteMessage(w, "successfully logged in")
	}
}

const (
	oauthStateCookieName  = "oauth_state"
	accessTokenCookieName = "access_token"
)

func checkOAuthState(r *http.Request) bool {
	expectedState, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return false
	}

	actualState := r.URL.Query().Get("state")
	return actualState == expectedState.Value
}

func setAuthTokenToCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func setOAuthStateToCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
