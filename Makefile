.PHONY: build run test clean install-deps setup-auth

# Build the binary
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o cc-server ./cmd/server

# Build for current OS (development)
build-local:
	go build -o cc-server ./cmd/server

# Run the server locally
run:
	go run cmd/server/main.go

# Run with custom config
run-with-config:
	go run cmd/server/main.go --config ~/.config/cc/config.json

# Setup authentication (interactive)
setup-auth:
	@echo "Setting up authentication for Command Center v0.2.0"
	@read -p "Enter username: " username; \
	read -s -p "Enter password: " password; \
	echo ""; \
	go run cmd/server/main.go --username $$username --password $$password

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f cc-server
	rm -f cc.db cc.db-shm cc.db-wal
	rm -f command-center-*.tar.gz
	rm -rf ~/.config/cc/backups/

# Install Go dependencies
install-deps:
	go mod download
	go mod tidy

# Create release package
release: build
	tar -czf command-center-v0.2.0.tar.gz \
		cc-server \
		web/ \
		migrations/ \
		config.example.json \
		README.md

# Development - run with auto-reload (requires air)
dev:
	air

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run
