# Desa API - Village Information System

A REST API for managing village (desa) information with role-based access control, built following Clean Architecture principles.

## Features

- JWT-based authentication
- RBAC authorization with Casbin
- User and role management
- Village profile management (desa)
- Organization structure (struktur)
- News articles (berita)
- Banner management
- Small business directory (UMKM)
- Facilities mapping (fasilitas)
- Public information documents (PPID)
- Database backup and restore
- File upload handling
- Rate limiting
- CORS support

## Technology Stack

- **Language**: Go 1.24.1
- **HTTP Router**: Chi v5.2.5
- **Database**: PostgreSQL with sqlx
- **Authorization**: Casbin v2.135.0
- **Authentication**: JWT (golang-jwt/jwt v5.3.1)
- **Migrations**: Goose v3.26.0
- **Observability**: OpenTelemetry
- **Testing**: testify v1.11.1

## Installation and Setup

### Prerequisites

- Go 1.24.1 or higher
- PostgreSQL 14 or higher

### Installation Steps

1. Clone the repository:
```bash
git clone <repository-url>
cd webdesa/api
```

2. Install dependencies:
```bash
go mod download
```

3. Copy and configure the application:
```bash
cp config.example.toml config.toml
```

4. Edit `config.toml` with your database credentials and settings

5. Run database migrations:
```bash
go run main.go migrate up
```

6. (Optional) Seed initial data:
```bash
go run main.go seed
```

7. Start the server:
```bash
go run main.go serve
```

The API will be available at `http://localhost:8080` (or the address specified in your config).

## Configuration

Configuration files are located in the `config/` directory. The application looks for `config.toml` in the current directory or home directory.

### Configuration Options

**Application Settings** (`config/app.go`):
```toml
[app]
name = "desa-api"
version = "1.0.0"
http_addr = ":8080"
timezone = "Asia/Jakarta"
read_timeout = "15s"
write_timeout = "15s"
idle_timeout = "60s"
proxy_url = ""  # Optional HTTP proxy
```

**Database Settings** (`config/postgres.go`):
```toml
[postgres]
dsn = "postgres://user:password@localhost:5432/dbname?sslmode=disable"
max_open_conns = 25
max_idle_conns = 5
conn_max_lifetime = "5m"
```

**JWT Settings** (`config/jwt.go`):
```toml
[jwt]
secret = "your-secret-key-change-this"
expiration = "24h"
```

**Backup Settings** (`config/backup.go`):
```toml
[backup]
directory = "./backups"
```

**CORS Settings** (`config/cors.go`):
```toml
[cors]
allowed_origins = ["http://localhost:3000"]
```

**Rate Limiting** (`config/ratelimit.go`):
```toml
[ratelimit]
auth_requests_per_minute = 5
public_requests_per_minute = 100
protected_requests_per_minute = 30
```

**File Upload** (`config/fileupload.go`):
```toml
[fileupload]
max_image_size_mb = 10
max_document_size_mb = 50
upload_directory = "./uploads"
```

**Observability** (`config/otel.go`):
```toml
[otel]
metric = true
trace = true
log = true
```

See `config/dev.toml`, `config/staging.toml`, and `config/prod.toml` for environment-specific examples.

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login with email and password
- `POST /api/auth/register` - Register new user
- `POST /api/auth/refresh` - Refresh JWT token

### Users
- `GET /api/users` - List users (admin only)
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create user (admin only)
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user (admin only)

### Roles
- `GET /api/roles` - List all roles
- `POST /api/roles` - Create role (admin only)
- `GET /api/roles/:id` - Get role by ID
- `PUT /api/roles/:id` - Update role (admin only)
- `DELETE /api/roles/:id` - Delete role (admin only)

### Desa (Village Profile)
- `GET /api/desa` - Get village profile
- `PUT /api/desa` - Update village profile

### Struktur (Organization Structure)
- `GET /api/struktur` - List organization members
- `POST /api/struktur` - Create member
- `GET /api/struktur/:id` - Get member by ID
- `PUT /api/struktur/:id` - Update member
- `DELETE /api/struktur/:id` - Delete member

