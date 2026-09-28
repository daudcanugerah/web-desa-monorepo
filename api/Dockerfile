# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Copy manifests of locally-replaced modules so `go mod download` can resolve them
COPY pkg/otel/go.mod pkg/otel/go.sum ./pkg/otel/
COPY pkg/profiler/go.mod pkg/profiler/go.sum ./pkg/profiler/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate swagger docs (excluded by .dockerignore, imported by cmd/serve.go)
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3 && \
    swag init -g main.go -o docs --parseDependency --parseInternal --parseDepth 10 --exclude stable && \
    go run ./scripts/swag-tags

# Build the application (GOARCH follows the target platform: amd64/arm64)
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -ldflags="-w -s" -o /app/bin/desa-api .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata ffmpeg

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/desa-api /app/desa-api

# Copy configuration files
COPY --from=builder /app/config /app/config

# Copy migrations
COPY --from=builder /app/db/migrations /app/db/migrations

# Copy RBAC policies if they exist
COPY --from=builder /app/rbac /app/rbac

# Create directories for uploads and backups
RUN mkdir -p /app/uploads /app/backups && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/desa-api", "serve"]
