# WebDesa API - Technical Specification

## 1. 🎯 High-Level Overview

### Core Purpose
RESTful backend API for managing village (desa) information, providing authentication, role-based access control (RBAC), content management for news (berita), UMKM businesses, facilities, public information disclosure (PPID), organizational structure, and database backup/restore operations.

### Tech Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.24.1 |
| **HTTP Framework** | Chi v5 (`go-chi/chi/v5` v5.2.5) |
| **Database** | PostgreSQL with `jmoiron/sqlx` (v1.4.0) |
| **RBAC** | Casbin v2 (v2.135.0) |
| **Authentication** | JWT (`golang-jwt/jwt/v5` v5.3.1) |
| **CLI** | Cobra + Viper |
| **Migrations** | Goose v3 (v3.26.0) |
| **Observability** | OpenTelemetry (OTLP HTTP exporter) |
| **Testing** | testify (v1.11.1) + testcontainers |
| **Request Parsing** | httpin (v0.20.3) |
| **Validation** | go-playground/validator (v10.30.1) |

### Architecture Pattern
**Clean Architecture** with 4 distinct layers following the Dependency Rule (source code dependencies point INWARD only):

```
┌─────────────────────────────────────────────────────────────┐
│                     Interface Layer                        │
│        (HTTP Handlers, Repositories, External Services)     │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      Use Case Layer                         │
│             (Business Logic + Repository Interfaces)        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                       Domain Layer                          │
│                    (Pure Entities Only)                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 📁 Directory Blueprint

```
├── cmd/                          # CLI commands
│   ├── root.go                   # Root command, config/otel initialization
│   ├── serve.go                  # HTTP server startup + DI wiring
│   ├── migrate.go                # Database migration runner
│   ├── seed.go                   # Database seeder
│   └── backup.go                 # Backup management (disabled)
├── config/                        # Configuration structs (Viper + TOML)
│   ├── app.go                    # Main Config struct
│   ├── postgres.go, jwt.go, dll
│   └── dev.toml, staging.toml, prod.toml
├── domain/                        # Pure domain entities (NO external dependencies)
│   ├── user/user.go
│   ├── banner/banner.go
│   ├── berita/berita.go
│   ├── umkm/umkm.go
│   ├── fasilitas/fasilitas.go
│   ├── ppid/ppid.go              # PPID + PPIDRequest entities
│   ├── struktur/struktur.go
│   ├── desa/desa.go
│   ├── profile/profile.go
│   ├── infographic/infographic.go
│   └── backup/backup.go
├── usecase/                       # Business logic + interface definitions
│   ├── auth/                     # Service + Repository interface
│   ├── user/
│   ├── banner/
│   ├── berita/
│   ├── umkm/
│   ├── fasilitas/
│   ├── ppid/                     # Includes approval workflow
│   ├── struktur/
│   ├── desa/
│   ├── profile/
│   ├── infographic/              # Metabase JWT integration
│   └── backup/
├── interface/                     # Adapters (implementations)
│   ├── http/
│   │   ├── router.go            # Route registration + middleware
│   │   └── handler/             # Thin HTTP handlers (delegate to usecase)
│   │       ├── auth.go, user.go, role.go
│   │       ├── banner.go, berita.go, umkm.go
│   │       ├── fasilitas.go, ppid.go, struktur.go
│   │       ├── desa.go, profile.go, infographic.go
│   │       └── file.go, backup.go
│   ├── http/middleware/
│   │   ├── auth.go              # JWT validation
│   │   ├── rbac.go              # Casbin permission check
│   │   ├── ratelimit.go         # Token bucket (in-memory)
│   │   ├── cors.go, logger.go
│   │   ├── recovery.go, request_id.go
│   │   └── timeout.go
│   ├── repository/               # PostgreSQL implementations
│   │   ├── user_postgres.go, role_postgres.go
│   │   ├── banner_postgres.go, berita_postgres.go
│   │   ├── umkm_postgres.go, fasilitas_postgres.go
│   │   ├── ppid_postgres.go, struktur_postgres.go
│   │   ├── desa_postgres.go
│   │   ├── profile_postgres.go, infographic_postgres.go
│   │   └── backup_postgres.go
│   ├── postgres/                 # Auth/user-specific DB code
│   │   ├── user_repository.go
│   │   └── auth_repository.go
│   ├── rbac/
│   │   └── casbin_enforcer.go   # Casbin adapter
│   ├── email/
│   │   └── ppid_email_service.go
│   └── file/
│       └── local_handler.go      # Local file storage
├── pkg/                          # Shared utilities
│   ├── jwt/                      # JWT generation/validation
│   ├── hash/                     # Bcrypt password hashing
│   ├── slug/                     # Slug generation
│   ├── clock/                    # Time interface (for testing)
│   ├── pagination/               # Pagination helpers
│   ├── casbin/                   # Custom Casbin adapter with sqlx
│   └── otel/                     # OpenTelemetry setup
├── db/migrations/                 # Goose SQL migrations (26 files)
│   ├── 00001_create_users_table.sql
│   ├── ...
│   └── 00026_add_thumbnail_to_ppid.sql
├── rbac/
│   ├── rbac_model.conf           # Casbin model definition
│   └── policy.csv                # Default RBAC policies
├── .kiro/                        # Kiro workflow specs
│   ├── specs/desa-api-management/
│   │   ├── requirements.md
│   │   ├── design.md
│   │   └── tasks.md
│   ├── specs/ppid-request-approval-workflow/
│   │   ├── requirements.md
│   │   ├── design.md
│   │   └── tasks.md
│   ├── specs/public-endpoints.md  # Planned but NOT implemented
│   ├── steering/                  # Architecture guidelines
│   └── tasks.md
├── integration/                   # Integration tests
│   ├── setup_test.go
│   ├── helpers_test.go
│   └── *_test.go
├── backups/                       # Backup storage directory
├── uploads/                       # File upload storage directory
├── main.go                        # Entry point
└── openapi.yaml                   # OpenAPI 3.0 specification
```

### Domain Boundaries

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Domain** | `domain/` | Pure entities with validation. No external dependencies. |
| **Use Case** | `usecase/` | Business logic orchestration. Defines repository interfaces. |
| **Interface** | `interface/` | HTTP handlers, DB repositories, external service adapters. |
| **Infrastructure** | `cmd/`, `config/` | CLI, configuration, bootstrapping. |

---

## 3. 🚀 Entry Points & Execution Flow

### Bootstrapping Sequence

```
main.go:8
    │
    ▼