### Berita (News)
- `GET /api/berita` - List news articles (with search and date filters)
- `POST /api/berita` - Create news article
- `GET /api/berita/:id` - Get article by ID
- `PUT /api/berita/:id` - Update article
- `DELETE /api/berita/:id` - Delete article
- `POST /api/berita/upload-media` - Upload media for Quill editor

### Berita Categories
Categories are a first-class, managed vocabulary (no free text). Create + delete only.

- `GET /api/public/berita/categories` - List categories (paginated, supports `q` for autocomplete, public)
- `POST /api/berita/categories` - Create category (auth, `berita:write`)
- `DELETE /api/berita/categories/:id` - Delete category (auth, `berita:write`; returns 409 if in use)

When creating or updating a berita, the `category` field must be the **UUID** of an existing category.

### Banner
- `GET /api/banners` - List banners
- `POST /api/banners` - Create banner
- `GET /api/banners/:id` - Get banner by ID
- `PUT /api/banners/:id` - Update banner
- `DELETE /api/banners/:id` - Delete banner

### UMKM (Small Business)
- `GET /api/umkm` - List businesses (with search)
- `POST /api/umkm` - Create business
- `GET /api/umkm/:id` - Get business by ID
- `PUT /api/umkm/:id` - Update business
- `DELETE /api/umkm/:id` - Delete business

### UMKM Categories
Categories are a first-class, managed vocabulary (no free text). Create + delete only.

- `GET /api/public/umkm/categories` - List categories (paginated, supports `q` for autocomplete, public)
- `POST /api/umkm/categories` - Create category (auth, `umkm:write`)
- `DELETE /api/umkm/categories/:id` - Delete category (auth, `umkm:write`; returns 409 if in use)

When creating or updating a UMKM, the `category` field must be the **UUID** of an existing category.

### PPID Categories
Categories are a first-class, managed vocabulary. PPID `category` is **optional** (nullable FK).

- `GET /api/public/ppid/categories` - List categories (paginated, supports `q` for autocomplete, public)
- `POST /api/ppid/categories` - Create category (auth, `ppid:write`)
- `DELETE /api/ppid/categories/:id` - Delete category (auth, `ppid:write`; returns 409 if in use)

### Fasilitas Categories
Categories are a first-class, managed vocabulary. Fasilitas `category` (formerly `type`) is **optional** (nullable FK).

- `GET /api/public/fasilitas/categories` - List categories (paginated, supports `q` for autocomplete, public)
- `POST /api/fasilitas/categories` - Create category (auth, `fasilitas:write`)
- `DELETE /api/fasilitas/categories/:id` - Delete category (auth, `fasilitas:write`; returns 409 if in use)

When creating or updating a fasilitas, the `category` field must be the **UUID** of an existing category (was `type`, now renamed for consistency).

### Infographic Categories
Categories are a first-class, managed vocabulary. Infographic `category` is **required** (NOT NULL FK, default `"Lainnya"`).

- `GET /api/public/infographic/categories` - List categories (paginated, supports `q` for autocomplete, public)
- `POST /api/infographic/categories` - Create category (auth, `infographic:write`)
- `DELETE /api/infographic/categories/:id` - Delete category (auth, `infographic:write`; returns 409 if in use)

### Fasilitas (Facilities)
- `GET /api/fasilitas` - List facilities (with bounding box filter)
- `POST /api/fasilitas` - Create facility
- `GET /api/fasilitas/:id` - Get facility by ID
- `PUT /api/fasilitas/:id` - Update facility
- `DELETE /api/fasilitas/:id` - Delete facility

### PPID (Public Information)
- `GET /api/ppid` - List documents (with category and year filters)
- `POST /api/ppid` - Create document
- `GET /api/ppid/:id` - Get document by ID
- `PUT /api/ppid/:id` - Update document
- `DELETE /api/ppid/:id` - Delete document

### Backup
- `POST /api/backup` - Create database backup (admin only)
- `POST /api/restore` - Restore from backup (admin only)

### Health
- `GET /health` - Health check endpoint

All protected endpoints require `Authorization: Bearer <token>` header.

## CLI Commands

The application provides several CLI commands:

### serve
Start the HTTP server:
```bash
go run main.go serve
```

Options:
- `--config` - Specify config file path
- `--log-level` - Set log level (debug, info, warn, error)
- `--no-proxy` - Disable proxy
- `--run-profile` - Enable profiling (cpu, ram, mutex, block, goroutine, trace)

