# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building and Testing
- `go build ./...` - Build all packages
- `go test ./... -cover -count=1` - Run all tests with coverage
- `make ci` - Run complete CI pipeline (lint, test, build)
- `make lint` - Run golangci-lint on all packages
- `make test` - Run tests with coverage
- `make build` - Build all packages

### Development Server
- `go run main.go start` - Start the CRM Lite API server (default port 8080)
- `air` - Start with hot-reload using Air (configured in .air.toml)

### Database Operations
- `go run main.go tools db:migrate` - Run database migrations
- `RUN_DB_TESTS=0 go run cmd/tools/db/migrate.go -env=dev -direction=up -steps=0` - Run dev migrations
- `RUN_DB_TESTS=0 go run cmd/tools/db/migrate.go -env=test -direction=up -steps=0` - Run test migrations

### API Documentation
- `make swag` or `swag init -g main.go -o docs` - Generate/update Swagger documentation
- Access interactive API docs at `http://localhost:8080/swagger/index.html` after starting server

### Docker Services
- `docker-compose up -d` - Start MariaDB, Redis, and phpMyAdmin services
- `docker-compose down` - Stop all services

## Architecture Overview

CRM Lite is a Go-based lightweight CRM system following domain-driven design principles with clear layered architecture.

### Core Technology Stack
- **Backend**: Go 1.24+, Gin web framework, Cobra CLI
- **Database**: GORM ORM with MariaDB/MySQL, Redis for caching
- **Security**: JWT authentication, Casbin RBAC authorization
- **Documentation**: Swagger/OpenAPI auto-generated from code annotations
- **Container**: Docker and Docker Compose ready

### Directory Structure
The application follows a layered architecture pattern:

```
internal/
├── bootstrap/     # Application initialization and resource setup
├── controller/    # HTTP request handlers (Gin controllers)
├── core/          # Core components: config, logging, resource management
├── dao/           # Data Access Objects (GORM generated code)
├── dto/           # Data Transfer Objects for API input/output
├── middleware/    # Gin middlewares (auth, logging, CORS)
├── domains/       # Domain-driven design entities and business logic
├── service/       # Business logic layer
├── routes/        # API route registration
└── policy/        # Authorization policies and Casbin rules
```

### Key Architectural Patterns
- **Resource Manager Pattern**: Central `resource.Manager` handles DB, Redis, logger, config
- **Domain-Driven Design**: Business logic organized in domain modules (customer, order, wallet, marketing)
- **Controller-Service-DAO**: Clear separation of concerns across layers
- **DTO Pattern**: API contracts isolated from internal data models
- **Configuration-Driven**: Environment-specific YAML configs in `config/` directory

### Module Organization
Core business modules include:
- **Customer Management**: Full CRUD operations for client data
- **Order Management**: Transaction processing with order items
- **Product Management**: Inventory and product catalog
- **Wallet System**: User balance and transaction history
- **Marketing**: Campaign management and tracking
- **User & Role Management**: RBAC with JWT authentication

### Database and Migrations
- Uses GORM Gen for type-safe query generation
- SQL migrations located in `db/migrations/`
- Supports multiple environments (dev, test, prod)
- Migration tool supports up/down migrations with step control

### Configuration Management
- Environment-specific configs: `config/app.{env}.yaml`
- Environment variables via `.env` file
- Auto-detection of config files based on ENV variable
- Supports container deployments with `/app/config/` path

### Development Standards (from .cursor/rules)
- CamelCase for Go variables and functions
- snake_case for database fields
- RESTful API design (GET /customers, POST /orders)
- Controllers must end with `Controller` suffix
- Use `*gin.Context` parameter named `c` consistently
- Single responsibility functions with clear naming

## Key Files to Understand
- `main.go` - Application entry point with Swagger annotations
- `cmd/root.go` - Cobra CLI configuration and initialization
- `cmd/start.go` - Server startup command
- `internal/bootstrap/` - Resource initialization
- `internal/routes/` - API route definitions
- `internal/core/config/` - Configuration management
- `docs/architecture/README.md` - Detailed architecture documentation

## Testing Strategy
- Unit tests throughout codebase with `_test.go` files
- Test environment with separate database configuration
- Coverage reporting via `go test -cover`
- Integration tests can be run with `RUN_DB_TESTS=1`