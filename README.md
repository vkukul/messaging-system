# Messaging System

An automatic message sending system that processes unsent messages from the database every 2 minutes.

## Features

- Automatic message sending system
- REST API endpoints for control and monitoring
- PostgreSQL database integration
- Redis caching for message IDs (bonus feature)
- Swagger documentation
- Docker support

## Prerequisites

- Go 1.21 or higher
- PostgreSQL
- Redis (optional, for bonus feature)
- Docker (optional, for containerization)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd messaging-system
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up PostgreSQL database:
```sql
CREATE DATABASE messaging_system;
```

## Configuration

The application uses the following default configurations:

- PostgreSQL: `host=localhost user=postgres password=postgres dbname=messaging_system port=5432`
- Redis: `localhost:6379`
- Server: `:8080`

## Running the Application

### Local Development

```bash
go run cmd/main.go
```

### Using Docker

1. Build and run the containers:
```bash
# Build and start all services (app, PostgreSQL, Redis)
docker-compose up --build

# Run in detached mode
docker-compose up -d

# View logs
docker-compose logs -f
```

2. Stop the containers:
```bash
docker-compose down

# To remove volumes as well
docker-compose down -v
```

### Environment Variables

The application supports the following environment variables:

#### Database Configuration
- `DB_HOST` - PostgreSQL host (default: "localhost")
- `DB_USER` - PostgreSQL user (default: "postgres")
- `DB_PASSWORD` - PostgreSQL password (default: "postgres")
- `DB_NAME` - Database name (default: "messaging")
- `DB_PORT` - PostgreSQL port (default: "5432")

#### Redis Configuration
- `REDIS_HOST` - Redis host (default: "localhost")
- `REDIS_PORT` - Redis port (default: "6379")

These variables are automatically set when using Docker Compose.

## Quick Start

Using Make commands:

```bash
# Complete development setup (starts PostgreSQL, sets up database, builds and runs the application)
make dev-setup

# Or step by step:
make start-db    # Start PostgreSQL
make db-setup    # Create database
make db-migrate  # Apply migrations
make db-seed     # Add test data
make run         # Build and run the application
```

To see all available commands:
```bash
make help
```

## API Endpoints

- `POST /api/v1/messages/start` - Start automatic message processing
- `POST /api/v1/messages/stop` - Stop automatic message processing
- `GET /api/v1/messages/sent` - Get list of sent messages

## API Documentation

Swagger documentation is available at: `http://localhost:8080/swagger/index.html`

## Database Schema

The message table schema:

```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    to VARCHAR NOT NULL,
    content VARCHAR(160) NOT NULL,
    sent BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMP,
    message_id VARCHAR,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## System Architecture

### Components

1. **Web Server**
   - Built with Go and Gin framework
   - RESTful API endpoints
   - Swagger documentation
   - Rate limiting support

2. **Database (PostgreSQL)**
   - Stores message records
   - Tracks message status
   - Maintains message history

3. **Cache (Redis)**
   - Message ID caching
   - Rate limiting implementation
   - 24-hour cache expiration

### Message Processing

- Processes 2 messages every 2 minutes
- Implements rate limiting (10 messages per minute per recipient)
- Uses worker pool for parallel processing
- Retries failed operations with exponential backoff

### Error Handling

- Graceful degradation when Redis is unavailable
- Automatic retries for failed operations
- Comprehensive error logging
- Transaction support for database operations

## Testing

Run all tests:
```bash
go test ./... -v
```

Run specific package tests:
```bash
go test ./pkg/redis -v    # Test Redis package
go test ./internal/api -v # Test API handlers
```

### Test Coverage

Generate test coverage report:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Rate Limiting

The system implements rate limiting to prevent message flooding:

- Maximum 10 messages per minute per recipient
- Rate limits are tracked in Redis
- Rate limit keys expire after 1 minute
- Graceful handling when Redis is unavailable

## Monitoring

Monitor the application using Docker:

```bash
# View all container logs
docker compose logs -f

# View specific service logs
docker compose logs -f app
docker compose logs -f redis
docker compose logs -f postgres

# View container status
docker compose ps

# View container resources
docker stats
```

## Troubleshooting

1. **Database Connection Issues**
   - Check PostgreSQL container status
   - Verify database credentials
   - Ensure database migrations are applied

2. **Redis Connection Issues**
   - Check Redis container status
   - Verify Redis connection settings
   - Application will continue to work without Redis

3. **API Issues**
   - Check application logs
   - Verify correct endpoints and methods
   - Check rate limit status
