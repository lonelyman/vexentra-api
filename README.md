# Vexentra API

A modern RESTful API built with Go, Fiber, PostgreSQL, and Redis. This project demonstrates clean architecture patterns with proper separation of concerns, dependency injection, and containerized deployment.

## 🏗️ Architecture

The project follows a layered architecture pattern:

```
cmd/
  └── api/                    # Entry point
internal/
  ├── adapters/               # External service adapters
  │   └── database/postgres/  # Database implementations
  ├── config/                 # Configuration management
  ├── init/                   # Application initialization
  ├── modules/                # Business logic modules
  │   └── user/               # User module
  └── transport/http/         # HTTP handlers and middleware
pkg/
  ├── custom_errors/          # Custom error types
  └── logger/                 # Logging utilities
```

## 🛠️ Tech Stack

- **Framework**: [Fiber](https://gofiber.io/) - Fast and minimalist web framework
- **Database**: PostgreSQL 18 Alpine
- **Cache**: Redis 7.4 Alpine
- **ORM**: GORM
- **Go Version**: 1.25.0

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.25.0 (for local development)
- PostgreSQL (if running without Docker)
- Redis (if running without Docker)

## 🚀 Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone git@github.com:lonelyman/vexentra-api.git
cd vexentra-api

# Set up environment variables
cp .env.example .env  # if needed

# Build and start services
docker compose up --build

# API will be available at http://localhost:8000
```

### Local Development

```bash
# Install dependencies
go mod download

# Set environment variables
export APP_PORT=8000
export POSTGRES_PRIMARY_HOST=localhost
export POSTGRES_PRIMARY_PORT=5432
export POSTGRES_PRIMARY_USER=postgres
export POSTGRES_PRIMARY_PASSWORD=password
export POSTGRES_PRIMARY_NAME=vexentra_db
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Run the application
go run cmd/api/main.go
```

## 📦 Project Structure

### Modules

- **User Module** (`internal/modules/user/`)
   - `user_entity.go` - Domain entity
   - `user_repository.go` - Repository interface
   - `usersvc/user_service.go` - Service layer with business logic

### Adapters

- **PostgreSQL Adapter** (`internal/adapters/database/postgres/pguser/`)
   - `pg_user_model.go` - Database model
   - `pg_user_repository.go` - Repository implementation

### HTTP Layer

- **Router** (`internal/transport/http/router.go`) - Route definitions
- **Handlers** (`internal/transport/http/user/`) - HTTP request handlers
- **Middleware** (`internal/transport/http/middlewares/`) - Request/response middleware
- **Presenter** (`internal/transport/http/presenter/`) - Response formatting

### Configuration

- `internal/config/config.go` - Configuration loader with validation
- `internal/init/` - Application initialization modules
   - `app.go` - Main app setup
   - `db.go` - Database connection
   - `http.go` - HTTP server configuration
   - `redis.go` - Redis connection

## 🔧 Configuration

Environment variables are loaded from `.env` file:

```env
# App Configuration
APP_ENV=development
APP_PORT=8000
TIMEZONE=UTC

# PostgreSQL Configuration
POSTGRES_PRIMARY_HOST=vexentra-pgsql
POSTGRES_PRIMARY_PORT=5432
POSTGRES_PRIMARY_USER=postgres
POSTGRES_PRIMARY_PASSWORD=postgres
POSTGRES_PRIMARY_NAME=vexentra_db
POSTGRES_EXTERNAL_PORT=5432

# Redis Configuration
REDIS_HOST=vexentra-redis
REDIS_PORT=6379
REDIS_PASSWORD=redis
REDIS_EXTERNAL_PORT=6379
```

## 📝 API Endpoints

### User Endpoints

- `GET /api/users` - List all users
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create new user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

## 🧪 Error Handling

Custom error handling is implemented in `pkg/custom_errors/errors.go` to provide consistent API responses.

## 📊 Logging

Structured logging is available via middleware in `internal/transport/http/middlewares/logger.go` for request/response tracking.

## 🐳 Docker Compose Services

- **vexentra-pgsql**: PostgreSQL 18 Alpine database
- **vexentra-redis**: Redis 7.4 Alpine cache
- **api**: Go Fiber application

## 🔄 Development Workflow

1. Create a feature branch from `dev`
2. Make your changes
3. Test locally with `docker compose up --build`
4. Commit and push to your feature branch
5. Create a Pull Request to `dev` branch
6. After review and merge to `dev`, deploy to staging
7. Merge `dev` to `main` for production release

## 📚 Additional Resources

- [Fiber Documentation](https://docs.gofiber.io/)
- [GORM Guide](https://gorm.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)

## 📝 License

This project is private and owned by Vexentra.

## 👤 Author

Created with focus on clean code, scalability, and maintainability.
