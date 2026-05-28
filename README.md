# Gin Web API Template

This is a modern and modular Web API project template built with Go and the Gin framework. The project utilizes dependency injection for component management and integrates common backend infrastructure.

## Features

- **Web Framework**: Gin
- **Dependency Injection**: samber/do
- **Configuration Management**: TOML files and environment variables
- **Database ORM**: GORM
- **Cache**: Redis
- **Object Storage**: MinIO Go SDK
- **Message Queue**: NATS
- **Access Control (RBAC)**: Custom implementation based on gorbac/v3 with database persistence
- **Authentication**: JWT Token

## Directory Structure

```text
├── cmd
│   └── api/             # Application entry point
├── config/              # Configuration logic
├── internal
│   ├── api
│   │   └── handler/     # HTTP handlers
│   ├── di/              # Dependency injection container and providers
│   └── service/         # Core business logic
├── pkg/                 # Reusable libraries
├── Makefile             # Build commands
├── go.mod               # Go dependencies
└── README.md            # Project documentation
```

## Requirements

- Go 1.25.0+
- PostgreSQL or MySQL
- Redis
- S3 compatible object storage
- NATS Server

## Configuration

The project supports configuration via TOML files or environment variables. The main configuration items are as follows:

| Module | Environment Variables | Default Value |
| :--- | :--- | :--- |
| **App** | APP_HOST / APP_JWT_SECRET | localhost:8080 / secret |
| **Database** | DATABASE_DB_TYPE / DATABASE_DSN | postgres / Default PostgreSQL DSN |
| **Redis** | REDIS_HOST / REDIS_PASSWORD / REDIS_DB | localhost:6379 |
| **S3** | S3_ENDPOINT / S3_ACCESS_KEY_ID / S3_SECRET_ACCESS_KEY / S3_BUCKET / S3_REGION | s3.amazonaws.com |
| **NATS** | NATS_URL | nats://localhost:4222 |

## Dependency Injection

This project uses github.com/samber/do for dependency injection. Various external components and internal services are registered centrally in internal/di/container.go:

1. **Infrastructure**: ProvideRedis, ProvideDB, ProvideRBAC, ProvideS3, ProvideNats
2. **Services**: service.NewAuthService, service.NewRBACService
3. **Handlers**: handler.NewAuthHandler

## Build and Run

You can build the project quickly using the Makefile:

```bash
make all

./build/api
```

Alternatively, you can run the service directly with Go:

```bash
go run ./cmd/api
```