cmd.Execute()                                    [root.go:92]
    │
    ▼
Initialize(ctx)                                 [root.go:43]
    ├── config.InitConfig(cfgFileInput)          Load TOML via Viper
    ├── otel.SetupOTelSDK()                      Initialize OpenTelemetry
    ├── debug.SetProfiler()                      Setup pprof profiling
    ├── config.SetUpTimezone()                   Set timezone
    └── config.SetProxy()                        Setup HTTP proxy (optional)
    │
    ▼
rootCmd.ExecuteContext(ctx)                      [root.go:106]
    │
    ▼
serveCmd.RunE → runServer()                     [serve.go:44-46]
    │
    ├── sqlx.Connect("postgres", DSN)           Database connection
    ├── db.SetMaxOpenConns(25)                   Connection pool config
    ├── db.SetMaxIdleConns(5)
    │
    ├── goose.Up(db, "db/migrations")            Apply migrations
    │
    ├── rbac.NewCasbinEnforcer()                 Initialize Casbin RBAC
    │
    ├── Initialize repositories                   Create PostgreSQL repos
    │
    ├── Initialize services                       Create business logic services
    │
    ├── Initialize handlers                       Create HTTP handlers
    │
    ├── httpserver.NewRouter(RouterConfig{})     Wire routes + middleware
    │
    └── http.Server.ListenAndServe()             Start HTTP server
