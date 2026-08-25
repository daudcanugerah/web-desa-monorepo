# Docker Images Smaller & Secure for Kubernetes

This guide provides clear, step‑by‑step instructions for building Docker images that are **smaller** and **more secure**. It focuses on **multi‑stage builds**, **distroless images**, and a set of battle‑tested best practices. Every concept is explained with a concrete, copy‑ready example. At the end you will find production‑grade `Dockerfile` examples for Go, Node.js, Laravel (serving PHP and static files with Apache) and a static website. All examples use the non‑root user ID `10001` and follow the principle of least privilege.

---

## Why invest time in smaller, more secure images?

In a Kubernetes (k8s) environment, image size and security directly affect:

- **Startup latency** – smaller images are pulled faster, allowing pods to become ready sooner.
- **Resource efficiency** – reduced network I/O and disk usage.
- **Attack surface** – fewer unnecessary tools (shell, package manager, utilities) mean fewer potential exploitation vectors.
- **Compliance** – running as a non‑root user with a read‑only filesystem is often a hard requirement.

---

## 1. Choose a minimal base image that fits your workload

The foundation of your image sets the floor for both size and security. The table below summarises common choices.

| Base image family          | Example tag                           | Approx. compressed size | Shell included? | Package manager? | Use case                                                                               |
| -------------------------- | ------------------------------------- | ----------------------- | --------------- | ---------------- | -------------------------------------------------------------------------------------- |
| **Debian full**            | `debian:bookworm`                     | ~50 MB                  | yes             | apt              | Legacy applications that need full libc and tools                                      |
| **Debian slim**            | `debian:bookworm-slim`                | ~27 MB                  | yes             | apt (limited)    | Python, Node.js apps that need glibc but fewer extras                                  |
| **Debian “stretch”** (old) | `debian:stretch-slim` (EOL)           | ~22 MB                  | yes             | apt              | **Avoid** – outdated, no security patches                                              |
| **Alpine**                 | `alpine:3.19`                         | ~3 MB                   | yes (busybox)   | apk              | Go (build), tiny utilities; uses musl libc                                             |
| **Distroless static**      | `gcr.io/distroless/static-debian12`   | ~1.5 MB                 | no              | none             | Go/C/C++ statically compiled binaries                                                  |
| **Distroless base**        | `gcr.io/distroless/base-debian12`     | ~8 MB                   | no              | none             | Apps needing glibc, libssl, ca‑certificates (e.g., Python, Node.js dynamically linked) |
| **Distroless Node.js**     | `gcr.io/distroless/nodejs18-debian12` | ~25 MB                  | no              | none             | Node.js apps (includes the Node runtime)                                               |
| **Scratch**                | `scratch`                             | 0 MB                    | none            | none             | Truly static binaries only (no DNS, no TLS unless compiled in)                         |

### Decision guideline

- **Go applications** → build as a static binary (`CGO_ENABLED=0`) and use `distroless/static`.
- **Node.js applications** → use `distroless/nodejs18-debian12` for runtime; `alpine` for build.
- **PHP applications** → `php:8.3-apache-alpine` balances size and functionality. Avoid the default Debian variant.
- **Static websites** → serve them via a tiny static file server (e.g., a 2‑line Go program) on `distroless/static`.

> **Why not always Alpine?** Alpine uses `musl libc` instead of `glibc`. This can cause subtle compatibility issues with pre‑compiled binaries or extensions (e.g., Python wheels with native code). Distroless provides a Debian/glibc environment _without_ a shell or package manager, giving you the best of both worlds.

---

## 2. Multi‑stage builds – isolate build dependencies from runtime

The build stage contains compilers, development headers, and all the heavy tooling. The final image only receives the compiled or bundled artifacts. This dramatically reduces size and removes sensitive build tools.

**Example – Go application**

```dockerfile
# Stage 1: build the binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /myapp .

# Stage 2: tiny runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /myapp /myapp
USER 10001:10001
ENTRYPOINT ["/myapp"]
```

The final image contains only the `myapp` binary and a few essential files. No shell, no compiler.

---

## 3. Run as a non‑root user with a fixed UID

Never run processes as `root`. Use the `USER` instruction. In this guide we standardise on UID/GID `10001`.

When using distroless images, you can simply specify `USER 10001` – the kernel only requires the numeric ID, a user name is not needed.

For images where you need to create a user (e.g., Alpine), do it explicitly:

```dockerfile
FROM alpine:3.19
RUN addgroup -g 10001 appgroup && \
    adduser -u 10001 -G appgroup -D appuser
USER appuser
```

