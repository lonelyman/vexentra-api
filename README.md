# Vexentra API — Master Template

Go REST API template ที่ออกแบบตาม Clean Architecture, BigTech standards, และ production-readiness ตั้งแต่ต้น

---

## Tech Stack

| Layer      | Technology                                         |
| ---------- | -------------------------------------------------- |
| Framework  | Fiber v3                                           |
| Database   | PostgreSQL + GORM                                  |
| Cache      | Redis                                              |
| Auth       | JWT (HS256) — Access + Refresh Token               |
| Logger     | slog (JSON production / PrettyHandler development) |
| Validation | go-playground/validator v10                        |
| Container  | Docker + Docker Compose                            |

---

## Project Structure

```
cmd/api/                          # Entrypoint
internal/
  bootstrap/                      # App wiring: DI, Fiber, DB, Redis
  config/                         # Config struct + env loader (Fail-Fast)
  adapters/database/postgres/     # DB model + repository implementations
  modules/user/                   # Domain entity + repository interface
    usersvc/                      # Business logic (UserService)
  transport/http/
    health/                       # /health/live + /health/ready
    middlewares/                  # AuthMiddleware, StructuredLogger
    presenter/                    # Response envelope (RenderItem/RenderList/RenderError)
    user/                         # UserHandler, Request, Response
pkg/
  auth/                           # JWT AuthService, Claims, GetClaims helper
  custom_errors/                  # AppError + standard error codes
  logger/                         # Logger interface + PrettyHandler + JSONHandler
```

---

## API Endpoints

| Method | Path                     | Auth   | Description                                |
| ------ | ------------------------ | ------ | ------------------------------------------ |
| `GET`  | `/health/live`           | —      | Liveness probe (process alive)             |
| `GET`  | `/health/ready`          | —      | Readiness probe (DB + Redis ping)          |
| `POST` | `/api/v1/users/register` | —      | Register + auto-login (returns token pair) |
| `GET`  | `/api/v1/me`             | Bearer | Get own profile (stub)                     |

---

## Quick Start

```bash
# 1. Clone and enter
git clone <repo> && cd vexentra-api

# 2. Copy env
cp .env .env.local   # แก้ค่าให้ตรงกับ local ของคุณ

# 3. Start infrastructure
docker compose up -d

# 4. Run API (with air for hot reload)
air
```

---

## Environment Variables

```env
# App
API_ENV=development
API_PORT=3000
API_TIMEZONE=Asia/Bangkok
API_CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# PostgreSQL
POSTGRES_PRIMARY_HOST=vexentra-pgsql
POSTGRES_PRIMARY_PORT=5432
POSTGRES_PRIMARY_USER=vexentra_user
POSTGRES_PRIMARY_PASSWORD=vexentra_pass
POSTGRES_PRIMARY_NAME=vexentra_db
POSTGRES_PRIMARY_SSL_MODE=disable

# Redis
REDIS_HOST=vexentra-redis
REDIS_PORT=6379
REDIS_PASSWORD=vexentra_redis_pass
REDIS_DB=0

# JWT
JWT_ACCESS_SECRET=change-me-in-production
JWT_ACCESS_EXPIRY=60m
JWT_REFRESH_SECRET=change-me-refresh-in-production
JWT_REFRESH_EXPIRY=336h
JWT_ISSUER=vexentra-api
```

---

## Architecture Decisions

### Clean Architecture Layers

```
Transport (HTTP) → Service (Business Logic) → Repository (Interface) → Adapter (DB/Redis)
```

- **Domain entity** (`modules/user/user_entity.go`) ไม่มี JSON tags — transport layer จัดการ serialization เพียงที่เดียว
- **Repository interface** อยู่ใน domain module ไม่ใช่ใน adapter — dependency inversion

### JWT Design (RFC 8725 compliant)

- `AccessClaims` และ `RefreshClaims` เป็น struct แยก → ป้องกัน Token Confusion Attack
- `jti` (JWT ID) ทุก token → รองรับ Token Blacklist ในอนาคต
- `sub` = userID ตาม RFC 7519 standard
- `DeviceID` ใน RefreshClaims → รองรับ per-device logout
- Refresh Secret แยกจาก Access Secret

