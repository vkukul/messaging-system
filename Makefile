.PHONY: all build run clean db-setup db-migrate db-seed test swagger

# Go related variables
BINARY_NAME=messaging-system
MAIN_FILE=cmd/main.go

# Database related variables
DB_NAME=messaging_system
MIGRATIONS_DIR=pkg/database/migrations

# Default target
all: clean build

# Build the application
build: swagger
	@echo "Building..."
	@go build -o $(BINARY_NAME) $(MAIN_FILE)

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if ! command -v $(HOME)/go/bin/swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@$(HOME)/go/bin/swag init -g cmd/main.go

# Run the application
run: build
	@echo "Running..."
	@./$(BINARY_NAME)

# Clean build files
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

# Database setup
db-setup:
	@echo "Setting up database..."
	@if ! psql -lqt | cut -d \| -f 1 | grep -qw $(DB_NAME); then \
		createdb $(DB_NAME); \
		echo "Database $(DB_NAME) created."; \
	else \
		echo "Database $(DB_NAME) already exists."; \
	fi

# Apply database migrations
db-migrate: db-setup
	@echo "Applying migrations..."
	@psql -d $(DB_NAME) -f $(MIGRATIONS_DIR)/001_create_messages_table.sql

# Seed database with test data
db-seed: db-migrate
	@echo "Seeding database..."
	@psql -d $(DB_NAME) -f $(MIGRATIONS_DIR)/002_insert_test_data.sql

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Start PostgreSQL service
start-db:
	@echo "Starting PostgreSQL..."
	@brew services start postgresql@14

# Stop PostgreSQL service
stop-db:
	@echo "Stopping PostgreSQL..."
	@brew services stop postgresql@14

# Reset database (drop and recreate)
db-reset:
	@echo "Resetting database..."
	@dropdb --if-exists $(DB_NAME)
	@make db-setup db-migrate db-seed

# Development setup (database + build + run)
dev-setup: start-db db-reset build run

# Help
help:
	@echo "Available targets:"
	@echo "  make          : Build the application"
	@echo "  make build    : Build the application"
	@echo "  make run      : Run the application"
	@echo "  make clean    : Clean build files"
	@echo "  make test     : Run tests"
	@echo "  make db-setup : Create database"
	@echo "  make db-migrate : Apply migrations"
	@echo "  make db-seed  : Seed test data"
	@echo "  make start-db : Start PostgreSQL"
	@echo "  make stop-db  : Stop PostgreSQL"
	@echo "  make db-reset : Reset database"
	@echo "  make dev-setup: Complete development setup"