> **Kubernetes enforcement** – Even if you forget `USER`, you can (and should) enforce non‑root in the pod spec:
>
> ```yaml
> securityContext:
>   runAsNonRoot: true
>   runAsUser: 10001
>   runAsGroup: 10001
> ```

---

## 4. Distroless – no shell, no package manager, no attack surface

Distroless images are production‑hardened. They contain **only** your application and its runtime dependencies. Debugging requires a separate sidecar or ephemeral container, which is actually a security advantage.

Because there is no `/bin/sh`, you cannot `docker exec` into a distroless container. If you absolutely need a shell for troubleshooting, use a debug image variant or attach an ephemeral debug container:

```bash
kubectl debug -it <pod> --image=busybox --target=<container>
```

---

## 5. Minimise layers and clean up within the same `RUN` command

Every `RUN`, `COPY`, and `ADD` creates a new layer. Combine commands and remove temporary files _in the same `RUN` instruction_ to avoid leaving cruft in intermediate layers.

**Bad** – leaves package cache in an untouchable layer:

```dockerfile
RUN apt-get update
RUN apt-get install -y curl
RUN rm -rf /var/lib/apt/lists/*
```

**Good** – single layer, no leftovers:

```dockerfile
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/*
```

For Alpine’s `apk`, the equivalent is `apk add --no-cache curl`.

---

## 6. Use a precise `.dockerignore`

Exclude files that are irrelevant to the build context. This prevents accidental leakage and speeds up the build.

```
node_modules
.git
.env
*.md
Dockerfile
.dockerignore
vendor
```

---

## 7. Pin image versions – never use `latest`

Tags like `latest` are mutable. Use a specific version and, ideally, the content‑addressable digest.

```dockerfile
# Avoid
FROM node:18

# Better
FROM node:18.17.1-alpine3.18

# Best (immutable)
FROM node@sha256:abc123def456...
```

---

## 8. Keep secrets out of image layers

Do **not** `COPY` or `ENV` credentials into an image. Use BuildKit’s secret mount for package registries:

```dockerfile
# syntax=docker/dockerfile:1
FROM node:18-alpine AS build
RUN --mount=type=secret,id=npmrc,target=/root/.npmrc \
    npm ci --only=production
```

Build with:

```bash
docker buildx build --secret id=npmrc,src=$HOME/.npmrc -t myapp .
```

---

## 9. Set file ownership at `COPY` time

When the runtime user is not `root`, use `--chown` to avoid additional `RUN chown` commands that bloat the layer.

```dockerfile
COPY --chown=10001:10001 --from=builder /app /app
```

---

## 10. Leverage BuildKit cache mounts for faster rebuilds

When you rebuild an image many times (e.g., during development or in CI/CD), you often re‑download the same dependencies over and over. Docker BuildKit’s `--mount=type=cache` lets you persist directories across builds, **dramatically speeding up subsequent builds** and reducing network usage.

> **Requirements:**
>
> - Docker Engine 18.09+
> - BuildKit enabled (set `DOCKER_BUILDKIT=1` or use `docker buildx build`)
> - Add `# syntax=docker/dockerfile:1` at the top of your Dockerfile

### How it works

You mount a cache volume into the build container at a known path (e.g., `/root/.npm`). The same cache volume is reused for every build of that Dockerfile. Unlike regular volumes, cache mounts do **not** persist in the final image – they only exist during the build.

### Common patterns

| Package manager | Typical cache path             | Mount flag example                              |
| --------------- | ------------------------------ | ----------------------------------------------- |
| npm / yarn      | `/root/.npm` or `/.yarn/cache` | `--mount=type=cache,target=/root/.npm`          |
| Go modules      | `/go/pkg/mod`                  | `--mount=type=cache,target=/go/pkg/mod`         |
| apt (Debian)    | `/var/cache/apt`               | `--mount=type=cache,target=/var/cache/apt`      |
| apk (Alpine)    | `/etc/apk/cache`               | `--mount=type=cache,target=/etc/apk/cache`      |
| composer (PHP)  | `/tmp/composer-cache`          | `--mount=type=cache,target=/tmp/composer-cache` |
| pip (Python)    | `/root/.cache/pip`             | `--mount=type=cache,target=/root/.cache/pip`    |

### Example – Node.js with npm cache

```dockerfile
# syntax=docker/dockerfile:1
FROM node:18-alpine AS builder
WORKDIR /app

# Cache npm packages globally
RUN --mount=type=cache,target=/root/.npm \
    npm install --cache /root/.npm

COPY . .
RUN npm run build
```

### Example – Go with module cache

