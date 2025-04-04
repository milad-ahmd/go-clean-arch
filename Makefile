.PHONY: build run test clean lint

# Build variables
BINARY_NAME=clean-arch-api
BUILD_DIR=./bin

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api

# Run the application
run:
	$(GOCMD) run ./cmd/api

# Test the application
test:
	$(GOTEST) -v ./...

# Clean the binary
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Download dependencies
deps:
	$(GOMOD) tidy

# Lint the code
lint:
	$(GOLINT) run

# Run tests with coverage
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Generate mocks for testing
mocks:
	mockgen -source=internal/domain/user.go -destination=internal/mocks/user_mock.go -package=mocks

# Run database migrations
migrate-up:
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/clean_arch?sslmode=disable" up

# Rollback database migrations
migrate-down:
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/clean_arch?sslmode=disable" down

# Help command
help:
	@echo "make build - Build the application"
	@echo "make run - Run the application"
	@echo "make test - Run tests"
	@echo "make clean - Clean the binary"
	@echo "make deps - Download dependencies"
	@echo "make lint - Lint the code"
	@echo "make test-coverage - Run tests with coverage"
	@echo "make mocks - Generate mocks for testing"
	@echo "make migrate-up - Run database migrations"
	@echo "make migrate-down - Rollback database migrations"
