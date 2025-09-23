APP_NAME := gophkeeper
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= gophkeeper
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_SSLMODE ?= disable
DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)


.PHONY: all
all: clean build-all

.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf build/
	@mkdir -p build

.PHONY: build
build: clean
	@echo "🔨 Building for current platform..."
	@go build  -o build/$(APP_NAME)-server ./cmd/server
	@go build -o build/$(APP_NAME)-client ./cmd/client
	@echo "✅ Build completed for current platform"

.PHONY: build-all
build-all: clean
	@echo "🌍 Building for all platforms..."
	@$(MAKE) build-linux
	@$(MAKE) build-windows
	@$(MAKE) build-darwin
	@echo "🎉 All builds completed!"

.PHONY: build-linux
build-linux:
	@echo "🐧 Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o build/$(APP_NAME)-server-linux-amd64 ./cmd/server
	@GOOS=linux GOARCH=amd64 go build -o build/$(APP_NAME)-client-linux-amd64 ./cmd/client
	@GOOS=linux GOARCH=arm64 go build -o build/$(APP_NAME)-server-linux-arm64 ./cmd/server
	@GOOS=linux GOARCH=arm64 go build -o build/$(APP_NAME)-client-linux-arm64 ./cmd/client
	@echo "✅ Linux builds completed"

.PHONY: build-windows
build-windows:
	@echo "🪟 Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o build/$(APP_NAME)-server-windows-amd64.exe ./cmd/server
	@GOOS=windows GOARCH=amd64 go build -o build/$(APP_NAME)-client-windows-amd64.exe ./cmd/client
	@GOOS=windows GOARCH=arm64 go build -o build/$(APP_NAME)-server-windows-arm64.exe ./cmd/server
	@GOOS=windows GOARCH=arm64 go build -o build/$(APP_NAME)-client-windows-arm64.exe ./cmd/client
	@echo "✅ Windows builds completed"

.PHONY: build-darwin
build-darwin:
	@echo "🍎 Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o build/$(APP_NAME)-server-darwin-amd64 ./cmd/server
	@GOOS=darwin GOARCH=amd64 go build -o build/$(APP_NAME)-client-darwin-amd64 ./cmd/client
	@GOOS=darwin GOARCH=arm64 go build -o build/$(APP_NAME)-server-darwin-arm64 ./cmd/server
	@GOOS=darwin GOARCH=arm64 go build -o build/$(APP_NAME)-client-darwin-arm64 ./cmd/client
	@echo "✅ macOS builds completed"

.PHONY: build-server
build-server: clean
	@echo "🔨 Building server for all platforms..."
	@GOOS=linux GOARCH=amd64 go build -o build/$(APP_NAME)-server-linux-amd64 ./cmd/server
	@GOOS=linux GOARCH=arm64 go build -o build/$(APP_NAME)-server-linux-arm64 ./cmd/server
	@GOOS=windows GOARCH=amd64 go build -o build/$(APP_NAME)-server-windows-amd64.exe ./cmd/server
	@GOOS=windows GOARCH=arm64 go build -o build/$(APP_NAME)-server-windows-arm64.exe ./cmd/server
	@GOOS=darwin GOARCH=amd64 go build -o build/$(APP_NAME)-server-darwin-amd64 ./cmd/server
	@GOOS=darwin GOARCH=arm64 go build -o build/$(APP_NAME)-server-darwin-arm64 ./cmd/server
	@echo "✅ Server builds completed"

.PHONY: build-client
build-client: clean
	@echo "🔨 Building client for all platforms..."
	@GOOS=linux GOARCH=amd64 go build -o build/$(APP_NAME)-client-linux-amd64 ./cmd/client
	@GOOS=linux GOARCH=arm64 go build -o build/$(APP_NAME)-client-linux-arm64 ./cmd/client
	@GOOS=windows GOARCH=amd64 go build -o build/$(APP_NAME)-client-windows-amd64.exe ./cmd/client
	@GOOS=windows GOARCH=arm64 go build -o build/$(APP_NAME)-client-windows-arm64.exe ./cmd/client
	@GOOS=darwin GOARCH=amd64 go build -o build/$(APP_NAME)-client-darwin-amd64 ./cmd/client
	@GOOS=darwin GOARCH=arm64 go build -o build/$(APP_NAME)-client-darwin-arm64 ./cmd/client
	@echo "✅ Client builds completed"

.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	@go vet ./...
	@golangci-lint run

.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	@migrate -path migrations -database "$(DSN)" up
	@echo "Migrations completed"

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	@migrate -path migrations -database "$(DSN)" down
	@echo "Rollback completed"

.PHONY: migrate-force
migrate-force:
	@echo "Force setting migration version..."
	@migrate -path migrations -database "$(DSN)" force $(VERSION)
	@echo "Migration version forced"

.PHONY: migrate-create
migrate-create:
	@echo "Creating new migration..."
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "Migration created"

.PHONY: env-setup
env-setup:
	@echo "Setting up environment..."
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "Created .env file from env.example"; \
		echo "Please edit .env file with your actual values"; \
	else \
		echo ".env file already exists"; \
	fi

.PHONY: load-env
load-env:
	@echo "Loading environment variables..."
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | xargs); \
		echo "Environment variables loaded"; \
	else \
		echo ".env file not found. Run 'make env-setup' first"; \
	fi