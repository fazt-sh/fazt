.PHONY: build run test clean install-deps setup-auth help

# Version injection
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -w -s -X main.Version=$(VERSION)

# Build the binary (release)
# Enforce CGO_ENABLED=0 as per GEMINI.md
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o fazt ./cmd/server

# Build for current OS (development)
build-local:
	CGO_ENABLED=0 go build -ldflags="-X main.Version=$(VERSION)" -o fazt ./cmd/server

# Run the server locally
run: build-local
	./fazt server start

# Run with custom config
run-with-config:
	go run cmd/server/main.go server start --config ~/.config/fazt/config.json

# Setup authentication (interactive)
setup-auth:
	@echo "Setting up authentication for fazt.sh v0.3.0"
	@read -p "Enter username: " username; \
	read -s -p "Enter password: " password; \
	echo ""; \
	go run cmd/server/main.go server set-credentials --username $$username --password $$password

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test ./... -cover

# Clean build artifacts
clean:
	rm -f fazt
	rm -f cc.db cc.db-shm cc.db-wal
	rm -f fazt-*.tar.gz
	rm -rf ~/.config/fazt/backups/

# Install Go dependencies
install-deps:
	go mod download
	go mod tidy

# Create release package
release: build
	tar -czf fazt-v0.3.0.tar.gz \
		fazt \
		web/ \
		migrations/ \
		examples/ \
		config.example.json \
		README.md \
		CLAUDE.md

# Development - run with auto-reload (requires air)
dev:
	air

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "fazt.sh v0.3.0 - Makefile Targets"
	@echo ""
	@echo "  make build       - Build release binary (linux/amd64)"
	@echo "  make build-local - Build for current OS"
	@echo "  make run         - Build and run server"
	@echo "  make test        - Run all tests"
	@echo "  make test-cover  - Run tests with coverage"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make setup-auth  - Setup authentication"
	@echo "  make release     - Create release tarball"
	@echo "  make help        - Show this help"
