# Переменные
include .env

MIGRATIONS_PATH ?= ./migrations
POSTGRES_URL ?= $(POSTGRES_URL)
MINIO_ROOT_USER ?= $(MINIO_ROOT_USER)
MINIO_ROOT_PASSWORD ?= $(MINIO_ROOT_PASSWORD)
MINIO_BUCKET = finance-tracker

.PHONY: migrate-create migrate-up migrate-down proto-gen dev-up dev-down dev-restart mc-setup

# Миграции
migrate-create:
ifndef name
	$(error "Usage: make migrate-create name=MigrationName")
endif
	migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(name)

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" up

migrate-down:
ifndef name
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" down 1
else
	migrate -path $(MIGRATIONS_PATH) -database "$(POSTGRES_URL)" down $(name)
endif

# gRPC
proto-gen:
	protoc --go_out=./auth/pkg/ --go-grpc_out=./auth/pkg/ -I. proto/auth.proto
	protoc --go_out=./gateway/pkg/ --go-grpc_out=./gateway/pkg/ -I. proto/auth.proto
	protoc --go_out=./profile/pkg/ --go-grpc_out=./profile/pkg/ -I. proto/profile.proto
	protoc --go_out=./gateway/pkg/ --go-grpc_out=./gateway/pkg/ -I. proto/profile.proto

# Docker Compose Dev
dev-up:
	docker compose -f docker-compose.dev.yml up  -d

dev-down:
	docker compose -f docker-compose.dev.yml down

dev-restart:
	docker compose -f docker-compose.dev.yml restart

# Minio
mc-setup:
	mc alias set local http://localhost:9000 $(MINIO_ROOT_USER) $(MINIO_ROOT_PASSWORD)
	mc mb --ignore-existing local/$(MINIO_BUCKET)
	mc anonymous set download local/$(MINIO_BUCKET)