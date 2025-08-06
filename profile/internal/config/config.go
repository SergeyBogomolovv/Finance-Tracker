package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port int
	Host string
	Env  string

	PostgresURL string
}

func New() Config {
	return Config{
		Port:        envInt("PORT", 50052),
		Host:        env("HOST", "localhost"),
		Env:         env("ENV", "development"),
		PostgresURL: env("POSTGRES_URL"),
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
