.PHONY: build build-seed run test clean docker-build docker-run migrate-up migrate-down seed

# Binary name
BINARY_NAME=config-server
SEED_BINARY=seed

# Build variables
BUILD_DIR=bin
MAIN_PATH=./cmd/server
SEED_PATH=./cmd/seed

# Docker variables
DOCKER_IMAGE=aami/config-server
DOCKER_TAG=latest

# Build the application
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build the seed command
build-seed:
	@echo "Building seed command..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(SEED_BINARY) $(SEED_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(SEED_BINARY)"

# Build all binaries
build-all: build build-seed

# Run the application
run: build
	@echo "Running..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run all tests
test:
	@echo "Running all tests..."
	@go test -v ./test/...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@go test -v ./test/unit/...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test -v ./test/integration/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./test/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 \
		-e DB_HOST=host.docker.internal \
		-e DB_PORT=5432 \
		-e DB_USER=postgres \
		-e DB_PASSWORD=postgres \
		-e DB_NAME=config_server \
		-e REDIS_HOST=host.docker.internal \
		-e REDIS_PORT=6379 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Run database migrations up
migrate-up:
	@echo "Running migrations up..."
	@psql -U postgres -d config_server -f migrations/001_initial_schema.sql
	@echo "Migrations complete"

# Run database migrations down
migrate-down:
	@echo "Rolling back migrations..."
	@psql -U postgres -d config_server -f migrations/001_initial_schema_down.sql
	@echo "Rollback complete"

# Seed default templates
seed: build-seed
	@echo "Seeding default templates..."
	@$(BUILD_DIR)/$(SEED_BINARY)
	@echo "Seed complete"

# Seed with force flag (overwrites existing builtin templates)
seed-force: build-seed
	@echo "Seeding default templates (force mode)..."
	@$(BUILD_DIR)/$(SEED_BINARY) --force
	@echo "Seed complete"

# Dry run seed (preview without changes)
seed-dry-run: build-seed
	@echo "Seeding default templates (dry run)..."
	@$(BUILD_DIR)/$(SEED_BINARY) --dry-run

# Development with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@air

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development dependencies installed"

# Help
help:
	@echo "Available targets:"
	@echo "  build             - Build the application"
	@echo "  build-seed        - Build the seed command"
	@echo "  build-all         - Build all binaries"
	@echo "  run               - Build and run the application"
	@echo "  test              - Run all tests"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration  - Run integration tests only"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  clean             - Clean build artifacts"
	@echo "  lint              - Run linter"
	@echo "  fmt               - Format code"
	@echo "  tidy              - Tidy dependencies"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-run        - Run Docker container"
	@echo "  migrate-up        - Run database migrations"
	@echo "  migrate-down      - Rollback database migrations"
	@echo "  seed              - Seed default templates"
	@echo "  seed-force        - Seed default templates (overwrite existing)"
	@echo "  seed-dry-run      - Preview seed changes without applying"
	@echo "  dev               - Start development server with hot reload"
	@echo "  dev-deps          - Install development dependencies"
	@echo "  help              - Show this help message"
