APP_NAME = gateway
BUILD_DIR = bin
MAIN = cmd/main.go
SWAGGER_DIR = docs

.PHONY: run build test lint swag clean deps coverage

run:
	go run $(MAIN)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

test:
	go test ./... -v

lint:
	golangci-lint run

swag:
	swag init -g $(MAIN) -o $(SWAGGER_DIR)

clean:
	rm -rf $(BUILD_DIR)

deps:
	go mod tidy
	go mod download

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

check: lint test coverage