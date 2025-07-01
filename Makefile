.PHONY: build start stop restart logs clean dev test

# Default target
all: build

# Build the Docker images
build:
	docker compose build

# Start all services
start:
	docker compose up -d

# Start with logs (foreground)
dev:
	docker compose up

# Stop all services
stop:
	docker compose down

# Restart all services
restart: stop start

# View logs
logs:
	docker compose logs -f

# View specific service logs
logs-api:
	docker compose logs -f api

logs-fetcher:
	docker compose logs -f fetcher

logs-db:
	docker compose logs -f postgres

# Clean up everything (including volumes)
clean:
	docker compose down -v
	docker system prune -f

# Build and start in one command
run: build start

# Check service health
health:
	curl -s http://localhost:8080/health | jq '.'

# Test API endpoints
test-api:
	@echo "Testing API endpoints..."
	@echo "Health check:"
	@curl -s http://localhost:8080/health | jq '.'
	@echo "\nRecent articles:"
	@curl -s "http://localhost:8080/api/articles/recent?limit=5" | jq '.articles | length'
	@echo "\nNews sources:"
	@curl -s "http://localhost:8080/api/sources" | jq '.sources | length'

# Enter API container shell
shell-api:
	docker compose exec api bash

# Enter database shell
shell-db:
	docker compose exec postgres psql -U postgres -d genje

# Show running containers
status:
	docker compose ps

# Show container stats
stats:
	docker stats 