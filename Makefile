.PHONY: build run test clean docker-build docker-run lint

# Build the application
build:
	go build -o bin/genje-api cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Lint the code
lint:
	golangci-lint run

# Format the code
format:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Docker build
docker-build:
	docker build -t genje-api:latest .

# Docker run
docker-run:
	docker run -p 8080:8080 --env-file .env genje-api:latest

# Setup environment (copy config template)
setup-env:
	cp .env.example .env
	@echo "🟢 Created .env file from template. Please edit it with your configuration."

# Development setup
dev-setup: setup-env
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "🟢 Development environment ready!"
	@echo "   Edit .env file with your configuration, then run 'make run'"

# Database reset (development only)
db-reset:
	rm -f genje.db genje.db-wal genje.db-shm