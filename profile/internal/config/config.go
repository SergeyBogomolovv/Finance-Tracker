package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port int
	Host string
	Env  string

	KafkaBrokers []string
	KafkaGroupID string

	S3AccessKey string
	S3SecretKey string
	S3Endpoint  string
	S3Region    string

	PostgresURL string
}

func New() Config {
	return Config{
		Port:         envInt("PORT", 50052),
		Host:         env("HOST", "localhost"),
		Env:          env("ENV", "development"),
		PostgresURL:  env("POSTGRES_URL"),
		S3AccessKey:  env("S3_ACCESS_KEY"),
		S3SecretKey:  env("S3_SECRET_KEY"),
		S3Endpoint:   env("S3_ENDPOINT", "http://localhost:9000"),
		S3Region:     env("S3_REGION", "local"),
		KafkaBrokers: envArray("KAFKA_BROKERS", "localhost:9092"),
		KafkaGroupID: env("KAFKA_GROUP_ID", "profile-service"),
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

func envArray(key string, fallback ...string) []string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Split(value, ",")
	}
	if len(fallback) == 0 {
		return []string{}
	}
	return fallback
}
