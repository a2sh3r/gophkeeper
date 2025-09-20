APP_NAME := gophkeeper
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")


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