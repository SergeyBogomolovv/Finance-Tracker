APP_NAME = profile
BUILD_DIR = bin
MAIN = cmd/main.go

.PHONY: run build test lint clean deps coverage

run:
	go run $(MAIN)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

test:
	go test ./... -v

lint:
	golangci-lint run

clean:
	rm -rf $(BUILD_DIR)

deps:
	go mod tidy
	go mod download

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

check: lint test coverage