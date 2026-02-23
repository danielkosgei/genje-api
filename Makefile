APP_NAME := jalada
BUILD_DIR := bin

.PHONY: build run test test-coverage lint format tidy clean \
        migrate-up migrate-down migrate-create \
        docker-build docker-run docker-compose-up docker-compose-down

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

test:
	go test -race ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

format:
	gofmt -s -w .
	goimports -w .

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

migrate-up:
	migrate -path internal/database/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path internal/database/migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir internal/database/migrations -seq $$name

docker-build:
	docker build -t $(APP_NAME):latest .

docker-run: docker-build
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest

docker-compose-up:
	docker compose up -d

docker-compose-down:
	docker compose down
