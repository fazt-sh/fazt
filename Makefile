.PHONY: build run test clean install-deps

# Build the binary
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -o cc-server ./cmd/server

# Build for current OS (development)
build-local:
	go build -o cc-server ./cmd/server

# Run the server locally
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f cc-server
	rm -f cc.db cc.db-shm cc.db-wal
	rm -f command-center-*.tar.gz

# Install Go dependencies
install-deps:
	go mod download
	go mod tidy

# Create release package
release: build
	tar -czf command-center-v0.1.0.tar.gz \
		cc-server \
		web/ \
		migrations/ \
		.env.example \
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
