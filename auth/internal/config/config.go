package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port int

	GoogleClientID     string
	GoogleClientSecret string

	YandexClientID     string
	YandexClientSecret string

	OAuthRedirectURL string
}

func New() Config {
	return Config{
		Port:               envInt("PORT", 50051),
		GoogleClientID:     env("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: env("GOOGLE_CLIENT_SECRET"),
		YandexClientID:     env("YANDEX_CLIENT_ID"),
		YandexClientSecret: env("YANDEX_CLIENT_SECRET"),
		OAuthRedirectURL:   env("OAUTH_REDIRECT_URL", "http://localhost:8080"),
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
