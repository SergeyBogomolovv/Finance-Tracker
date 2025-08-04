package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port int
	Host string
	Env  string

	KafkaBrokers      []string
	KafkaBatchTimeout time.Duration

	PostgresURL string

	OAuth OAuth

	JwtSecret []byte
	JwtTTL    time.Duration
}

type OAuth struct {
	RedirectURL        string
	GoogleClientID     string
	GoogleClientSecret string
	YandexClientID     string
	YandexClientSecret string
}

func New() Config {
	return Config{
		Port:              envInt("PORT", 50051),
		Host:              env("HOST", "localhost"),
		Env:               env("ENV", "development"),
		KafkaBrokers:      envArray("KAFKA_BROKERS", "localhost:9092"),
		KafkaBatchTimeout: envDuration("KAFKA_BATCH_TIMEOUT", 10*time.Millisecond),
		OAuth: OAuth{
			RedirectURL:        env("OAUTH_REDIRECT_URL", "http://localhost:8080"),
			GoogleClientID:     env("GOOGLE_CLIENT_ID"),
			GoogleClientSecret: env("GOOGLE_CLIENT_SECRET"),
			YandexClientID:     env("YANDEX_CLIENT_ID"),
			YandexClientSecret: env("YANDEX_CLIENT_SECRET"),
		},
		PostgresURL: env("POSTGRES_URL"),
		JwtTTL:      envDuration("JWT_TTL", 24*time.Hour),
		JwtSecret:   []byte(env("JWT_SECRET", "secret")),
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

func envDuration(key string, fallback ...time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	if len(fallback) == 0 {
		return 0
	}
	return fallback[0]
}

func envArray(key string, fallback ...string) []string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Split(value, ",")
	}
	if len(fallback) == 0 {
		return []string{}
	}
	return fallback
}