### Error Handling

- `AppError` struct พร้อม `HTTPStatus`, `Code`, `Message`, `Details`
- Global Error Handler ใน Fiber — handler ไม่ต้อง handle transport error เอง
- Repository แปลง `gorm.ErrRecordNotFound` → `AppError` ก่อนส่งขึ้น

### Health Check (Kubernetes Ready)

- `/health/live` — ไม่แตะ DB/Redis → Kubernetes `livenessProbe`
- `/health/ready` — ping DB + Redis + per-component status + HTTP 503 เมื่อ unhealthy → Kubernetes `readinessProbe`

---

## Current Status — Review Session (2026-04-08)

### ✅ Completed

| #   | รายการ                                                                                       |
| --- | -------------------------------------------------------------------------------------------- |
| 1   | `JWTConfig` struct (Access + Refresh + Issuer) ใน config                                     |
| 2   | `auth_service.go` — TokenPair, AccessClaims/RefreshClaims, jti, DeviceID, BigTech grade      |
| 3   | Rename `internal/init` → `internal/bootstrap`                                                |
| 4   | `UserRepository` — เพิ่ม `GetByEmail`, fix `GetByID` type `string` → `uint`                  |
| 5   | `UserService` — duplicate email check, inject AuthService, return RegisterResult + TokenPair |
| 6   | Register `StructuredLogger`, CORS อ่านจาก config แยกตาม env                                  |
| 7   | Domain entity ลบ JSON tags ออก                                                               |
| 8   | `PrettyHandler.WithAttrs` / `WithGroup` fix — ไม่ทิ้ง pre-set attributes แล้ว                |
| 9   | Health Check — `/health/live` + `/health/ready` (Kubernetes standard)                        |
| 10  | `AuthService` inject ครบ — ไม่มี nil pointer panic ใน protected routes แล้ว                  |
| 11  | ลบ dead code `registerRoutes()`                                                              |
| 12  | `auth.GetClaims(c)` helper — type-safe claims extraction                                     |
| R1  | `GetProfile` — ใช้ `auth.GetClaims(c)` + `svc.GetProfile()` จริง ไม่ใช่ stub แล้ว            |
| R2  | `GetByUsername` — duplicate username check ก่อน insert                                       |
| R3  | `bootstrap/redis.go` — comment อัพเดตจากชื่อ package เก่า `init`                             |
| R4  | `custom_errors/errors.go` — ลบ Order domain error codes ออก (domain leakage)                 |
| R5  | `godotenv` — auto-load `.env` ใน non-production, skip ใน production                          |
| R6  | `gormLoggerAdapter.LogMode` — เก็บ level ใน struct, filter ก่อน log ทุก method               |
| R7  | Soft delete — `gorm.DeletedAt` ใน DB model, `*time.Time` ใน domain entity                    |

---

### 🟡 Minor / Future Improvements

| #   | รายการ                                                                                                                 |
| --- | ---------------------------------------------------------------------------------------------------------------------- |
| M1  | `User.ID` เป็น `uint` แต่ JWT `sub` แปลงด้วย `fmt.Sprint` → ควร migrate เป็น `uuid.UUID` สำหรับ security + scalability |
| M2  | Password validation เฉพาะ `min=8` — ควรเพิ่ม complexity rule                                                           |
| M3  | ไม่มี `.env.example` file — สำคัญสำหรับ template                                                                       |
| M4  | ยังไม่มี Login endpoint (`POST /api/v1/auth/login`)                                                                    |
| M5  | ยังไม่มี Refresh Token endpoint (`POST /api/v1/auth/refresh`)                                                          |
| M6  | ยังไม่มี Logout endpoint (`POST /api/v1/auth/logout`)                                                                  |

## 🏗️ Architecture

The project follows a layered architecture pattern:

```
cmd/
  └── api/                    # Entry point
internal/
  ├── adapters/               # External service adapters
  │   └── database/postgres/  # Database implementations
  ├── config/                 # Configuration management
  ├── bootstrap/              # Application initialization (DI wiring)
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