```

### Middleware Chain

Applied in order for every request:

```
1. RecoveryMiddleware     → Panic recovery + error logging
2. RequestIDMiddleware    → Generate unique request ID (X-Request-ID)
3. LoggerMiddleware       → Structured request/response logging
4. CORSMiddleware         → Cross-origin resource sharing
5. TimeoutMiddleware      → 60 second request timeout
6. RateLimitMiddleware    → Token bucket (varies by route tier)
```

### Data Lifecycle Example: Create Berita

```
HTTP Request: POST /api/v1/berita
    │
    ▼
Chi Router matches route
    │
    ▼
Middleware Chain (auth + RBAC pass through)
    │
    ▼
BeritaHandler.Create()                           [interface/http/handler/berita.go]
    │
    ▼
beritaService.Create(ctx, req)                   [usecase/berita/service.go]
    │   - Validates input
    │   - Hashes/processes image if present
    │   - Calls repository
    ▼
beritaRepo.Create(ctx, entity)                    [interface/repository/berita_postgres.go]
    │
    ▼
sqlx.DB.ExecContext()                            [PostgreSQL]
    │
    ▼
Return created entity back up the chain
    │
    ▼
HTTP 201 Response: { "success": true, "data": {...} }
```

### API Route Structure

```
/health                                          # Health check (public)
/uploads/*                                       # Static file serving (public)
/api/v1/
├── auth/                                        # Public (rate limited: 5/min)
│   ├── POST /auth/login
│   ├── POST /auth/password-reset/request
│   ├── GET  /auth/password-reset/check/{token}
│   └── PUT  /auth/password-reset
│
├── public/                                     # Public (rate limited: 100/min) [NOT IMPLEMENTED]
│   ├── GET /public/banner/list
│   ├── GET /public/berita/list
│   └── ...
│
├── banners                                      # Protected (JWT + RBAC + 30/min)
├── berita
├── umkm
├── fasilitas
├── ppid
│   ├── GET    /ppid/requests                   # List requests (admin)
│   ├── POST    /ppid/requests/{id}/approve     # Approve (admin)
│   ├── POST    /ppid/requests/{id}/revoke      # Revoke (admin)
│   └── GET    /ppid/documents/{id}/download    # Public (token auth)
├── struktur-organisasi
├── users
├── roles
├── profile
├── infographic
├── backup                                      # Routes exist but commented out
└── restore                                     # Routes exist but commented out
```

---

## 4. ⚠️ Technical Debt & Security Review

### 🔴 Red Flags

| Issue | Location | Impact | Severity |
|-------|----------|--------|----------|
| **Public endpoints not implemented** | `router.go` | Feature defined in `.kiro/specs/public-endpoints.md` but routes not registered | Medium |
| **Backup routes commented out** | `router.go:~280` | Backup/restore functionality exists but disabled | Medium |
| **Synchronous email sending** | `usecase/ppid/service.go` | PPID approval email blocks request | Medium |
| **In-memory rate limiting** | `middleware/ratelimit.go` | Cannot scale horizontally (sticky sessions required) | Medium |
| **No cache layer** | N/A | Every read request hits database | Medium |
| **No message queue** | N/A | Async workflows require external implementation | Low |

### 🟡 Design Divergence from Kiro Spec

| Kiro Spec Item | Implementation Status |
|---------------|----------------------|
| Property-based tests with gopter | Not implemented |
| Public endpoints (`/public/*`) | Defined in spec, NOT implemented |
| Backup/Restore endpoints | Code exists, routes commented out |
| Test execution via Makefile | Tests run directly with `go test` |

### 🟢 Security Best Practices (Implemented)

| Practice | Status | Location |
|----------|--------|----------|
| Parameterized SQL queries | ✅ | All repository implementations |
| Bcrypt password hashing (cost 12) | ✅ | `pkg/hash/` |
| JWT token authentication | ✅ | `pkg/jwt/` |
| RBAC with Casbin | ✅ | `interface/rbac/` |
| Input validation | ✅ | Domain entities + httpin |
| Rate limiting | ✅ | `middleware/ratelimit.go` |
| CORS configuration | ✅ | `middleware/cors.go` |
| Error wrapping with stack traces | ✅ | `braces.dev/errtrace` |
| Clock interface for time operations | ✅ | `pkg/clock/` |
| No secrets in logs | ✅ | Passwords/tokens redacted |

### 🟢 Code Quality Patterns (Correct)

| Pattern | Status | Evidence |
|---------|--------|----------|
| Interfaces defined at consumer | ✅ | `usecase/*/repository.go` defines interfaces |
| Domain entities have no deps | ✅ | `domain/` imports only stdlib |
| Services accept interfaces | ✅ | Constructor injection pattern |
| Concrete returns, interface params | ✅ | Standard Go idiom |
| Clean Architecture layers | ✅ | Proper dependency direction |

---

## 5. 💡 Onboarding Recommendations

### Quick Start

```bash
# 1. Clone and enter directory
cd api

