# ============================================================
# Desa API — Makefile
# ============================================================

CONFIG_FILE ?= config.toml
API_URL     ?= http://localhost:8080
DB_PORT     ?= 5433

GREEN  := \033[0;32m
BLUE   := \033[0;34m
YELLOW := \033[1;33m
RED    := \033[0;31m
NC     := \033[0m

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo '$(BLUE)Desa API — available targets$(NC)'
	@echo ''
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================================
# Development
# ============================================================

.PHONY: dev start serve dev-watch serve-bg stop

dev: db-up migrate seed dev-watch ## Full dev setup with hot reload (db + migrate + seed + watch)

start: db-up migrate seed ## Start PostgreSQL, run migrations, and seed data
	@echo "$(GREEN)✓ Development environment ready$(NC)"
	@echo "$(YELLOW)Run 'make serve' to start the API server$(NC)"

serve: ## Start the API server
	@echo "$(BLUE)Starting Desa API server...$(NC)"
	@go run main.go --config $(CONFIG_FILE) serve

dev-watch: ## Start the API server with hot reload (Air)
	@echo "$(BLUE)Starting Desa API server with hot reload...$(NC)"
	@if ! command -v air > /dev/null; then \
		echo "$(YELLOW)Air not found. Installing...$(NC)"; \
		go install github.com/air-verse/air@latest; \
	fi
	@air

serve-bg: ## Start the API server in the background
	@echo "$(BLUE)Starting Desa API server in background...$(NC)"
	@go run main.go --config $(CONFIG_FILE) serve > /tmp/desa-api.log 2>&1 & echo $$! > /tmp/desa-api.pid
	@sleep 2
	@if ps -p $$(cat /tmp/desa-api.pid) > /dev/null; then \
		echo "$(GREEN)✓ API server started (PID: $$(cat /tmp/desa-api.pid))$(NC)"; \
	else \
		echo "$(RED)✗ Failed to start API server$(NC)"; \
		exit 1; \
	fi

stop: ## Stop the background API server
	@if [ -f /tmp/desa-api.pid ]; then \
		kill $$(cat /tmp/desa-api.pid) 2>/dev/null || true; \
		rm -f /tmp/desa-api.pid; \
		echo "$(GREEN)✓ API server stopped$(NC)"; \
	else \
		echo "$(YELLOW)API server is not running$(NC)"; \
	fi

# ============================================================
# Database
# ============================================================

.PHONY: db-up db-down db-logs db-shell db-reset

db-up: ## Start PostgreSQL
	@echo "$(BLUE)Starting PostgreSQL...$(NC)"
	@podman compose up -d postgres
	@echo "$(YELLOW)Waiting for PostgreSQL to be ready...$(NC)"
	@until podman compose exec -T postgres pg_isready -U desauser -d desa > /dev/null 2>&1; do \
		echo "Waiting for PostgreSQL..."; \
		sleep 2; \
	done
	@echo "$(GREEN)✓ PostgreSQL is ready$(NC)"

db-down: ## Stop PostgreSQL
	@echo "$(YELLOW)Stopping PostgreSQL...$(NC)"
	@podman compose down
	@echo "$(GREEN)✓ PostgreSQL stopped$(NC)"

db-logs: ## Tail PostgreSQL logs
	@podman compose logs -f postgres

db-shell: ## Open a psql shell to the database
	@podman compose exec postgres psql -U desauser -d desa

db-reset: db-down db-up migrate seed ## Reset database (down + up + migrate + seed)
	@echo "$(GREEN)✓ Database reset complete$(NC)"

# ============================================================
# Migrations
# ============================================================

.PHONY: migrate migrate-down migrate-status seed

migrate: ## Apply all pending migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	@go run main.go --config $(CONFIG_FILE) migrate up
	@echo "$(GREEN)✓ Migrations completed$(NC)"

migrate-down: ## Roll back the last migration
	@echo "$(BLUE)Rolling back last migration...$(NC)"
	@go run main.go --config $(CONFIG_FILE) migrate down
	@echo "$(GREEN)✓ Migration rolled back$(NC)"

migrate-status: ## Show migration status
	@echo "$(BLUE)Checking migration status...$(NC)"
	@go run main.go --config $(CONFIG_FILE) migrate status

seed: ## Seed initial admin and operator data
	@echo "$(BLUE)Seeding data...$(NC)"
	@go run main.go --config $(CONFIG_FILE) seed
	@echo "$(GREEN)✓ Data seeded$(NC)"

# ============================================================
# Backup & Restore
# ============================================================

.PHONY: backup-create backup-list backup-restore

backup-create: ## Create a new database backup
	@echo "$(BLUE)Creating database backup...$(NC)"
	@go run main.go --config $(CONFIG_FILE) backup create
	@echo "$(GREEN)✓ Backup created$(NC)"

backup-list: ## List all available database backups
	@echo "$(BLUE)Listing database backups...$(NC)"
	@go run main.go --config $(CONFIG_FILE) backup list

backup-restore: ## Restore database from a backup (usage: make backup-restore BACKUP_ID=<id>)
	@if [ -z "$(BACKUP_ID)" ]; then \
		echo "$(RED)✗ Error: BACKUP_ID is required$(NC)"; \
		echo "$(YELLOW)Usage: make backup-restore BACKUP_ID=<backup-id>$(NC)"; \
		echo "$(YELLOW)Run 'make backup-list' to see available backups$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)⚠ WARNING: This will drop and recreate all database objects$(NC)"
	@read -p "Are you sure you want to continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "$(BLUE)Restoring database from backup ID: $(BACKUP_ID)...$(NC)"; \
		go run main.go --config $(CONFIG_FILE) backup restore $(BACKUP_ID); \
		echo "$(GREEN)✓ Database restored$(NC)"; \
	else \
		echo "$(YELLOW)Restore cancelled$(NC)"; \
	fi

# ============================================================
# Testing
# ============================================================

.PHONY: test test-unit test-integration test-cover

test: ## Run all Go tests (unit + integration)
	@echo "$(BLUE)Running all tests...$(NC)"
	@go test -v ./...

test-unit: ## Run unit tests only (exclude ./integration)
	@echo "$(BLUE)Running unit tests...$(NC)"
	@go test -v $$(go list ./... | grep -v /integration)

test-integration: ## Run integration tests (testcontainers by default; set INTEGRATION_TEST_DB_URL to skip)
	@echo "$(BLUE)Running integration tests...$(NC)"
	@if [ -n "$$INTEGRATION_TEST_DB_URL" ]; then \
		echo "$(YELLOW)Using INTEGRATION_TEST_DB_URL (skipping testcontainers)$(NC)"; \
	fi
	@go test -v ./integration/... -timeout 10m

test-integration-local: ## Run integration tests against the local docker-compose Postgres (no testcontainers)
	@echo "$(BLUE)Running integration tests against local Postgres :5433...$(NC)"
	@make db-up migrate >/dev/null 2>&1
	@INTEGRATION_TEST_DB_URL='postgres://desauser:desapass@localhost:5433/desa?sslmode=disable' \
		go test -v ./integration/... -timeout 10m

test-cover: ## Run tests and generate an HTML coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

# ============================================================
# Swagger / OpenAPI
# ============================================================

.PHONY: swagger-up swagger-down swagger-validate swagger-gen swagger-clean

SWAG := $(shell go env GOPATH)/bin/swag


swagger-sync:
	cp ./docs/swagger.yaml ../admin/swagger.yaml
	cp ./docs/swagger.yaml ../web/openapi.yaml

swagger-gen: ## Regenerate OpenAPI docs (docs/swagger.{json,yaml}) from code annotations via swag
	@echo "$(BLUE)Regenerating OpenAPI docs from code annotations...$(NC)"
	@if [ ! -x "$(SWAG)" ]; then \
		echo "$(YELLOW)Installing swag CLI...$(NC)"; \
		go install github.com/swaggo/swag/cmd/swag@v1.16.3; \
	fi
	@$(SWAG) init -g main.go -o docs --parseDependency --parseInternal --parseDepth 10
	@echo "$(BLUE)Post-processing: injecting top-level tags array...$(NC)"
	@go run ./scripts/swag-tags
	@echo "$(GREEN)✓ Generated docs/swagger.json, docs/swagger.yaml, docs/docs.go$(NC)"

swagger-up: ## Start Swagger UI container serving docs/swagger.yaml at http://localhost:8081
	@echo "$(BLUE)Starting Swagger UI from docs/swagger.yaml on http://localhost:8081...$(NC)"
	@podman stop desa-api-swagger-ui 2>/dev/null || true
	@if ! podman network inspect desa-network > /dev/null 2>&1; then \
		echo "$(BLUE)Creating desa-network...$(NC)"; \
		podman network create desa-network > /dev/null 2>&1; \
	fi
	@if [ ! -f docs/swagger.yaml ]; then \
		echo "$(YELLOW)docs/swagger.yaml not found — generating first$(NC)"; \
		$(MAKE) swagger-gen; \
	fi
	@podman run -d --rm \
		--name desa-api-swagger-ui \
		--network desa-network \
		-p 8081:8080 \
		-e SWAGGER_JSON=/swagger/swagger.yaml \
		-v "$$(pwd)/docs/swagger.yaml:/swagger/swagger.yaml:ro" \
		swaggerapi/swagger-ui > /dev/null 2>&1
	@sleep 2
	@echo "$(GREEN)✓ Swagger UI started$(NC)"
	@echo "$(BLUE)Open in browser: http://localhost:8081$(NC)"

swagger-down: ## Stop Swagger UI container
	@podman stop desa-api-swagger-ui 2>/dev/null || true
	@echo "$(GREEN)✓ Swagger UI stopped$(NC)"

swagger-validate: ## Validate the generated OpenAPI spec against redocly.yaml ruleset
	@echo "$(BLUE)Validating generated OpenAPI spec (docs/swagger.yaml) against redocly.yaml...$(NC)"
	@if command -v swagger-cli > /dev/null 2>&1; then \
		swagger-cli validate docs/swagger.yaml; \
	elif command -v npx > /dev/null 2>&1; then \
		npx --yes @redocly/cli lint docs/swagger.yaml --config redocly.yaml; \
	else \
		echo "$(YELLOW)No validator CLI found; running swag generation as syntax check$(NC)"; \
		$(MAKE) swagger-gen; \
	fi
	@echo "$(GREEN)✓ OpenAPI validation passed$(NC)"

# ============================================================
# Build
# ============================================================

.PHONY: build build-linux run-binary

build: ## Build the binary
	@echo "$(BLUE)Building binary...$(NC)"
	@go build -o desa-api
	@echo "$(GREEN)✓ Binary created: ./desa-api$(NC)"

build-linux: ## Build a Linux/amd64 binary
	@echo "$(BLUE)Building for Linux/amd64...$(NC)"
	@GOOS=linux GOARCH=amd64 go build -o desa-api-linux
	@echo "$(GREEN)✓ Binary created: ./desa-api-linux$(NC)"

run-binary: build ## Build and run the binary
	@./desa-api --config $(CONFIG_FILE) serve

# ============================================================
# podman (full stack)
# ============================================================

.PHONY: podman-build podman-up podman-down podman-logs

podman-build: ## Build the podman image
	@echo "$(BLUE)Building podman image...$(NC)"
	@podman build -t desa-api:latest .
	@echo "$(GREEN)✓ podman image built: desa-api:latest$(NC)"

podman-up: ## Start the full podman-compose stack
	@echo "$(BLUE)Starting services with podman compose...$(NC)"
	@podman compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"

podman-down: ## Stop the full podman-compose stack
	@echo "$(YELLOW)Stopping services with podman compose...$(NC)"
	@podman compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

podman-logs: ## Tail podman-compose logs
	@podman compose logs -f

# ============================================================
# Code Quality
# ============================================================

.PHONY: deps fmt vet lint check

deps: ## Download, verify, and tidy module dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated and verified$(NC)"

fmt: ## Format all Go code
	@echo "$(BLUE)Formatting code...$(NC)"
	@gofmt -w -s .
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet passed$(NC)"

lint: ## Run golangci-lint (auto-installs if missing)
	@echo "$(BLUE)Running linter...$(NC)"
	@if ! command -v golangci-lint > /dev/null 2>&1; then \
		echo "$(YELLOW)golangci-lint not found. Installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run
	@echo "$(GREEN)✓ Lint passed$(NC)"

check: fmt vet lint test ## Run fmt + vet + lint + tests

# ============================================================
# Cleanup
# ============================================================

.PHONY: clean

clean: stop podman-down ## Stop containers and remove build artifacts
	@echo "$(YELLOW)Cleaning up...$(NC)"
	@podman compose down -v
	@rm -rf backups/*
	@rm -f  desa-api desa-api-linux
	@rm -f  coverage.out coverage.html
	@rm -f  /tmp/desa-api.log
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

# ============================================================
# Status
# ============================================================

.PHONY: status logs

status: ## Show status of services
	@echo "$(BLUE)Service Status$(NC)"
	@echo ""
	@echo "$(YELLOW)podman containers:$(NC)"
	@podman compose ps 2>/dev/null || echo "  (podman compose not available)"
	@echo ""
	@if [ -f /tmp/desa-api.pid ] && ps -p $$(cat /tmp/desa-api.pid) > /dev/null 2>&1; then \
		echo "$(GREEN)✓ API server: running (PID $$(cat /tmp/desa-api.pid))$(NC)"; \
	else \
		echo "$(RED)✗ API server: not running$(NC)"; \
	fi
	@if podman ps --format '{{.Names}}' 2>/dev/null | grep -q '^desa-api-swagger-ui$$'; then \
		echo "$(GREEN)✓ Swagger UI: running$(NC)"; \
	else \
		echo "$(RED)✗ Swagger UI: not running$(NC)"; \
	fi
	@echo ""
	@echo "$(YELLOW)Endpoints:$(NC)"
	@echo "  • API:     $(GREEN)$(API_URL)$(NC)"
	@echo "  • Swagger: $(GREEN)http://localhost:8081$(NC)"
	@echo "  • DB:      $(GREEN)localhost:$(DB_PORT)$(NC)"

logs: ## Tail the background API server log
	@tail -f /tmp/desa-api.log
