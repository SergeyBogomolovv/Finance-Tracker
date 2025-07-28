package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env string

	KafkaGroupID string
	KafkaBrokers []string

	SMTP        SMTP
	PostgresURL string
}

type SMTP struct {
	Host string
	Port int
	User string
	Pass string
}

func New() *Config {
	return &Config{
		Env:          env("ENV", "development"),
		KafkaGroupID: env("KAFKA_GROUP_ID", "notification-service"),
		KafkaBrokers: envArray("KAFKA_BROKERS", "localhost:9092"),
		PostgresURL:  env("POSTGRES_URL"),
		SMTP: SMTP{
			Host: env("SMTP_HOST"),
			Port: envInt("SMTP_PORT"),
			User: env("SMTP_USER"),
			Pass: env("SMTP_PASS"),
		},
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

func envArray(key string, fallback ...string) []string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Split(value, ",")
	}
	if len(fallback) == 0 {
		return []string{}
	}
	return fallback
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
