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
