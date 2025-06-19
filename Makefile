include .env
MIGRATIONS_PATH=./migrations

migrate-create:
	@name=$(name);
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(name)

migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(POSTGRES_URL) up

migrate-down:
	@name=$(name);
	@migrate -path=$(MIGRATIONS_PATH) -database=$(POSTGRES_URL) down $(name)

proto-gen:
	@protoc --go_out=./auth/pkg/ --go-grpc_out=./auth/pkg/  -I. proto/auth.proto
	@protoc --go_out=./gateway/pkg/ --go-grpc_out=./gateway/pkg/  -I. proto/auth.proto
