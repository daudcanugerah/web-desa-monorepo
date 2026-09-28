# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM node:22-alpine AS build
WORKDIR /app

# Install deps first for better layer caching
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Public API origin. The web client normalizes to end with /api/v1.
ARG VITE_API_BASE_URL=https://api.desapalasari.my.id/api/v1
ARG VITE_METABASE_URL=https://bi-embed.desapalasari.my.id
ARG VITE_FEATURE_GALLERY_PUBLIC=true
ARG VITE_FEATURE_BERITA_SEARCH=false
ARG VITE_FEATURE_INFOGRAPHIC_SORT=false

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_METABASE_URL=$VITE_METABASE_URL \
    VITE_FEATURE_GALLERY_PUBLIC=$VITE_FEATURE_GALLERY_PUBLIC \
    VITE_FEATURE_BERITA_SEARCH=$VITE_FEATURE_BERITA_SEARCH \
    VITE_FEATURE_INFOGRAPHIC_SORT=$VITE_FEATURE_INFOGRAPHIC_SORT

RUN npm run build

# ---- Runtime stage ----
FROM nginx:1.27-alpine AS runtime

# Build-time defaults, available to the runtime env hook. A container started
# with -e VITE_* overrides these without rebuilding.
ARG VITE_API_BASE_URL=https://api.desapalasari.my.id/api/v1
ARG VITE_METABASE_URL=https://bi-embed.desapalasari.my.id
ARG VITE_FEATURE_GALLERY_PUBLIC=true
ARG VITE_FEATURE_BERITA_SEARCH=false
ARG VITE_FEATURE_INFOGRAPHIC_SORT=false

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_METABASE_URL=$VITE_METABASE_URL \
    VITE_FEATURE_GALLERY_PUBLIC=$VITE_FEATURE_GALLERY_PUBLIC \
    VITE_FEATURE_BERITA_SEARCH=$VITE_FEATURE_BERITA_SEARCH \
    VITE_FEATURE_INFOGRAPHIC_SORT=$VITE_FEATURE_INFOGRAPHIC_SORT

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
# Runtime env hook: regenerates /usr/share/nginx/html/env.js at container start.
COPY docker-entrypoint.d/ /docker-entrypoint.d/
RUN chmod +x /docker-entrypoint.d/*.sh

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
