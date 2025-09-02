# Makefile for Agro Mas Backend
.PHONY: help build run clean test deps docker-build docker-run docker-stop migrate-up migrate-down lint format

# Variables
BINARY_NAME=agro-mas-backend
MAIN_PATH=./cmd/api
DOCKER_IMAGE=agro-mas-backend
DOCKER_TAG=latest

# Colors for terminal output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

# Default target
help: ## Show this help message
	@echo "$(GREEN)Agro Mas Backend - Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(YELLOW)<target>$(NC)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Development commands
deps: ## Install Go dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	go mod tidy
	go mod download

build: ## Build the application
	@echo "$(GREEN)Building application...$(NC)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) $(MAIN_PATH)

build-local: ## Build for local development
	@echo "$(GREEN)Building for local development...$(NC)"
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application locally
	@echo "$(GREEN)Running application...$(NC)"
	go run $(MAIN_PATH)

run-dev: ## Run with live reload (requires air)
	@echo "$(GREEN)Running with live reload...$(NC)"
	air

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf bin/
	go clean

# Testing
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem ./...

# Code quality
lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

format: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	go vet ./...

# Database operations
migrate-up: ## Run database migrations up
	@echo "$(GREEN)Running migrations up...$(NC)"
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/agro_mas_dev?sslmode=disable" up

migrate-down: ## Run database migrations down
	@echo "$(GREEN)Running migrations down...$(NC)"
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/agro_mas_dev?sslmode=disable" down

migrate-force: ## Force migration version (use with VERSION=<version>)
	@echo "$(GREEN)Forcing migration to version $(VERSION)...$(NC)"
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/agro_mas_dev?sslmode=disable" force $(VERSION)

migrate-create: ## Create new migration (use with NAME=<migration_name>)
	@echo "$(GREEN)Creating migration $(NAME)...$(NC)"
	migrate create -ext sql -dir migrations -seq $(NAME)

# Docker commands
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -f deployments/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop: ## Stop Docker containers
	@echo "$(GREEN)Stopping Docker containers...$(NC)"
	docker stop $$(docker ps -q --filter ancestor=$(DOCKER_IMAGE):$(DOCKER_TAG))

# Docker Compose commands
compose-up: ## Start all services with Docker Compose
	@echo "$(GREEN)Starting services with Docker Compose...$(NC)"
	docker-compose -f deployments/docker-compose.yml up -d

compose-down: ## Stop all services with Docker Compose
	@echo "$(GREEN)Stopping services with Docker Compose...$(NC)"
	docker-compose -f deployments/docker-compose.yml down

compose-logs: ## Show logs from Docker Compose services
	@echo "$(GREEN)Showing Docker Compose logs...$(NC)"
	docker-compose -f deployments/docker-compose.yml logs -f

compose-restart: ## Restart Docker Compose services
	@echo "$(GREEN)Restarting Docker Compose services...$(NC)"
	docker-compose -f deployments/docker-compose.yml restart

# Development environment setup
setup-dev: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)Creating .env file from template...$(NC)"; \
		cp .env.example .env; \
		echo "$(YELLOW)Please edit .env file with your configuration$(NC)"; \
	fi
	@echo "$(GREEN)Installing development tools...$(NC)"
	@if ! command -v air >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing air for live reload...$(NC)"; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing migrate tool...$(NC)"; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	fi
	@echo "$(GREEN)Development environment setup complete!$(NC)"

# Database seeding
seed: ## Seed database with sample data
	@echo "$(GREEN)Seeding database...$(NC)"
	go run scripts/seed.go

seed-users: ## Seed database with sample users
	@echo "$(GREEN)Seeding users...$(NC)"
	go run scripts/seed_users.go

seed-products: ## Seed database with sample products
	@echo "$(GREEN)Seeding products...$(NC)"
	go run scripts/seed_products.go

# Production deployment
deploy-staging: ## Deploy to staging environment
	@echo "$(GREEN)Deploying to staging...$(NC)"
	gcloud builds submit --config deployments/cloudbuild.yaml --substitutions BRANCH_NAME=staging .