# 2. Copy configuration
cp config.example.toml config.toml

# 3. Edit config.toml
# Required: PostgreSQL connection (POSTGRES_* vars)
# Required: JWT secret (JWT_SECRET)
# Optional: SMTP for email notifications

# 4. Start PostgreSQL (Docker)
docker-compose up -d postgres

# 5. Run migrations
go run main.go migrate up

# 6. Seed initial data (optional)
go run main.go seed

# 7. Start server
go run main.go serve --config config.toml

# 8. Run tests
go test ./... -v -count=1
```

### Required Configuration

```toml
# config.toml

[app]
http_addr = ":8080"
timezone = "Asia/Jakarta"
domain_addr = "http://localhost:8080"

[database]
host = "localhost"
port = 5432
name = "desa_dev"
user = "postgres"
password = "your-password"
max_open_conns = 25
max_idle_conns = 5

[jwt]
secret = "your-256-bit-secret-key-change-in-production"
expiration = "24h"
refresh_expiration = "168h"  # 7 days

[file_upload]
max_image_size_mb = 10
max_document_size_mb = 50
upload_dir = "./uploads"
ppid_upload_dir = "./uploads/ppid"

[rate_limit]
auth_requests_per_minute = 5
public_requests_per_minute = 100
protected_requests_per_minute = 30

[cors]
allowed_origins = ["http://localhost:3000"]
```

### Environment Variables (Override TOML)

```bash
JWT_SECRET=your-secret
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=desa_dev
```

### First 3 Files to Read (Understanding the System)

| Priority | File | Why |
|----------|------|-----|
| 1 | `cmd/serve.go` | Shows complete dependency injection wiring - how all layers connect |
| 2 | `interface/http/router.go` | Maps routes to handlers, applies middleware chain |
| 3 | `usecase/berita/service.go` | Example of Clean Architecture service pattern with repository interface |

### Subsequent Reading Path

```
1. cmd/serve.go              # Bootstrap + DI
2. interface/http/router.go   # Route structure
3. usecase/berita/service.go  # Business logic pattern
4. interface/repository/berita_postgres.go  # Data persistence
5. domain/berita/berita.go    # Entity definition + validation
6. interface/http/handler/berita.go  # HTTP adapter
7. pkg/jwt/                   # Authentication mechanism
8. interface/rbac/casbin_enforcer.go  # Authorization
```

---

## 6. 📊 Database Schema

### Entity Relationship Overview

```
users (1) ←→ (N) user_roles
users (1) ←→ (N) password_reset_tokens
users (1) ←→ (N) ppid_requests (approved_by, revoked_by FK)

ppid (1) ←→ (N) ppid_requests