```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /src

# Cache downloaded modules
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go build -o /app .
```

### Example – Debian/Ubuntu apt cache (multi‑stage)

```dockerfile
# syntax=docker/dockerfile:1
FROM debian:bookworm-slim AS builder
RUN --mount=type=cache,target=/var/cache/apt \
    apt-get update && \
    apt-get install -y --no-install-recommends curl ca-certificates
```

**Important:** Do **not** use `--mount=type=cache` for directories that contain secrets (e.g., `.npmrc` with auth tokens). Use `--mount=type=secret` for those instead (see Section 8).

With cache mounts, your CI pipelines will be faster, your developers will spend less time waiting, and you will still produce the same tiny, secure images – because cache mounts never end up in the final layer.

---

## Docker Image Example

### 🟢 Example 1 – Go HTTP Server (distroless + cached modules)

**Improvements:**

- BuildKit cache for Go modules
- static binary optimized for distroless/static
- no CGO
- read-only, no shell runtime

```dockerfile
# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache Go modules (FAST rebuilds)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go env -w CGO_ENABLED=0

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /server .

############################
# Runtime stage
############################
FROM gcr.io/distroless/static-debian12

COPY --from=builder /server /server

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["/server"]
```

---

### 🟡 Example 2 – Node.js Express API (distroless + npm cache + prune)

**Improvements:**

- BuildKit npm cache mount
- production-only dependencies
- no dev dependencies in runtime
- distroless Node runtime (no shell)

```dockerfile
# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM node:18.17.1-alpine3.18 AS builder

WORKDIR /app

COPY package*.json ./

# Fast installs with cache + no dev deps
RUN --mount=type=cache,target=/root/.npm \
    npm ci --omit=dev

COPY . .

# Optional: build step (if TypeScript etc.)
# RUN npm run build

############################
# Runtime stage
############################
FROM gcr.io/distroless/nodejs18-debian12

WORKDIR /app

COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/app.js ./app.js

USER 10001

EXPOSE 3000

CMD ["app.js"]
```

---

### 🟣 Example 3 – Laravel (Apache Alpine + optimized caching)

**Improvements:**

- Composer cache mount
- no extra chown layers (single COPY --chown)
- optimized Laravel cache layering
- minimal runtime footprint

```dockerfile
# syntax=docker/dockerfile:1

############################
# Dependencies stage
############################
FROM composer:2 AS vendor

WORKDIR /app

COPY composer.json composer.lock ./

RUN --mount=type=cache,target=/tmp/composer-cache \
    COMPOSER_CACHE_DIR=/tmp/composer-cache \
    composer install \
    --no-dev \
    --no-interaction \
    --no-scripts \
    --optimize-autoloader

COPY . .

############################
# Runtime stage
############################
FROM php:8.3-apache-alpine

# Install only required extensions (no extra packages)
RUN docker-php-ext-install pdo pdo_mysql

RUN a2enmod rewrite

ENV APACHE_LISTEN_PORT=8080 \
    APACHE_RUN_USER=#10001 \
    APACHE_RUN_GROUP=#10001

RUN sed -i 's/Listen 80/Listen 8080/' /etc/apache2/ports.conf

WORKDIR /var/www/html

# Single-layer ownership + copy (no extra chown RUN)
COPY --chown=10001:10001 . .
COPY --chown=10001:10001 --from=vendor /app/vendor ./vendor

# Laravel optimizations (single layer)
RUN php artisan config:cache && \
    php artisan route:cache && \
    php artisan view:cache && \
    chmod -R 775 storage bootstrap/cache

USER 10001

EXPOSE 8080
```

---

### 🔵 Example 4 – Static Website (ultra minimal + proper cache build)

**Improvements:**

- BuildKit npm cache mount
- no unnecessary image layers
- optimized Nginx config
- strict non-root runtime
- cleaned cache directories

```dockerfile
# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM node:18.17.1-alpine3.18 AS builder

WORKDIR /app

COPY package*.json ./

RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .

RUN npm run build

############################
# Runtime stage
############################
FROM nginx:stable-alpine

# Non-root user
RUN addgroup -g 10001 appgroup && \
    adduser -u 10001 -G appgroup -D appuser

# Minimal config (single layer)
RUN cat > /etc/nginx/conf.d/default.conf <<'EOF'
server {
    listen 8080;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF

# Required runtime dirs only
RUN mkdir -p /var/cache/nginx /var/log/nginx && \
    chown -R appuser:appgroup /var/cache/nginx /var/log/nginx

COPY --chown=appuser:appgroup --from=builder /app/dist /usr/share/nginx/html

USER appuser

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
```