### migrate
Manage database migrations:

```bash
# Apply all pending migrations
go run main.go migrate up

# Rollback last migration
go run main.go migrate down

# Show migration status
go run main.go migrate status
```

### seed
Seed the database with initial data:
```bash
go run main.go seed
```

### backup
Create a database backup:
```bash
go run main.go backup
```

Options:
- `--output` - Specify backup file path

## Testing

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Specific Package Tests
```bash
go test ./domain/user
go test ./usecase/auth
go test ./pkg/pagination
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Generate Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Structure

This project follows Clean Architecture principles with clear separation of concerns:

```
webdesa/api/
├── cmd/                    # CLI commands and application entry points
│   ├── root.go            # Root command and initialization
│   ├── serve.go           # HTTP server command
│   ├── migrate.go         # Database migration commands
│   ├── seed.go            # Data seeding command
│   └── backup.go          # Backup command
│
├── config/                 # Configuration management
│   ├── app.go             # Application config
│   ├── postgres.go        # Database config
│   ├── jwt.go             # JWT config
│   ├── backup.go          # Backup config
│   ├── cors.go            # CORS config
│   ├── ratelimit.go       # Rate limiting config
│   ├── fileupload.go      # File upload config
│   ├── otel.go            # Observability config
│   └── *.toml             # Environment-specific configs
│
├── domain/                 # Domain layer (entities)
│   ├── user/              # User entity and validation
│   ├── desa/              # Village entity
│   ├── struktur/          # Organization structure entity
│   ├── berita/            # News entity
│   ├── banner/            # Banner entity
│   ├── umkm/              # Business entity
│   ├── fasilitas/         # Facility entity
│   └── ppid/              # Public document entity
│
├── usecase/                # Use case layer (business logic)
│   ├── auth/              # Authentication service
│   ├── user/              # User management service
│   ├── role/              # Role management service
│   ├── desa/              # Village service
│   ├── struktur/          # Structure service
│   ├── berita/            # News service
│   ├── banner/            # Banner service
│   ├── umkm/              # Business service
│   ├── fasilitas/         # Facility service
│   ├── ppid/              # Document service
│   └── backup/            # Backup service
│
├── interface/              # Interface adapters layer
│   ├── http/              # HTTP delivery
│   │   ├── handler/       # HTTP handlers
│   │   ├── middleware/    # HTTP middleware
│   │   └── router.go      # Route definitions
│   ├── repository/        # PostgreSQL repositories
│   ├── postgres/          # Auth-specific repositories
│   ├── rbac/              # Casbin enforcer implementation
│   └── file/              # File storage handler
│
├── pkg/                    # Shared utilities (framework-agnostic)
│   ├── clock/             # Time utilities
│   ├── jwt/               # JWT utilities
│   ├── password/          # Password hashing
│   ├── response/          # HTTP response helpers
│   ├── slug/              # Slug generation
│   ├── pagination/        # Pagination utilities
│   └── validator/         # Validation utilities
│
├── db/migrations/          # Database migrations (Goose)
├── rbac/                   # Casbin RBAC configuration
│   ├── rbac_model.conf    # Casbin model definition
│   └── rbac_policy.csv    # Casbin policies
│
├── system/                 # System constants
├── main.go                 # Application entry point
└── config.toml             # Runtime configuration
```

### Architecture Layers

**Domain Layer** (`domain/`):
- Contains business entities and domain logic
- No dependencies on external packages
- Defines validation rules for entities
- Represents the core business concepts

**Use Case Layer** (`usecase/`):
- Contains business logic and application services
- Defines repository interfaces (dependency inversion)
- Orchestrates domain entities
- Independent of delivery mechanisms

**Interface Layer** (`interface/`):
- HTTP handlers expose use cases via REST API
- PostgreSQL repositories implement use case interfaces
- Casbin enforcer implements authorization interface
- File handlers implement storage interface

**Infrastructure** (`pkg/`, `config/`, `cmd/`):
- Reusable utilities and helpers
- Configuration management
- CLI commands and application bootstrap

This structure ensures:
- Dependencies point inward (Dependency Rule)
- Business logic is independent of frameworks
- Easy to test each layer in isolation
- Clear separation of concerns