berita_categories        (1) ←→ (N) berita          [category FK, ON DELETE RESTRICT]
umkm_categories          (1) ←→ (N) umkm            [category FK, ON DELETE RESTRICT]
ppid_categories          (1) ←→ (N) ppid            [category FK, ON DELETE RESTRICT, nullable]
fasilitas_categories     (1) ←→ (N) fasilitas       [category FK, ON DELETE RESTRICT, nullable]
infographic_categories   (1) ←→ (N) infographic     [category FK, ON DELETE RESTRICT]

banners, struktur_organisasi, profile
→ No foreign keys (except as noted), standalone content tables
```

### Key Tables

| Table | Purpose | Key Indexes |
|-------|---------|-------------|
| `users` | User accounts | `idx_users_email` |
| `user_roles` | Role assignments | `idx_user_roles_user_id`, `idx_user_roles_role` |
| `password_reset_tokens` | Password reset tokens | `idx_reset_tokens_token`, `idx_reset_tokens_user_id` |
| `banners` | Promotional banners | `idx_banners_status` |
| `desa` | Village profile | (PK only) |
| `berita_categories` | **Managed vocabulary for berita** | `idx_berita_categories_name` (unique) |
| `umkm_categories` | **Managed vocabulary for UMKM** | `idx_umkm_categories_name` (unique) |
| `ppid_categories` | **Managed vocabulary for PPID** | `idx_ppid_categories_name` (unique) |
| `fasilitas_categories` | **Managed vocabulary for Fasilitas** | `idx_fasilitas_categories_name` (unique) |
| `infographic_categories` | **Managed vocabulary for Infographic** | `idx_infographic_categories_name` (unique) |
| `berita` | News articles | `idx_berita_category` (FK to berita_categories) |
| `umkm` | Small businesses | `idx_umkm_category` (FK to umkm_categories) |
| `ppid` | Public info documents | `idx_ppid_category` (FK to ppid_categories, nullable) |
| `fasilitas` | Village facilities | `idx_fasilitas_location` (lat, lon), `idx_fasilitas_category` (FK) |
| `infographic` | Dashboard widgets | `idx_infographic_section_name`, `idx_infographic_state`, `idx_infographic_category` (FK) |
| `berita` | News articles | `idx_berita_created_at` |
| `umkm` | Business listings | `idx_umkm_name` |
| `ppid` | Public info documents | `idx_ppid_category`, `idx_ppid_title` |
| `ppid_requests` | Document access requests | `idx_ppid_requests_ppid_id`, `idx_ppid_requests_status` |
| `struktur_organisasi` | Organization members | `idx_struktur_name`, `idx_struktur_position` |
| `backups` | Backup metadata | `idx_backups_created_at` |
| `profile` | Village profile sections | `idx_profile_section_name`, `idx_profile_state` |
| `infographic` | Metabase infographics | `idx_infographic_section_name`, `idx_infographic_state` |
| `casbin_rule` | RBAC policies | (managed by Casbin) |

---

## 7. 🔐 Security Model

### Authentication Flow

```
1. User submits credentials to POST /auth/login
2. AuthService.Login() validates against bcrypt hash
3. Generates JWT access token (24h) + refresh token (7 days)
4. Returns tokens in response
5. Client includes access token as "Bearer <token>" in Authorization header
6. AuthMiddleware validates token, extracts user ID + roles
7. RBACMiddleware checks permissions against Casbin policies
```

### RBAC Permission Matrix

| Role | Permissions |
|------|-------------|
| **Admin** | `*:*` (all access) |
| **Operator** | `berita:*`, `umkm:*`, `fasilitas:*`, `ppid:*` |
| **User** | Varies by assignment |

### Rate Limiting Tiers

| Tier | Limit | Scope |
|------|-------|-------|
| Auth endpoints | 5 req/min | Per IP |
| Public endpoints | 100 req/min | Per IP |
| Protected endpoints | 30 req/min | Per User |
| PPID download | 10 req/min | Per IP |
| Password reset | 5 req/hour | Per email |

---

## 8. 🔌 External Integrations

### Metabase (Infographics)

- JWT token generation for embedding dashboards/questions
- HMAC-SHA256 signing with configurable secret
- 10-minute token expiration
- Supports `question` and `dashboard` component types

### Email (PPID Approvals)

- SMTP-based email service
- HTML templates for approval notifications
- Async sending (non-blocking)
- Configurable village branding

---

## 9. 📝 OpenAPI Coverage

The `openapi.yaml` (52KB) documents all API endpoints with:
- Request/response schemas
- Authentication requirements
- RBAC permission annotations
- Example payloads

---

## 10. 🧪 Testing Strategy

### Test Types

| Type | Location | Framework |
|------|----------|-----------|
| Integration tests | `integration/` | testcontainers + Chi server |
| Unit tests | Alongside code (`*_test.go`) | testify |
| Handler tests | `integration/*_test.go` | HTTP tests |

### Running Tests

```bash
# All tests
go test ./... -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Specific package
go test ./usecase/berita/... -v
```

---

## Appendix A: Kiro Spec Status

### Completed Specs

| Spec | Status |
|------|--------|
| desa-api-management | ✅ Implemented |
| ppid-request-approval-workflow | ✅ Implemented |

### Planned Specs (NOT Implemented)

| Spec | Status | Notes |
|------|--------|-------|
| public-endpoints | ❌ Not implemented | Routes defined in `.kiro/specs/public-endpoints.md` but not in router |
| Backup/Restore endpoints | ⚠️ Partial | Code exists, routes commented out |

---

## Appendix B: Category Refactor (2026)

### Summary

The free-text `category` column on `berita` and `umkm` has been replaced by a
FK to a dedicated category table per feature. Categories are now a
first-class, managed vocabulary with **create + delete only** operations —
intentionally no `Update` endpoint (to rename, delete and re-create).

### Schema changes

- New table `berita_categories (id UUID, name VARCHAR(100) UNIQUE, created_at, updated_at)`
- New table `umkm_categories (id UUID, name VARCHAR(100) UNIQUE, created_at, updated_at)`
- `berita.category` and `umkm.category` migrated from `VARCHAR(100)` → `UUID NOT NULL` with `ON DELETE RESTRICT` FK
- Migrations `00027`–`00030` perform the conversion with backfill
  (`"Lainnya"` for berita, `"Umum"` for UMKM as the empty-row fallback)

### Endpoints (per feature)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/public/{berita,umkm}/categories` | public | Paginated, with `q` for autocomplete; returns UUIDs + `usage_count` |
| `POST` | `/api/v1/{berita,umkm}/categories` | `{feature}:write` | Create; 409 on duplicate name |
| `DELETE` | `/api/v1/{berita,umkm}/categories/{id}` | `{feature}:write` | Delete; 409 if referenced by any record |

### Why no Update?

Categories are a small, stable vocabulary. Allowing edits leads to:
- Broken URLs / slugs that point to the old name
- Confusion when autocomplete matches a renamed category
- Audit log complexity

A delete + re-create flow keeps the data model clean and the audit trail
honest.

### Files added

```
domain/beritacategory/beritacategory.go
domain/umkmcategory/umkmcategory.go
usecase/beritacategory/{repository,service,service_test}.go
usecase/umkmcategory/{repository,service}.go
usecase/berita/category_lookup.go
usecase/umkm/category_lookup.go
interface/repository/beritacategory_postgres.go
interface/repository/umkmcategory_postgres.go
interface/http/handler/beritacategory.go
interface/http/handler/umkmcategory.go
integration/beritacategory_test.go
integration/umkmcategory_test.go
db/migrations/00027_create_berita_categories_table.sql
db/migrations/00028_create_umkm_categories_table.sql
db/migrations/00029_convert_berita_category_to_fk.sql
db/migrations/00030_convert_umkm_category_to_fk.sql
```

### Out of scope

- PPID still uses a free-text nullable `category` column.
- No bulk-import endpoint; categories are created one at a time.

---

*Document generated from implementation analysis. Kiro specs used as reference only.*