deploy-prod: ## Deploy to production environment
	@echo "$(GREEN)Deploying to production...$(NC)"
	gcloud builds submit --config deployments/cloudbuild.yaml --substitutions BRANCH_NAME=main .

# Monitoring and debugging
logs: ## Show application logs (local)
	@echo "$(GREEN)Showing application logs...$(NC)"
	tail -f logs/app.log

logs-cloud: ## Show Cloud Run logs
	@echo "$(GREEN)Showing Cloud Run logs...$(NC)"
	gcloud logging read "resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"agro-mas-backend\"" --limit 50 --format json

health: ## Check application health
	@echo "$(GREEN)Checking application health...$(NC)"
	curl -f http://localhost:8080/health || echo "$(RED)Health check failed$(NC)"

# Security
security-scan: ## Run security scan
	@echo "$(GREEN)Running security scan...$(NC)"
	gosec ./...

vuln-check: ## Check for vulnerabilities
	@echo "$(GREEN)Checking for vulnerabilities...$(NC)"
	govulncheck ./...

# Performance
profile: ## Run performance profiling
	@echo "$(GREEN)Running performance profiling...$(NC)"
	go run $(MAIN_PATH) -cpuprofile=cpu.prof -memprofile=mem.prof

load-test: ## Run load tests (requires hey)
	@echo "$(GREEN)Running load tests...$(NC)"
	hey -n 1000 -c 10 http://localhost:8080/health

# Documentation
docs: ## Generate API documentation
	@echo "$(GREEN)Generating API documentation...$(NC)"
	swag init -g $(MAIN_PATH)/main.go

# Git hooks
pre-commit: ## Run pre-commit checks
	@echo "$(GREEN)Running pre-commit checks...$(NC)"
	$(MAKE) format
	$(MAKE) lint
	$(MAKE) test

# Quick commands for common workflows
dev: compose-up run-dev ## Start development environment and run with live reload

full-test: lint vet test ## Run complete test suite

quick-deploy: build docker-build ## Quick build and Docker image creation

# Environment specific runs
run-staging: ## Run with staging configuration
	@echo "$(GREEN)Running with staging configuration...$(NC)"
	ENVIRONMENT=staging go run $(MAIN_PATH)

run-prod: ## Run with production configuration
	@echo "$(GREEN)Running with production configuration...$(NC)"
	ENVIRONMENT=production GIN_MODE=release go run $(MAIN_PATH)

# Cleanup commands
clean-all: clean ## Clean everything including Docker images and volumes
	@echo "$(GREEN)Cleaning Docker images and volumes...$(NC)"
	docker system prune -f
	docker volume prune -f

reset-db: ## Reset database (drops and recreates)
	@echo "$(RED)Resetting database...$(NC)"
	docker-compose -f deployments/docker-compose.yml down -v
	docker-compose -f deployments/docker-compose.yml up -d postgres
	sleep 5
	$(MAKE) migrate-up
	$(MAKE) seed

# Version and release
version: ## Show current version
	@echo "$(GREEN)Current version:$(NC)"
	@git describe --tags --always --dirty

tag: ## Create a new tag (use with VERSION=v1.0.0)
	@echo "$(GREEN)Creating tag $(VERSION)...$(NC)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Status and info
status: ## Show status of services
	@echo "$(GREEN)Service Status:$(NC)"
	@echo "Application: $$(curl -s http://localhost:8080/health | grep -o '\"status\":\"[^\"]*\"' || echo 'Not running')"
	@echo "Database: $$(docker-compose -f deployments/docker-compose.yml ps postgres | grep -o 'Up.*' || echo 'Not running')"
	@echo "Redis: $$(docker-compose -f deployments/docker-compose.yml ps redis | grep -o 'Up.*' || echo 'Not running')"

info: ## Show project information
	@echo "$(GREEN)Project Information:$(NC)"
	@echo "Name: Agro Mas Backend"
	@echo "Language: Go"
	@echo "Framework: Gin"
	@echo "Database: PostgreSQL with PostGIS"
	@echo "Cloud: Google Cloud Platform"
	@echo "Container: Docker"
	@echo "Binary: $(BINARY_NAME)"
	@echo "Main Path: $(MAIN_PATH)"