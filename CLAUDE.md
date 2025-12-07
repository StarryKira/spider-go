# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Spider-Go is a Go-based educational administration system crawler and management platform for CSUFT (Changsha University of Science and Technology). It provides REST APIs for querying grades, course schedules, exam schedules, and grade analysis.

## Development Commands

### Running the Application
```bash
# Run directly
go run main.go

# Build and run
go build -o spider-go
./spider-go

# Build for Windows
go build -o spider-go.exe
```

### Dependencies
```bash
# Install/update dependencies
go mod download

# Tidy dependencies
go mod tidy
```

### Database
The application uses GORM with auto-migration. Database tables are created automatically on startup. No migration commands needed.

## Architecture

### Dependency Injection Container
The entire application is built around a centralized dependency injection container (`internal/app/container.go`). **Never use global variables** - all dependencies flow through the container:

1. **Initialization order** (in `NewContainer`):
   - Config → DB → Redis → Repositories → Caches → Services → Middlewares → Controllers
   - RSA public key fetched on startup
   - Default admin created if not exists

2. **Adding new components**: Follow the existing pattern:
   - Add field to `Container` struct
   - Create initialization method (`initXxx`)
   - Call in proper order in `NewContainer`

### Layered Architecture

```
Routes (Gin) → Controllers → Services → Repositories/Caches/Crawler → DB/Redis/HTTP
```

- **Controllers** (`internal/controller`): HTTP request/response handling only
- **Services** (`internal/service`): All business logic lives here
- **Repositories** (`internal/repository`): Database access (GORM)
- **Caches** (`internal/cache`): Redis operations
- **DTOs** (`internal/dto`): Request/response data structures

### Configuration System

Configuration uses Viper with YAML (`config/config.yaml`). Two modes for educational system access:

- **campus**: For on-campus network (direct URLs to jwgl.csuft.edu.cn)
- **webvpn**: For off-campus access (WebVPN URLs)

Switch modes by setting `jwc.mode` in config. The container automatically injects the correct URLs into services.

### Session Management

Educational system sessions are cached in Redis (DB 0) with 1-hour expiration:
- `SessionService` handles login and cookie caching
- Retry logic: 3 attempts for login
- RSA encryption for passwords using public key from CAS server
- Session cookies are extracted and cached per user (uid)

### Caching Strategy

**Redis DB 0** (session Redis):
- User login sessions (1 hour expiration)
- DAU (Daily Active Users) statistics (30 days retention)
- System configuration (permanent)
- User data cache (grades, courses, exams)

**Redis DB 1** (captcha Redis):
- Email verification codes (5 minutes expiration)

### Authentication

**User JWT** (`middleware.AuthMiddleWare`):
- Uses JWT secret from config
- Automatically records DAU on each authenticated request
- Token contains uid, email

**Admin JWT** (`middleware.AdminAuthMiddleware`):
- Separate authentication from users
- Uses same JWT secret but different claims

### Scheduled Tasks

Defined in `internal/app/scheduler.go` using cron:
- **Daily 2 AM**: Data prewarming (pre-caches user data)
- **Every hour**: RSA public key refresh from CAS server

### Error Handling

Custom error system in `internal/common/errors.go`:
- `AppError` with error codes
- Standardized response format via `internal/common/response.go`
- Always use `NewAppError` for service-level errors

### Crawler Service

`CrawlerService` is a thin wrapper around HTTP client:
- Used by grade/course/exam services
- Takes cookies from `SessionService`
- Parses HTML with goquery

## Key Design Patterns

### Service Dependencies
Services are composed, not global:
```go
// ✓ Good: Dependencies injected
func NewGradeService(
    userRepo repository.UserRepository,
    sessionService service.SessionService,
    crawlerService service.CrawlerService,
    cache cache.UserDataCache,
    gradeURL string,
) GradeService

// ✗ Bad: Global variables
var globalDB *gorm.DB
```

### Mode-Aware URL Injection
Services receive URLs based on config mode:
```go
currentMode := c.Config.Jwc.GetCurrentModeConfig()
c.GradeService = service.NewGradeService(
    ...,
    currentMode.GradeURL,      // Campus or WebVPN URL
    currentMode.GradeLevelURL,
)
```

### Context Propagation
Always pass `context.Context` through service calls for cancellation and timeouts:
```go
func (s *gradeService) GetAllGrades(ctx context.Context, uid int) ([]Grade, error)
```

## Common Tasks

### Adding a New API Endpoint

1. **Define DTO** in `internal/dto/xxx_request.go`
2. **Add service method** in `internal/service/xxx_service.go`
3. **Add controller method** in `internal/controller/xxx_controller.go`
4. **Register route** in `internal/api/routes.go`
5. **Update container** if new service is needed

### Adding Redis Cache

1. Define interface in `internal/cache/xxx_cache.go`
2. Implement with Redis client
3. Add to container initialization
4. Inject into service that needs it

### Changing Educational System Mode

Edit `config/config.yaml`:
```yaml
jwc:
  mode: "webvpn"  # or "campus"
```
Restart application. No code changes needed.

## Important Notes

- **No tests exist yet** - the project currently lacks unit and integration tests
- **No Makefile** - use `go run`/`go build` directly
- **Auto-migration only** - GORM creates tables on startup, no manual migrations
- **Default admin**: `admin@spider-go.com` / `123456` (change immediately in production)
- **CORS**: Configured in `config.yaml` under `cors` section
- **Graceful shutdown**: Implemented in `main.go` with signal handling

## Database Schema

Tables auto-created by GORM:
- `users`: User accounts and bound educational system credentials
- `administrators`: Admin accounts (separate from users)
- `notices`: System notifications with display flags

## Configuration Checklist for Deployment

Before deploying, update `config/config.yaml`:
1. Change JWT secret to a strong random value
2. Update database credentials
3. Update Redis credentials
4. Configure email SMTP settings (for verification codes)
5. Set correct CORS allowed origins
6. Choose appropriate `jwc.mode` (campus vs webvpn)
7. Change default admin password after first login
