package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              int
	OAuthRedirectURL  string
	GoogleClientID    string
	YandexClientID    string
	ClientRedirectURL string
	AuthServiceAddr   string
	CorsOrigins       []string
}

func New() Config {
	return Config{
		Port:              envInt("PORT", 8080),
		OAuthRedirectURL:  env("OAUTH_REDIRECT_URL", "http://localhost:8080"),
		GoogleClientID:    env("GOOGLE_CLIENT_ID"),
		YandexClientID:    env("YANDEX_CLIENT_ID"),
		ClientRedirectURL: env("CLIENT_REDIRECT_URL", "http://localhost:3000"),
		AuthServiceAddr:   env("AUTH_SERVICE_ADDR", "localhost:50051"),
		CorsOrigins:       strings.Split(env("CORS_ORIGINS", "http://localhost:3000"), ","),
	}
}

func env(key string, fallback ...string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if len(fallback) == 0 {
		return ""
	}
	return fallback[0]
}

func envInt(key string, fallback ...int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	if len(fallback) == 0 {
		return 0
	}
	return fallback[0]
}
